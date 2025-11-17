package loopia

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kolo/xmlrpc"
	"github.com/libdns/libdns"
)

const (
	apiurl = "https://api.loopia.se/RPCSERV"
)

type client struct {
	rpc   *xmlrpc.Client
	mutex sync.Mutex
}

type libdnsKey string

var libdnsKeyTrace libdnsKey = "libdns.loopia.trace"

func writeTrace(ctx context.Context, trace string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, libdnsKeyTrace, trace)
}

func getTrace(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if trace, ok := ctx.Value(libdnsKeyTrace).(string); ok {
		return trace
	}
	return ""
}

// addTrace concats a trace to the existing trace in context. If the context is nil, it creates a new
func addTrace(ctx context.Context, trace string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if t := getTrace(ctx); t != "" {
		trace = fmt.Sprintf("%s -> %s", t, trace)
	}
	return writeTrace(ctx, trace)
}

// cleanZone removes the trailing dot from the zone name if it exists.
func cleanZone(zone string) string {
	return strings.TrimSuffix(zone, ".")
}

func validZone(zone string) bool {
	if zone == "" || len(zone) < 4 {
		return false
	}
	return true
}

func validRecord(r libdns.Record) bool {
	rr := r.RR()

	if rr.Name == "" {
		return false
	}
	if rr.Type == "" {
		return false
	}
	if rr.Data == "" {
		return false
	}
	if rr.TTL < 0 || rr.TTL > (time.Hour*8*24) {
		return false
	}

	return true
}

func (p *Provider) getRPC() *xmlrpc.Client {
	if p.rpc == nil {
		rpc, err := xmlrpc.NewClient(apiurl, nil)
		if err != nil {
			panic(err)
		}
		p.rpc = rpc
	}
	return p.rpc
}

func (p *Provider) call(serviceMethod string, args []interface{}, reply interface{}) error {
	params := []interface{}{
		p.Username,
		p.Password,
	}
	if p.Customer != "" {
		params = append(params, p.Customer)
	}
	params = append(params, args...)
	err := p.getRPC().Call(
		serviceMethod,
		params,
		reply,
	)
	if p.logging {
		Log().Debugw("called rpc", "method", serviceMethod, "params", args, "error", err)
	}
	return err
}

func (p *Provider) getLoopiaRecords(ctx context.Context, zone, name string, records *[]loopiaRecord) error {
	if !validZone(zone) {
		return fmt.Errorf("invalid zone '%s'", zone)
	}
	if name == "" {
		return fmt.Errorf("invalide name '%s'", name)
	}
	if p.logging {
		Log().Debugw("getLoopiaRecords", "zone", zone, "name", name, "trace", getTrace(ctx))
	}
	names := []string{}
	err := p.call("getSubdomains", params(cleanZone(zone)), &names)
	if err != nil {
		return fmt.Errorf("unexpected error getting subdomains: %w", err)
	}
	if len(names) == 0 {
		if p.logging {
			Log().Debugw("no subdomains found", "zone", zone, "name", name, "trace", getTrace(ctx))
		}
		return nil
	}

	// records := []loopiaRecord{}
	if p.logging {
		Log().Debugw("getLoopiaRecords", "zone", zone, "name", name, "trace", getTrace(ctx))
	}
	err = p.call("getZoneRecords", params(zone, name), records)
	if err != nil {
		if p.logging {
			Log().Errorw("error calling getZoneRecords", "err", err, "zone", zone, "name", name, "trace", getTrace(ctx))
		}
		return fmt.Errorf("error calling getZoneRecords: %w", err)
	}
	return nil
}

func (p *Provider) getRecords(ctx context.Context, zone, name string) ([]libdns.Record, error) {
	if p.logging {
		Log().Debugw("getRecords", "zone", zone, "name", name)
		ctx = addTrace(ctx, "getRecords")
	}
	records := []loopiaRecord{}
	if err := p.getLoopiaRecords(ctx, zone, name, &records); err != nil {
		return nil, err
	}

	result := []libdns.Record{}
	for _, r := range records {
		rr, err := r.libdnsRecord(name)
		if err != nil {
			return nil, fmt.Errorf("unexpected error converting record: %w", err)
		}
		result = append(result, rr)
	}
	return result, nil
}

func (p *Provider) addRecord(ctx context.Context, zone string, record libdns.Record, withSubdomain bool) (out libdns.Record, id int64, err error) {
	if p.logging {
		Log().Debugw("addRecord",
			"zone", zone,
			"record", record,
			"withSubdomain", withSubdomain,
		)
		ctx = addTrace(ctx, "addRecord")
	}
	name := record.RR().Name
	loopiaToAdd, err := toLoopiaRecord(record, 0)
	if err != nil {
		return nil, 0, fmt.Errorf("unexpected error converting record: %w", err)
	}
	if withSubdomain {
		var response string
		err := p.call("addSubdomain", params(zone, name), &response)
		if err != nil {
			return nil, 0, fmt.Errorf("unexpected error adding subdomain: %w", err)
		}
	}

	var result string
	if err := p.call("addZoneRecord", params(zone, name, loopiaToAdd), &result); err != nil || result != "OK" {
		return nil, 0, fmt.Errorf("unexpected error adding zone record: %w", err)
	}
	if p.logging {
		Log().Debugw("getting records to fetch ID", "zone", zone, "name", name)
	}
	records := []loopiaRecord{}
	if err := p.getLoopiaRecords(ctx, zone, name, &records); err != nil {
		return nil, 0, fmt.Errorf("unexpected error getting zone records after add: %w", err)
	}

	for _, r := range records {
		out, err = r.libdnsRecord(name)
		if err != nil {
			return nil, 0, fmt.Errorf("unexpected error converting record: %w", err)
		}
		if libdnsRecordEqual(record, out) {
			return out, r.ID, nil
		}

	}
	return nil, 0, fmt.Errorf("unable to retreive new record to get it's ID")
}

func params(args ...interface{}) []interface{} {
	return args
}

func (p *Provider) getZoneRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	if p.logging {
		Log().Debugw("getZoneRecords", "zone", zone)
	}
	if !validZone(zone) {
		return nil, fmt.Errorf("invalide zone '%s'", zone)
	}
	zone = cleanZone(zone)
	names := []string{}
	err := p.call("getSubdomains", params(zone), &names)
	if err != nil {
		return nil, fmt.Errorf("unexpected error getting subdomains: %w", err)
	}
	result := []libdns.Record{}
myloop:
	for _, name := range names {
		select {
		case <-ctx.Done():
			break myloop
		default:
			records, err := p.getRecords(ctx, zone, name)
			if err != nil {
				return nil, fmt.Errorf("error getting zone records for %s: %w", name, err)
			}
			result = append(result, records...)
		}
	}
	return result, err
}

func (p *Provider) addDNSEntries(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	if p.logging {
		Log().Debugw("addDNSEntries",
			"zone", zone,
			"records", len(records),
		)
		ctx = addTrace(ctx, "addDNSEntries")
	}
	if !validZone(zone) {
		return nil, fmt.Errorf("invalide zone '%s'", zone)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("records is nil or empty")
	}
	for i, r := range records {
		if !validRecord(r) {
			return nil, fmt.Errorf("record %d is invalid", i)
		}
	}
	zone = cleanZone(zone)
	result := []libdns.Record{}
	cache := make(map[string][]loopiaRecord)
	subsCreated := make(map[string]bool)
OUTER:
	for _, new := range records {
		rrNew := new.RR()
		select {
		case <-ctx.Done():
			break OUTER
		default:
			if cache[rrNew.Name] == nil {
				existingRecords := []loopiaRecord{}
				err := p.getLoopiaRecords(ctx, zone, rrNew.Name, &existingRecords)
				if err != nil {
					return result, err
				}
				cache[rrNew.Name] = existingRecords
				if p.logging {
					Log().Debugw("cached record", "zone", zone, "name", rrNew.Name, "count", len(existingRecords))
				}
			}
			withSubdomain := false
			if len(cache[rrNew.Name]) == 0 && !subsCreated[rrNew.Name] {
				withSubdomain = true
			}
			for _, existing := range cache[rrNew.Name] {
				if libdnsEqualLoopia(new, existing) {
					if p.logging {
						Log().Debugw("identical record exists, skipping",
							"record", new,
							"id", existing.ID)
					}
					result = append(result, existing.mustLibdnsRecord(rrNew.Name))
					continue OUTER
				}
			}
			if withSubdomain {
				subsCreated[rrNew.Name] = true
			}

			cn, _, err := p.addRecord(ctx, zone, new, withSubdomain)
			if err != nil {
				return result, err
			}
			result = append(result, cn)
		}
	}
	return result, nil
}

// setRecords ensures that for any (name, type) pair in the input is the only
// records in the output zone with that (name, type) pair are those that were
// provided in the input.
func (p *Provider) setRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	ctx = addTrace(ctx, "setRecords")
	for _, r := range records {
		n, z := loopify(r.RR().Name, zone)
		existing := []loopiaRecord{}
		err := p.getLoopiaRecords(ctx, z, n, &existing)
		if err != nil {
			return nil, fmt.Errorf("unexpected error getting zone records: %w", err)
		}
	}
	return nil, errors.New("not implemented")
}

func (p *Provider) updateZoneRecord(ctx context.Context, zone string, record libdns.Record, id int64) (*loopiaRecord, error) {
	if !validZone(zone) {
		return nil, fmt.Errorf("invalide zone '%s'", zone)
	}
	if id == 0 {
		return nil, fmt.Errorf("invalid ID")
	}

	zone = cleanZone(zone)
	updated := mustToLoopiaRecord(record, id)

	var response string
	n, z := loopify(record.RR().Name, zone)
	err := p.call("updateZoneRecord", params(z, n, updated), &response)
	if err != nil {
		return nil, fmt.Errorf("unexpected error updating zone record: %w", err)
	}
	if response != "OK" {
		return nil, fmt.Errorf("unexpected error updating zone record: %s", response)
	}

	return &updated, nil
}

func (p *Provider) deleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	if p.logging {
		Log().Debugw("deleteRecords", "zone", zone, "records", len(records), "trace", getTrace(ctx))
	}
	if !validZone(zone) {
		return nil, fmt.Errorf("invalide zone '%s'", zone)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("records is nil or empty")
	}
	zone = cleanZone(zone)
	ctx = addTrace(ctx, "deleteRecords")
	type args struct {
		zone   string
		name   string
		record loopiaRecord
	}
	toDelete := []args{}
	for i, r := range records {
		n, z := loopify(r.RR().Name, zone)
		ctx2 := addTrace(ctx, fmt.Sprintf("toDelete[%d]", i))
		existing, err := p.getMatchingRecordsByName(ctx2, z, n)
		if err != nil {
			if p.logging {
				Log().Warnw("unexpected error getting remaining records", "err", err, "zone", z, "name", n, "trace", getTrace(ctx2))
			}
			return nil, fmt.Errorf("unexpected error deleting records: %w", err)
		}
		rr := r.RR()
		if len(existing) > 0 {
			for _, er := range existing {
				erl := er.mustLibdnsRecord(rr.Name).RR()

				if rr.Type != "" && rr.Type != erl.Type {
					continue
				}
				if rr.Data != "" && rr.Data != erl.Data {
					continue
				}
				if rr.TTL != 0 && rr.TTL != erl.TTL {
					continue
				}

				toDelete = append(toDelete, args{
					zone:   z,
					name:   n,
					record: er,
				})
			}
		}
	}
	result := []libdns.Record{}
	for _, arg := range toDelete {
		err := p.removeDNSEntry(ctx, arg.zone, arg.name, arg.record.ID)
		if err != nil {
			return nil, fmt.Errorf("unexpected error removing zone record: %w", err)
		}
		result = append(result, arg.record.mustLibdnsRecord(arg.name))
	}

	return result, nil
}

// getMatchingRecordsByName will NOT loopify the name
func (p *Provider) getMatchingRecordsByName(ctx context.Context, zone, name string) ([]loopiaRecord, error) {
	if !validZone(zone) {
		return nil, fmt.Errorf("invalide zone '%s'", zone)
	}
	if name == "" {
		return nil, fmt.Errorf("invalid name '%s'", name)
	}
	ctx = addTrace(ctx, "getMatchingRecordsByName")
	records := []loopiaRecord{}
	err := p.getLoopiaRecords(ctx, zone, name, &records)
	if err != nil {
		return nil, fmt.Errorf("unexpected error getting zone records: %w", err)
	}
	return records, nil
}

func (p *Provider) removeDNSEntry(ctx context.Context, zone, name string, id int64) error {
	if p.logging {
		Log().Debugw("removeDNSEntry", "zone", zone, "name", name, "id", id)
	}
	if !validZone(zone) {
		return fmt.Errorf("invalide zone '%s'", zone)
	}
	if id == 0 {
		return fmt.Errorf("invalid ID")
	}
	ctx = addTrace(ctx, "removeDNSEntry")
	zone = cleanZone(zone)
	var response string
	err := p.call("removeZoneRecord", params(zone, name, id), &response)
	if err != nil {
		return fmt.Errorf("unexpected error removing zone record: %w", err)
	}
	if response != "OK" {
		return fmt.Errorf("unexpected error removing zone record: %s", response)
	}
	records, err := p.getMatchingRecordsByName(ctx, zone, name)
	if err != nil {
		if p.logging {
			Log().Warnw("unexpected error removing zone record", "err", err, "zone", zone, "name", name, "trace", getTrace(ctx))
		}
		return fmt.Errorf("unexpected error removing zone record: %w", err)
	}
	if len(records) == 0 {
		// remove the subdomain if no records left
		var response string
		if p.logging {
			Log().Debugw("removing subdomain", "zone", zone, "name", name, "trace", getTrace(ctx))
		}
		err := p.call("removeSubdomain", params(zone, name), &response)
		if err != nil {
			if p.logging {
				Log().Warnw("unexpected error deleting subdomain", "err", err, "response", response, "trace", getTrace(ctx))
			}
		}
	}
	return nil
}
