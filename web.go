package main

import (
	"context"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/libdns/libdns"

	"github.com/miekg/dns"

	"github.com/mjl-/bstore"
	"github.com/mjl-/sherpa"
	"github.com/mjl-/sherpadoc"
)

// API is the webapi used by the admin frontend.
type API struct{}

var apiDoc sherpadoc.Section

func httpBasicAuth(fn func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "admin" || password != adminpassword {
			w.Header().Set("WWW-Authenticate", `Basic realm="dnsclay"`)
			http.Error(w, "401 - unauthorized", http.StatusUnauthorized)
			return
		}
		fn(w, r)
	}
}

// write a copy of the database from within a readonly transaction, for a consistent view.
func exportDatabase(w http.ResponseWriter, r *http.Request) {
	log := cidlog(r.Context())
	h := w.Header()
	h.Set("Content-Type", "application/octet-stream")
	h.Set("Cache-Control", "no-cache, max-age=0")
	err := database.Read(r.Context(), func(tx *bstore.Tx) error {
		_, err := tx.WriteTo(w)
		return err
	})
	if err != nil && !isClosed(err) {
		log.Error("exporting database", "err", err)
	}
}

// NOTE: Functions starting with an underscore can panic with a *sherpa.Error. They
// are are recognized by the sherpa handler and turned into regular error
// conditions.

func _zone(tx *bstore.Tx, zone string) (z Zone) {
	z = Zone{Name: zone}
	err := tx.Get(&z)
	_checkf(err, "get zone")
	return
}

func _record(tx *bstore.Tx, zoneName string, id int64) (r Record) {
	r = Record{ID: id}
	err := tx.Get(&r)
	if err == nil && r.Zone != zoneName {
		err = bstore.ErrAbsent
	}
	_checkf(err, "get record")
	return
}

// Zones returns all zones.
func (x API) Zones(ctx context.Context) []Zone {
	zones, err := bstore.QueryDB[Zone](ctx, database).List()
	_checkf(err, "listing zones")
	return zones
}

// Zone returns details about a single zone, the provider config, dns notify
// destinations, credentials with access to the zone, and record sets. The returned
// record sets includes those no long active (i.e. deleted). The
// history/propagation state fo the record sets only includes those that may still
// be in caches. Use ZoneRecordSetHistory for the full history for a single record
// set.
func (x API) Zone(ctx context.Context, zone string) (z Zone, pc ProviderConfig, notifies []ZoneNotify, credentials []Credential, sets []RecordSet) {
	var records []Record
	_dbread(ctx, func(tx *bstore.Tx) {
		z = _zone(tx, zone)

		pc = ProviderConfig{Name: z.ProviderConfigName}
		err := tx.Get(&pc)
		_checkf(err, "get provider config")

		notifies, err = bstore.QueryTx[ZoneNotify](tx).FilterNonzero(ZoneNotify{Zone: zone}).List()
		_checkf(err, "listing notify addresses")

		err = bstore.QueryTx[ZoneCredential](tx).FilterNonzero(ZoneCredential{Zone: zone}).ForEach(func(zc ZoneCredential) error {
			c := Credential{ID: zc.ID}
			err := tx.Get(&c)
			_checkf(err, "get credential for zone")
			credentials = append(credentials, c)
			return nil
		})
		_checkf(err, "listing zone credentials")

		records, err = bstore.QueryTx[Record](tx).FilterNonzero(Record{Zone: zone}).List()
		_checkf(err, "list records")
	})

	sets = _propagationStates(records)

	return
}

// ZoneRecords returns all records for a zone, including historic records, without
// grouping them into record sets.
func (x API) ZoneRecords(ctx context.Context, zone string) (records []Record) {
	_dbread(ctx, func(tx *bstore.Tx) {
		z := _zone(tx, zone)

		var err error
		records, err = bstore.QueryTx[Record](tx).FilterNonzero(Record{Zone: z.Name}).List()
		_checkf(err, "list records")
	})
	return
}

// ZoneRefresh starts a sync of the records from the provider into the local
// database, sending dns notify if needed. ZoneRefresh returns all records
// (included deleted) from after the synchronization.
func (x API) ZoneRefresh(ctx context.Context, zone string) (z Zone, sets []RecordSet) {
	log := cidlog(ctx)

	var provider Provider
	_dbread(ctx, func(tx *bstore.Tx) {
		var err error
		z, provider, err = zoneProvider(tx, zone)
		_checkf(err, "get zone and provider")
	})

	unlock := lockZone(z.Name)
	defer unlock()

	var cancel func()
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	latest, err := getRecords(ctx, log, provider, zone)
	_checkf(err, "getting latest records through provider")

	var notify bool
	defer possiblyZoneNotify(log, zone, &notify)

	_dbwrite(ctx, func(tx *bstore.Tx) {
		z = _zone(tx, zone) // Again.

		notify, _, _, _, err = syncRecords(log, tx, z, latest)
		_checkf(err, "storing latest records in database")

		records, err := bstore.QueryTx[Record](tx).FilterNonzero(Record{Zone: zone}).List()
		_checkf(err, "list records")
		sets = _propagationStates(records)
	})

	return
}

// ZonePurgeHistory removes historic records from the database, those marked "deleted".
func (x API) ZonePurgeHistory(ctx context.Context, zone string) (z Zone, sets []RecordSet) {
	_dbwrite(ctx, func(tx *bstore.Tx) {
		z = _zone(tx, zone) // Again.

		q := bstore.QueryTx[Record](tx)
		q.FilterNonzero(Record{Zone: z.Name})
		q.FilterFn(func(r Record) bool { return r.Deleted != nil })
		_, err := q.Delete()
		_checkf(err, "removing record history")

		records, err := bstore.QueryTx[Record](tx).FilterNonzero(Record{Zone: z.Name}).List()
		_checkf(err, "listing records ")
		sets = _propagationStates(records)
	})

	return
}

// ZoneAdd adds a new zone to the database. A TSIG credential is created
// automatically. Records are fetched returning the new zone, in the background.
//
// If pc.ProviderName is non-empty, a new ProviderConfig is added.
func (x API) ZoneAdd(ctx context.Context, z Zone, notifies []ZoneNotify) (nzone Zone) {
	log := cidlog(ctx)
	var provider Provider

	_dbwrite(ctx, func(tx *bstore.Tx) {
		now := time.Now()

		z.Name = _cleanAbsName(strings.TrimSuffix(z.Name, ".") + ".")
		z.SyncInterval = 24 * time.Hour
		z.RefreshInterval = time.Hour
		z.NextSync = now.Add(z.SyncInterval)
		z.NextRefresh = now.Add(z.RefreshInterval / (5 * 10))
		err := tx.Insert(&z)
		_checkf(err, "adding zone")

		_, provider, err = zoneProvider(tx, z.Name)
		_checkf(err, "get zone and provider")

		for _, n := range notifies {
			n.Zone = z.Name
			switch n.Protocol {
			case "tcp", "udp":
			default:
				_checkuserf(fmt.Errorf("unknown protocol %q", n.Protocol), "checking notify")
			}
			_, _, err := net.SplitHostPort(n.Address)
			_checkuserf(err, "checking notify address")
			err = tx.Insert(&n)
			_checkf(err, "inserting notify")
		}

		tsigbuf := make([]byte, 32)
		_, err = io.ReadFull(cryptorand.Reader, tsigbuf)
		_checkf(err, "read random")
		cred := Credential{
			Name:       "zone-default-tsig-" + strings.TrimSuffix(z.Name, "."),
			Type:       "tsig",
			TSIGSecret: base64.StdEncoding.EncodeToString(tsigbuf),
		}
		err = tx.Insert(&cred)
		_checkf(err, "inserting tsig credential")
		zonecred := ZoneCredential{
			Zone:         z.Name,
			CredentialID: cred.ID,
		}
		err = tx.Insert(&zonecred)
		_checkf(err, "inserting tsig zone credential")
	})

	go func() {
		defer recoverPanic(log, "fetching records for new zone")

		unlock := lockZone(z.Name)
		defer unlock()

		var cancel func()
		ctx, cancel = context.WithTimeout(shutdownCtx, 30*time.Second)
		defer cancel()
		latest, err := getRecords(ctx, log, provider, z.Name)
		if err != nil {
			log.Debug("getting latest records through provider", "err", err)
			return
		}

		var changed bool
		defer possiblyZoneNotify(log, z.Name, &changed)

		err = database.Write(ctx, func(tx *bstore.Tx) error {
			_zone(tx, z.Name) // Again.

			changed, _, _, _, err = syncRecords(log, tx, z, latest)
			return err
		})
		if err != nil {
			log.Debug("updating records in database", "err", err)
		}
	}()

	return z
}

// ZoneDelete removes a zone and all its records, credentials and dns notify addresses, from the database.
func (x API) ZoneDelete(ctx context.Context, zone string) {
	_dbwrite(ctx, func(tx *bstore.Tx) {
		z := _zone(tx, zone)

		_, err := bstore.QueryTx[ZoneNotify](tx).FilterNonzero(ZoneNotify{Zone: z.Name}).Delete()
		_checkf(err, "deleting notify addresses for zone")

		zonecreds, err := bstore.QueryTx[ZoneCredential](tx).FilterNonzero(ZoneCredential{Zone: z.Name}).List()
		_checkf(err, "listing zone credentials")
		for _, zc := range zonecreds {
			err = tx.Delete(&zc)
			_checkf(err, "deleting zone credential")
			err := tx.Delete(&Credential{ID: zc.CredentialID})
			_checkf(err, "deleting credential")
		}

		_, err = bstore.QueryTx[Record](tx).FilterNonzero(Record{Zone: z.Name}).Delete()
		_checkf(err, "deleting records for zone")

		err = tx.Delete(&z)
		_checkf(err, "deleting zone")

		exists, err := bstore.QueryTx[Zone](tx).FilterNonzero(Zone{ProviderConfigName: z.ProviderConfigName}).Exists()
		_checkf(err, "checking if references to provider config still exists")
		if !exists {
			pc := ProviderConfig{Name: z.ProviderConfigName}
			err := tx.Delete(&pc)
			_checkf(err, "deleting provider config")
		}
	})
}

// ZoneUpdate updates the provider config and refresh & sync interval for a zone.
func (x API) ZoneUpdate(ctx context.Context, z Zone) (nz Zone) {
	_dbwrite(ctx, func(tx *bstore.Tx) {
		oz := _zone(tx, z.Name)

		oz.ProviderConfigName = z.ProviderConfigName
		oz.RefreshInterval = z.RefreshInterval
		oz.SyncInterval = z.SyncInterval
		err := tx.Update(&oz)
		_checkf(err, "update zone")
		nz = oz
	})
	return
}

// ZoneNotify send a DNS notify message to an address.
func (x API) ZoneNotify(ctx context.Context, zoneNotifyID int64) {
	log := cidlog(ctx)

	zn := ZoneNotify{ID: zoneNotifyID}
	var soa dns.SOA
	_dbread(ctx, func(tx *bstore.Tx) {
		err := tx.Get(&zn)
		_checkf(err, "get zone notify details")

		q := bstore.QueryTx[Record](tx)
		q.FilterNonzero(Record{Type: Type(dns.TypeSOA), Zone: zn.Zone})
		q.FilterFn(func(r Record) bool { return r.Deleted == nil })
		r, err := q.Get()
		_checkf(err, "get soa from db")

		soarr, err := r.RR()
		_checkf(err, "get rr for db soa record")
		soa = *soarr.(*dns.SOA)
	})

	err := dnsNotify(log, zn, soa)
	_checkf(err, "notifying")
}

// ZoneNotifyAdd adds a new DNS NOTIFY destination to a zone.
func (x API) ZoneNotifyAdd(ctx context.Context, zn ZoneNotify) (nzn ZoneNotify) {
	_dbwrite(ctx, func(tx *bstore.Tx) {
		zn.Created = time.Time{}
		err := tx.Insert(&zn)
		_checkf(err, "inserting zone notify")
		nzn = zn
	})
	return
}

// ZoneNotifyDelete removes a DNS NOTIFY destination from a zone.
func (x API) ZoneNotifyDelete(ctx context.Context, zoneNotifyID int64) {
	_dbwrite(ctx, func(tx *bstore.Tx) {
		zn := ZoneNotify{ID: zoneNotifyID}
		err := tx.Delete(&zn)
		_checkf(err, "deleting zone notify")
	})
}

// ZoneCredentialAdd adds a new TSIG or TLS public key credential to a zone.
func (x API) ZoneCredentialAdd(ctx context.Context, zone string, c Credential) (nc Credential) {
	_dbwrite(ctx, func(tx *bstore.Tx) {
		_zone(tx, zone)

		// Name must be valid for use in DNS, we store it without trailing dot.
		name := _cleanAbsName(strings.TrimSuffix(c.Name, ".") + ".")
		c.Name = strings.TrimSuffix(name, ".")

		c.Created = time.Time{}
		switch c.Type {
		case "tsig":
			if c.TSIGSecret == "" {
				randbuf := make([]byte, 32)
				_, err := io.ReadFull(cryptorand.Reader, randbuf)
				_checkf(err, "reading random bytes")
				c.TSIGSecret = base64.StdEncoding.EncodeToString(randbuf)
			} else {
				_, err := base64.StdEncoding.DecodeString(c.TSIGSecret)
				_checkuserf(err, "parsing tsig secret %q", c.TSIGSecret)
			}
			c.TLSPublicKey = ""

		case "tlspubkey":
			if c.TLSPublicKey == "" {
				_checkuserf(errors.New("must not be empty"), "checking tls public key")
			}
			buf, err := base64.RawURLEncoding.DecodeString(c.TLSPublicKey)
			if len(buf) != sha256.Size {
				err = fmt.Errorf("got %d bytes, need %d", len(buf), sha256.Size)
			}
			_checkuserf(err, "parsing tls public key")
			c.TSIGSecret = ""

			q := bstore.QueryTx[Credential](tx)
			q.FilterNonzero(Credential{TLSPublicKey: c.TLSPublicKey, Type: "tlspubkey"})
			ok, err := q.Exists()
			if err == nil && ok {
				err = errors.New("public key already present")
			}
			_checkf(err, "checking tlspubkey")

		default:
			_checkuserf(fmt.Errorf("unknown value %q", c.Type), "checking type")
		}

		err := tx.Insert(&c)
		_checkf(err, "inserting credential")

		zc := ZoneCredential{0, zone, c.ID}
		err = tx.Insert(&zc)
		_checkf(err, "inserting zone credential")

		nc = c
	})
	return
}

// ZoneCredentialDelete removes a TSIG/TLS public key credential from a zone.
func (x API) ZoneCredentialDelete(ctx context.Context, credentialID int64) {
	_dbwrite(ctx, func(tx *bstore.Tx) {
		c := Credential{ID: credentialID}
		err := tx.Get(&c)
		_checkf(err, "get credential")

		n, err := bstore.QueryTx[ZoneCredential](tx).FilterNonzero(ZoneCredential{CredentialID: c.ID}).Delete()
		if err == nil && n != 1 {
			err = fmt.Errorf("deleted %d records, expected 1", n)
		}
		_checkf(err, "deleting zone credential")

		err = tx.Delete(&c)
		_checkf(err, "delete credential")
	})
}

// ZoneImportRecords parses records in zonefile, assuming standard zone file syntax,
// and adds the records via the provider and syncs the newly added records to the
// local database. The latest records, included historic/deleted records after the
// sync are returned.
func (x API) ZoneImportRecords(ctx context.Context, zone, zonefile string) []Record {
	log := cidlog(ctx)

	var z Zone
	var provider Provider
	_dbread(ctx, func(tx *bstore.Tx) {
		var err error
		z, provider, err = zoneProvider(tx, zone)
		_checkf(err, "get zone and provider")
	})

	zp := dns.NewZoneParser(strings.NewReader(zonefile), z.Name, "")
	zp.SetDefaultTTL(300)
	var l []Record
	for rr, ok := zp.Next(); ok; rr, ok = zp.Next() {
		hex, value, err := recordData(rr)
		_checkf(err, "parsing record value")

		h := rr.Header()
		if h.Class != dns.ClassINET {
			_checkuserf(errors.New("only class INET supported"), "checking record")
		}
		h.Name = _cleanAbsName(h.Name)

		l = append(l, Record{0, z.Name, 0, 0, time.Time{}, nil, h.Name, Type(h.Rrtype), dns.ClassINET, TTL(h.Ttl), hex, value, ""})
	}
	err := zp.Err()
	if err == nil && len(l) == 0 {
		err = errors.New("no records found")
	}
	_checkuserf(err, "parsing zone file")

	unlock := lockZone(z.Name)
	defer unlock()

	// Get latest.
	latest, err := getRecords(ctx, log, provider, z.Name)
	_checkf(err, "get latest records")

	var notify bool
	defer possiblyZoneNotify(log, z.Name, &notify)

	var soa Record
	_dbwrite(ctx, func(tx *bstore.Tx) {
		notify, _, _, _, err = syncRecords(log, tx, z, latest)
		_checkf(err, "updating records from latest before adding")

		soa = zoneSOA(log, tx, z.Name)
	})

	var cancel func()
	ctx, cancel = context.WithTimeout(ctx, time.Minute)
	defer cancel()

	ladded, err := appendRecords(ctx, log, provider, z.Name, libdnsRecords(l))
	if err == nil && len(ladded) < len(l) {
		err = fmt.Errorf("provider added %d records, expected %d (%v != %v)", len(ladded), len(l), ladded, l)
	}
	_checkf(err, "adding records via provider")
	log.Debug("added record through provider", "records", l, "ladded", ladded)

	var rkl []recordKey
	for _, r := range l {
		rkl = append(rkl, r.recordKey())
	}
	inserted, _, err := ensurePropagate(ctx, log, provider, z, rkl, nil, soa.SerialFirst)
	_checkf(err, "ensuring record propagation")
	return inserted
}

// RecordNew is a new or updated record.
type RecordNew struct {
	RelName string
	TTL     TTL
	Type    Type
	Value   string
}

func _parseNew(zone string, nr RecordNew) Record {
	absname := nr.RelName
	if absname != "" {
		absname += "."
	}
	absname += zone
	absname = _cleanAbsName(absname)

	typ := dns.TypeToString[uint16(nr.Type)]
	if typ == "" {
		_checkuserf(fmt.Errorf("unknown type %d", nr.Type), "checking record type")
	}
	if nr.TTL == 0 {
		_checkuserf(errors.New("ttl must be > 0"), "checking ttl")
	}

	text := fmt.Sprintf("%s %d %s %s", absname, nr.TTL, typ, nr.Value)
	rr, err := dns.NewRR(text)
	_checkuserf(err, "parsing new record")

	hex, value, err := recordData(rr)
	_checkf(err, "parsing record value")

	return Record{0, zone, 0, 0, time.Time{}, nil, absname, nr.Type, dns.ClassINET, nr.TTL, hex, value, ""}
}

// RecordAdd adds a single record through the provider, then waits for it to
// synchronize back to the local database. The newly added database record is
// returned.
func (x API) RecordAdd(ctx context.Context, zone string, nr RecordNew) (r Record) {
	log := cidlog(ctx)

	var z Zone
	var provider Provider
	var soa Record
	_dbread(ctx, func(tx *bstore.Tx) {
		var err error
		z, provider, err = zoneProvider(tx, zone)
		_checkf(err, "get zone and provider")
	})

	xnr := _parseNew(zone, nr)

	unlock := lockZone(z.Name)
	defer unlock()

	// Get latest.
	latest, err := getRecords(ctx, log, provider, zone)
	_checkf(err, "get latest records")

	var notify bool
	defer possiblyZoneNotify(log, zone, &notify)

	_dbwrite(ctx, func(tx *bstore.Tx) {
		notify, _, _, _, err = syncRecords(log, tx, z, latest)
		_checkf(err, "updating records from latest before looking record to delete")

		soa = zoneSOA(log, tx, z.Name)
	})

	var cancel func()
	ctx, cancel = context.WithTimeout(ctx, time.Minute)
	defer cancel()

	ladded, err := appendRecords(ctx, log, provider, z.Name, []libdns.Record{xnr.libdnsRecord()})
	if err == nil && len(ladded) != 1 {
		err = fmt.Errorf("provider added %d records, expected 1", len(ladded))
	}
	_checkf(err, "adding records via provider")
	log.Debug("added record through provider", "record", xnr, "ladded", ladded)

	inserted, _, err := ensurePropagate(ctx, log, provider, z, []recordKey{xnr.recordKey()}, nil, soa.SerialFirst)
	_checkf(err, "ensuring record propagation")
	return inserted[0]
}

// RecordUpdate updates and existing record, replacing it with the new version. The
// name and type must be the same. Only the TTL and value can be changed for an
// existing record. The updated or new local database record is returned after a sync.
func (x API) RecordUpdate(ctx context.Context, zone string, recordID int64, rn RecordNew) (xr Record) {
	log := cidlog(ctx)

	var z Zone
	var provider Provider
	_dbread(ctx, func(tx *bstore.Tx) {
		var err error
		z, provider, err = zoneProvider(tx, zone)
		_checkf(err, "get zone and provider")
	})

	nr := _parseNew(z.Name, rn)

	unlock := lockZone(z.Name)
	defer unlock()

	// Get latest.
	latest, err := getRecords(ctx, log, provider, zone)
	_checkf(err, "get latest records")

	var notify bool
	defer possiblyZoneNotify(log, zone, &notify)

	var or Record
	var orrset []Record
	_dbwrite(ctx, func(tx *bstore.Tx) {
		notify, _, _, _, err = syncRecords(log, tx, z, latest)
		_checkf(err, "updating records from latest before looking record to delete")

		or = _record(tx, z.Name, recordID)
		if or.Deleted != nil {
			_checkuserf(errors.New("record is marked deleted"), "checking current record")
		}

		if nr.AbsName != or.AbsName {
			_checkuserf(errors.New("cannot change name"), "checking current record")
		}
		if nr.Type != or.Type {
			_checkuserf(errors.New("cannot change type"), "checking current record")
		}

		var err error
		q := bstore.QueryDB[Record](ctx, database)
		q.FilterNonzero(Record{Zone: zone, AbsName: or.AbsName})
		q.FilterFn(func(r Record) bool { return r.Deleted == nil })
		orrset, err = q.List()
		_checkf(err, "get old rrset")
	})

	var cancel func()
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	nr.ProviderID = or.ProviderID
	nrrset := []Record{nr}
	for _, r := range orrset {
		if r.ID != or.ID {
			nrrset = append(nrrset, r)
		}
	}

	// We're only going to wait for update propagation when the record actually
	// changed. Otherwise, the provider may not be making any changes, and we won't get
	// a new serial, and no updated records when we sync.
	var adds []recordKey
	var dels []Record
	if nr.recordKey() != or.recordKey() {
		adds = []recordKey{nr.recordKey()}
		dels = []Record{or}
	}

	// If we don't have explicit IDs for records tracked by the provider, we don't know
	// what "setting" a record means. If that's the case, we delete & add the record.
	if nrrset[0].ProviderID != "" {
		lupdated, err := setRecords(ctx, log, provider, z.Name, libdnsRecords(nrrset))
		if err == nil && len(lupdated) < len(adds) {
			slog.Info("provider reported updates", "updated", lupdated, "oldrecord", or, "newrecord", nr, "nrrset", nrrset)
			err = fmt.Errorf("provider reports %d records were updated, expected at least %d", len(lupdated), len(adds))
		}
		_checkf(err, "updating record through provider")
		log.Debug("records updated through provider", "nrrset", nrrset, "lupdated", lupdated, "adds", adds, "dels", dels)

		if nr.recordKey() != or.recordKey() {
			var oldUpdated bool
			for _, r := range lupdated {
				if r.ID != "" && r.ID == or.ProviderID {
					oldUpdated = true
					break
				}
			}
			if !oldUpdated {
				_, err := deleteRecords(ctx, log, provider, z.Name, []libdns.Record{or.libdnsRecord()})
				_checkf(err, "removing old record")
			}
		}
	} else {
		ldeleted, err := deleteRecords(ctx, log, provider, z.Name, libdnsRecords(orrset))
		if err == nil && len(ldeleted) < len(adds) {
			slog.Info("provider reported deletes", "deleted", ldeleted, "oldrecord", or, "newrecord", nr, "orrset", orrset)
			err = fmt.Errorf("provider reports %d records were updated, expected at least %d", len(ldeleted), len(adds))
		}
		_checkf(err, "deleting old record(s) through provider")
		log.Debug("records deleted through provider", "orrset", orrset, "ldeted", ldeleted, "adds", adds, "dels", dels)

		ladd, err := appendRecords(ctx, log, provider, z.Name, libdnsRecords(nrrset))
		if err == nil && len(ladd) < len(adds) {
			slog.Info("provider reported appends", "add", ladd, "oldrecord", or, "newrecord", nr, "nrrset", nrrset)
			err = fmt.Errorf("provider reports %d records were updated, expected at least %d", len(ladd), len(adds))
		}
		_checkf(err, "adding old record(s) through provider")
		log.Debug("records added through provider", "nrrset", nrrset, "ladd", ladd, "adds", adds, "dels", dels)
	}

	inserted, _, err := ensurePropagate(ctx, log, provider, z, adds, dels, or.SerialFirst)
	_checkf(err, "ensuring propagation")
	for _, ins := range inserted {
		if ins.ID == or.ID || ins.recordKey() == nr.recordKey() {
			return ins
		}
	}
	return or
}

// RecordDelete removes a record through the provider and waits for the change to
// be synced to the local database. The historic/deleted record is returned.
func (x API) RecordDelete(ctx context.Context, zone string, recordID int64) (r Record) {
	log := cidlog(ctx)

	var z Zone
	var provider Provider
	var or Record
	var soa Record
	_dbread(ctx, func(tx *bstore.Tx) {
		var err error
		z, provider, err = zoneProvider(tx, zone)
		_checkf(err, "get zone and provider")
	})

	unlock := lockZone(z.Name)
	defer unlock()

	// Get latest.
	latest, err := getRecords(ctx, log, provider, zone)
	_checkf(err, "get latest records")

	var notify bool
	defer possiblyZoneNotify(log, zone, &notify)

	// Sync and get record to delete.
	_dbwrite(ctx, func(tx *bstore.Tx) {
		notify, _, _, _, err = syncRecords(log, tx, z, latest)
		_checkf(err, "updating records from latest before looking record to delete")

		or = _record(tx, z.Name, recordID)
		if or.Deleted != nil {
			_checkuserf(errors.New("record is already marked deleted"), "checking current record")
		}

		soa = zoneSOA(log, tx, z.Name)
	})

	var cancel func()
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	removed, err := deleteRecords(ctx, log, provider, z.Name, []libdns.Record{or.libdnsRecord()})
	if err == nil && len(removed) != 1 {
		err = fmt.Errorf("provider reports %d records were removed, expected 1", len(removed))
	}
	_checkf(err, "deleting record through provider")
	log.Debug("records removed", "records", removed)

	_, dels, err := ensurePropagate(ctx, log, provider, z, nil, []Record{or}, soa.SerialFirst)
	_checkf(err, "ensuring propagation")
	return dels[0]
}

// Version returns the version of this build of the application.
func (x API) Version(ctx context.Context) string {
	return version
}

// DNSTypeNames returns a mapping of DNS type numbers to strings.
func (x API) DNSTypeNames(ctx context.Context) map[uint16]string {
	return dns.TypeToString
}

// KnownProviders is a dummy method whose sole purpose is to get an API description
// of all known providers in the API documentation, for use in TypeScript.
func (x API) KnownProviders(ctx context.Context) (KnownProviders, sherpadoc.Section) {
	return KnownProviders{}, sherpadoc.Section{}
}

// Docs returns the API docs. The TypeScript code uses this documentation to build
// a UI for the fields in configurations for providers (as included through
// KnownProviders).
func (x API) Docs(ctx context.Context) sherpadoc.Section {
	return apiDoc
}

// ProviderConfigTest tests the provider configuration for zone. Used before
// creating a zone with a new config or updating the config for an existing zone.
func (x API) ProviderConfigTest(ctx context.Context, zone string, provider string, providerConfigJSON string) (nrecords int) {
	log := cidlog(ctx)

	zone = strings.TrimSuffix(zone, ".") + "."
	zone = _cleanAbsName(zone)

	p, err := providerForConfig(provider, providerConfigJSON)
	if err != nil && errors.Is(err, errProviderUserError) {
		_checkuserf(err, "checking provider")
	}
	_checkf(err, "checking provider")

	records, err := getRecords(ctx, log, p, zone)
	_checkuserf(err, "fetching records for testing provider config")

	return len(records)
}

// ProviderConfigs returns all provider configs.
func (x API) ProviderConfigs(ctx context.Context) (providerConfigs []ProviderConfig) {
	_dbread(ctx, func(tx *bstore.Tx) {
		var err error
		providerConfigs, err = bstore.QueryTx[ProviderConfig](tx).List()
		_checkf(err, "listing provider configs")
	})
	return
}

// ProviderConfigAdd adds a new provider config.
func (x API) ProviderConfigAdd(ctx context.Context, pc ProviderConfig) (npc ProviderConfig) {
	_dbwrite(ctx, func(tx *bstore.Tx) {
		_, err := providerForConfig(pc.ProviderName, pc.ProviderConfigJSON)
		if err != nil && errors.Is(err, errProviderUserError) {
			_checkuserf(err, "checking provider config")
		}
		_checkf(err, "checking provider config")

		err = tx.Insert(&pc)
		_checkf(err, "update providerconfig")

		npc = pc
	})
	return
}

// ProviderConfigUpdate updates a provider config.
func (x API) ProviderConfigUpdate(ctx context.Context, pc ProviderConfig) (npc ProviderConfig) {
	_dbwrite(ctx, func(tx *bstore.Tx) {
		opc := ProviderConfig{Name: pc.Name}
		err := tx.Get(&opc)
		_checkf(err, "get provider config")

		_, err = providerForConfig(pc.ProviderName, pc.ProviderConfigJSON)
		if err != nil && errors.Is(err, errProviderUserError) {
			_checkuserf(err, "checking provider config")
		}
		_checkf(err, "checking provider config")

		err = tx.Update(&pc)
		_checkf(err, "update providerconfig")

		npc = pc
	})
	return
}

func _propagationStates(records []Record) (sets []RecordSet) {
	m, err := propagationStates(time.Now(), records, "", -1, true)
	_checkf(err, "get record sets and propagation states")

	// Ensure we return sets sorted, for tests.
	keys := slices.Collect(maps.Keys(m))
	sort.Slice(keys, func(i, j int) bool {
		a, b := keys[i], keys[j]
		return a.AbsName < b.AbsName || a.AbsName == b.AbsName && a.Type < b.Type
	})
	for _, k := range keys {
		versions := m[k]
		sets = append(sets, versions[len(versions)-1])
	}

	return
}

// ZoneRecordSets returns the current record sets including propagation states that
// are not the latest version but that may still be in caches. For the full history
// of a record set, see ZoneRecordSetHistory.
func (x API) ZoneRecordSets(ctx context.Context, zone string) (sets []RecordSet) {
	_dbread(ctx, func(tx *bstore.Tx) {
		z := _zone(tx, zone)
		records, err := bstore.QueryTx[Record](tx).FilterNonzero(Record{Zone: z.Name}).List()
		_checkf(err, "list records")
		sets = _propagationStates(records)
	})
	return
}

// ZoneRecordSetHistory returns the propagation state history for a record set,
// including the current value.
func (x API) ZoneRecordSetHistory(ctx context.Context, zone, relName string, typ Type) (history []PropagationState) {
	var records []Record

	_dbread(ctx, func(tx *bstore.Tx) {
		z := _zone(tx, zone)
		var err error
		records, err = bstore.QueryTx[Record](tx).FilterNonzero(Record{Zone: z.Name}).List()
		_checkf(err, "list records")
	})

	m, err := propagationStates(time.Now(), records, relName, int(typ), false)
	_checkf(err, "get record sets and propagation states")
	if len(m) != 1 {
		_checkf(fmt.Errorf("got %#v, expected 1 set", m), "get history for record set")
	}
	var versions []RecordSet
	for _, versions = range m {
	}
	if len(versions) == 0 {
		panic(&sherpa.Error{Code: "user:notFound", Message: "record set not found"})
	}
	return versions[len(versions)-1].States
}
