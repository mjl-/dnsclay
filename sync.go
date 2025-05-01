package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/libdns/libdns"

	"github.com/miekg/dns"

	"github.com/mjl-/bstore"
)

// We only want one sync for a zone at a time.
var zoneBusy = map[string]bool{}
var zoneCond = sync.NewCond(&sync.Mutex{})

// lockZone locks a sync for a sync, and returns an unlock function. Callers should
// call it with a defer, and may call it earlier as well.
func lockZone(zone string) (unlock func()) {
	zoneCond.L.Lock()
	for zoneBusy[zone] {
		zoneCond.Wait()
	}
	zoneBusy[zone] = true
	zoneCond.L.Unlock()
	var done bool
	return func() {
		if done {
			return
		}
		done = true
		zoneCond.L.Lock()
		zoneBusy[zone] = false
		zoneCond.L.Unlock()
		zoneCond.Broadcast()
	}
}

func serial(tm time.Time) Serial {
	return Serial(uint32(100 * (tm.Day() + 100*(int(tm.Month())+100*tm.Year()))))
}

// syncRecords syncs the records in the database with the latest records from the
// provider. If changed is true, the caller must queue a dns notify for the zone.
func syncRecords(log *slog.Logger, tx *bstore.Tx, z Zone, latest []libdns.Record) (changed bool, latestSOA *Record, inserted, deleted []Record, rerr error) {
	// note: for dns xfr, the first and last records are the SOA records. for SOA, we only keep the last.

	now := time.Now()

	defer func() {
		if rerr != nil {
			metricSyncErrors.Inc()
		}
	}()

	// We will be fetching the last known records from the database and group them into
	// rrsets. When a record changes, we delete + insert the whole rrset from/into the
	// database, so future IXFR will work properly.

	var knownSOA, prevSOA *Record

	type xlibdnsRecord struct {
		Key    recordKey
		Record libdns.Record
		Value  string
	}

	rrsetLatest := map[rrsetKey][]xlibdnsRecord{}
	rrsetKnown := map[rrsetKey][]Record{}

	q := bstore.QueryTx[Record](tx)
	q.FilterNonzero(Record{Zone: z.Name})
	q.FilterFn(func(r Record) bool { return r.Deleted == nil })
	known, err := q.List()
	if err != nil {
		return false, nil, nil, nil, fmt.Errorf("listing records in db: %v", err)
	}

	pc := ProviderConfig{Name: z.ProviderConfigName}
	if err := tx.Get(&pc); err != nil {
		return false, nil, nil, nil, fmt.Errorf("get provider config: %v", err)
	}

	for i, r := range known {
		log.Debug("known record", "record", r)
		if r.Type == Type(dns.TypeSOA) && r.AbsName == z.Name {
			knownSOA = &known[i]
			prevSOA = &known[i]
			continue
		}

		rk := r.rrsetKey()
		rrsetKnown[rk] = append(rrsetKnown[rk], r)
	}

	for _, lr := range latest {
		log.Debug("latest record", "record", lr)
		name := lr.Name
		if strings.HasSuffix(name, ".") {
			if !strings.EqualFold(name, z.Name) && (len(name) <= len(z.Name) || !strings.EqualFold(name[len(name)-len(z.Name)-1:], "."+z.Name)) {
				return false, nil, nil, nil, fmt.Errorf("received out of zone absolute name %q", name)
			}
		} else {
			name = libdns.AbsoluteName(name, z.Name)
		}
		text := fmt.Sprintf("%s %d %s %s", name, lr.TTL/time.Second, lr.Type, lr.Value)
		rr, err := dns.NewRR(text)
		if err != nil {
			return false, nil, nil, nil, fmt.Errorf("parsing record %q from remote: %v", text, err)
		}

		hex, value, err := recordData(rr)
		if err != nil {
			return false, nil, nil, nil, fmt.Errorf("getting data from record: %v", err)
		}

		h := rr.Header()
		name, err = cleanAbsName(h.Name)
		if err != nil {
			return false, nil, nil, nil, fmt.Errorf("clean name for %q: %v", h.Name, err)
		}
		// For AXFR, we will be getting two SOA records, at the start and end. The rfc2136
		// passes both of them on.
		if x, ok := rr.(*dns.SOA); ok && name == z.Name {
			latestSOA = &Record{0, z.Name, Serial(x.Serial), 0, now, nil, name, Type(h.Rrtype), Class(h.Class), TTL(h.Ttl), hex, value, lr.ID}
			continue
		}

		k := recordKey{h.Name, Type(h.Rrtype), Class(h.Class), TTL(h.Ttl), hex}
		xlr := xlibdnsRecord{k, lr, value}
		rk := rrsetKey{k.Name, k.Type, k.Class}
		rrsetLatest[rk] = append(rrsetLatest[rk], xlr)
	}

	if latestSOA == nil {
		return false, nil, nil, nil, fmt.Errorf("missing soa record")
	}

	// We may be using a different serial locally. Some name servers, like AWS Route53,
	// don't update serials when records are changed.
	newSerialRemote := latestSOA.SerialFirst

	if knownSOA != nil && latestSOA.SerialFirst <= 1 {
		latestSOA.SerialFirst = knownSOA.SerialFirst
	}

	// Ensure we have a SOA record with a new serial.
	var newSOA bool
	ensureSOA := func() error {
		if newSOA {
			return nil
		}
		// Ensure we have a new serial. We will make up one of the form YYYYMMDDNN, unless
		// the current serial is beyond that.
		if latestSOA.SerialFirst <= 1 {
			latestSOA.SerialFirst = serial(now)
		} else if knownSOA != nil && knownSOA.SerialFirst == latestSOA.SerialFirst {
			s := serial(now)
			if latestSOA.SerialFirst < serial(now.AddDate(0, 0, 1))-1 && latestSOA.SerialFirst < s {
				latestSOA.SerialFirst = s
			} else {
				latestSOA.SerialFirst++
			}
		}
		if knownSOA != nil {
			knownSOA.Deleted = &now
			knownSOA.SerialDeleted = latestSOA.SerialFirst
			if err := tx.Update(knownSOA); err != nil {
				return fmt.Errorf("marking known soa as deleted: %v", err)
			}
		}
		log.Debug("inserting new soa", "soa", latestSOA)
		if err := tx.Insert(latestSOA); err != nil {
			return fmt.Errorf("inserting new soa: %v", err)
		}
		knownSOA = latestSOA
		newSOA = true
		return nil
	}

	// If there is no SOA yet, create one.
	if knownSOA == nil {
		if err := ensureSOA(); err != nil {
			return false, nil, nil, nil, fmt.Errorf("ensuring first soa: %w", err)
		}
	}

	rrsetEqual := func(ll []xlibdnsRecord, kl []Record) bool {
		if len(ll) != len(kl) {
			return false
		}
		lkeys := map[recordKey]int{}
		kkeys := map[recordKey]int{}
		for _, xlr := range ll {
			lkeys[xlr.Key]++
		}
		for _, kr := range kl {
			kkeys[kr.recordKey()]++
		}
		for k, n := range lkeys {
			if n != kkeys[k] {
				return false
			}
		}
		return true
	}

	rrsetDel := func(kl []Record) error {
		if len(kl) > 0 && prevSOA == nil {
			return fmt.Errorf("cannot delete records without a previous soa (%v)", kl)
		}
		if err := ensureSOA(); err != nil {
			return err
		}
		for _, r := range kl {
			r.Deleted = &now
			r.SerialDeleted = latestSOA.SerialFirst
			if err := tx.Update(&r); err != nil {
				return fmt.Errorf("delete record: %v", err)
			}
			deleted = append(deleted, r)
		}
		return nil
	}

	rrsetAdd := func(ll []xlibdnsRecord) error {
		if err := ensureSOA(); err != nil {
			return err
		}
		for _, lr := range ll {
			k := lr.Key
			r := Record{0, z.Name, latestSOA.SerialFirst, 0, now, nil, k.Name, k.Type, k.Class, k.TTL, k.DataHex, lr.Value, lr.Record.ID}
			if err := tx.Insert(&r); err != nil {
				return fmt.Errorf("insert record: %v", err)
			}
			inserted = append(inserted, r)
		}
		return nil
	}

	// Replace/add with the updated/new rrsets.
	for rk, ll := range rrsetLatest {
		kl := rrsetKnown[rk]
		if rrsetEqual(ll, kl) {
			continue
		}
		if err := rrsetDel(kl); err != nil {
			return false, nil, nil, nil, err
		}
		if err := rrsetAdd(ll); err != nil {
			return false, nil, nil, nil, err
		}
	}

	// Remove old rrsets.
	for rk, kl := range rrsetKnown {
		if _, ok := rrsetLatest[rk]; !ok {
			if err := rrsetDel(kl); err != nil {
				return false, nil, nil, nil, err
			}
		}
	}

	log.Debug("insert/remove rrsets", "inserted", inserted, "deleted", deleted)

	nz := Zone{Name: z.Name}
	if err := tx.Get(&nz); err != nil {
		return false, nil, nil, nil, fmt.Errorf("get zone for update: %v", err)
	}
	nz.LastSync = &now
	nz.NextSync = now.Add(max(nz.SyncInterval, time.Minute))
	if newSOA {
		nz.LastRecordChange = &now
		nz.SerialLocal = Serial(latestSOA.SerialFirst)
		nz.SerialRemote = newSerialRemote
		if nz.RefreshInterval > 0 {
			ival := nz.RefreshInterval / (5 * 10)
			nz.NextRefresh = now.Add(max(ival, 5*time.Second))
		}
	}
	if err := tx.Update(&nz); err != nil {
		return false, nil, nil, nil, fmt.Errorf("update zone with time of last sync/change: %v", err)
	}

	return newSOA, latestSOA, inserted, deleted, nil
}

// ensurePropagate will do several attempts to ensure added/removed records are
// seen through the provider. Records are fetched from remote, records in the local
// database updated, and compared against the expected changes. Once all changes
// are seen, this function returns.
//
// Records in expDel may or may not be existing records (with ID nonzero). if their
// ID is nonzero, those exact records are checked for deletion.
//
// Must be called with zone lock held.
func ensurePropagate(ctx context.Context, log *slog.Logger, provider Provider, z Zone, expAdd []recordKey, expDel []Record, prevSerial Serial) (inserted, deleted []Record, rerr error) {
	var notify bool
	defer possiblyZoneNotify(log, z.Name, &notify)
	defer func() {
		if rerr != nil {
			metricPropagateErrors.Inc()
		}
	}()

	log.Debug("ensuring propagation", "zone", z.Name, "adds", expAdd, "deletes", expDel, "prevserial", prevSerial)

	var done bool

	checkDone := func(current []Record) {
		var il, dl []Record

		defer func() {
			log.Debug("checking if all records have been propagated", "done", done, "inserted", inserted, "deleted", deleted)
		}()

		mid := map[int64]Record{}
		mrk := map[recordKey]Record{}
		for _, r := range current {
			mid[r.ID] = r
			mrk[r.recordKey()] = r
		}

		for _, a := range expAdd {
			if r, ok := mrk[a]; !ok {
				log.Debug("record not yet added/updated", "record", a, "exists", ok, "serial", r.SerialFirst)
				return
			} else {
				il = append(il, r)
			}
		}
		for _, d := range expDel {
			if d.ID > 0 {
				if r, ok := mid[d.ID]; ok {
					log.Debug("record not yet deleted", "record", d)
					return
				} else {
					dl = append(dl, r)
				}
			} else {
				if r, ok := mrk[d.recordKey()]; ok {
					log.Debug("record not yet deleted", "record", d)
					return
				} else {
					dl = append(dl, r)
				}
			}
		}
		inserted = il
		deleted = dl
		done = true
	}

	sync := func() error {
		latest, err := getRecords(ctx, log, provider, z.Name, false)
		if err != nil {
			return fmt.Errorf("getting latest records through provider: %v", err)
		}

		err = database.Write(ctx, func(tx *bstore.Tx) error {
			var ch bool
			ch, _, _, _, err = syncRecords(log, tx, z, latest)
			if err != nil {
				return fmt.Errorf("updating records from latest: %w", err)
			}

			notify = notify || ch

			q := bstore.QueryTx[Record](tx)
			q.FilterNonzero(Record{Zone: z.Name})
			q.FilterFn(func(r Record) bool { return r.Deleted == nil })
			current, err := q.List()
			if err != nil {
				return fmt.Errorf("listing records after processing updates: %v", err)
			}

			checkDone(current)
			return nil
		})
		if err != nil {
			return fmt.Errorf("checking for propagation: %w", err)
		}
		return nil
	}

	waits := []time.Duration{propagationFirstWait, time.Second, 2 * time.Second, 3 * time.Second}
	for _, w := range waits {
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-time.After(w):
		}

		err := sync()
		if err != nil {
			return nil, nil, err
		} else if done {
			break
		}

		log.Debug("waiting to check for propagation again", "wait", w)
	}
	if !done {
		return nil, nil, fmt.Errorf("not all changes found")
	}
	return
}

// possiblyZoneNotify is a convenience function for use with "defer", to send DNS
// notifications for a zone.
func possiblyZoneNotify(log *slog.Logger, zone string, notify *bool) {
	if !*notify {
		return
	}
	go func() {
		defer recoverPanic(log, "notifying zones")
		sendZoneNotify(log, zone)
	}()
}

// Best-effort sending of dns notify for zone, typically called in goroutine.
func sendZoneNotify(log *slog.Logger, zone string) {
	ctx := shutdownCtx

	// Reschedule next zone refresh/sync.
	refreshKick()

	var z Zone
	var znl []ZoneNotify
	var soa dns.SOA
	err := database.Read(ctx, func(tx *bstore.Tx) error {
		z = Zone{Name: zone}
		err := tx.Get(&z)
		if err != nil {
			return fmt.Errorf("get zone: %w", err)
		}
		q := bstore.QueryTx[ZoneNotify](tx)
		q.FilterNonzero(ZoneNotify{Zone: z.Name})
		znl, err = q.List()
		if err != nil {
			return fmt.Errorf("listing zone notify destinations: %w", err)
		}

		qs := bstore.QueryTx[Record](tx)
		qs.FilterNonzero(Record{Type: Type(dns.TypeSOA), Zone: zone})
		qs.FilterFn(func(r Record) bool { return r.Deleted == nil })
		r, err := qs.Get()
		if err != nil {
			return fmt.Errorf("get soa for zone for notify: %w", err)
		}

		soarr, err := r.RR()
		if err != nil {
			return fmt.Errorf("rr for soa db record: %v", err)
		}
		soa = *soarr.(*dns.SOA)

		return nil
	})
	if err != nil {
		log.Error("gathering notify destinations", "err", err)
		return
	}

	log.Debug("preparing to send dns notify", "ndestinations", len(znl))
	for i := range znl {
		zn := znl[i]
		go func() {
			defer recoverPanic(log, "sending dns notify")

			err := dnsNotify(log, zn, soa)
			if err != nil {
				log.Info("sending dns notify", "err", err, "zonenotify", zn)
			}
		}()
	}
}

// dnsNotify sends a a single DNS notification to a server address.
func dnsNotify(log *slog.Logger, zn ZoneNotify, soa dns.SOA) error {
	log = log.With("zone", zn.Zone, "proto", zn.Protocol, "addr", zn.Address)

	var om dns.Msg
	om.SetNotify(zn.Zone)
	om.Answer = []dns.RR{&soa}
	client := dns.Client{}
	switch zn.Protocol {
	case "tcp":
		client.Net = "tcp"
	case "udp":
	default:
		return fmt.Errorf("unknown protocol %q", zn.Protocol)
	}
	log.Debug("outgoing dns notify request", "outmsg", om)
	im, _, err := client.Exchange(&om, zn.Address)
	log.Debug("dns notify transaction", "err", err, "inmsg", im)
	if err == nil {
		err = responseError(im)
	}
	if err != nil {
		return fmt.Errorf("dns notify transaction: %w", err)
	}
	return nil
}
