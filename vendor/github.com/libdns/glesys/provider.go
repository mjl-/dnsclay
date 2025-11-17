// SPDX-FileCopyrightText: 2022 Peter Magnusson <me@kmpm.se>
// SPDX-License-Identifier: MIT
package glesys

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/libdns/glesys/internal/impl"
	"github.com/libdns/libdns"
)

var debug bool

const _DebugKey_ = "LIBDNS_GLESYS_DEBUG"

func init() {
	if b, err := strconv.ParseBool(os.Getenv(_DebugKey_)); err == nil && b {
		debug = true
	}
}

type Provider struct {
	mutex       sync.Mutex
	clientCache *impl.Client
	Project     string `json:"project,omitempty"`
	APIKey      string `json:"api_key,omitempty"`
}

func (p *Provider) client() *impl.Client {
	if p.clientCache == nil {
		p.clientCache = impl.NewClient(p.Project, p.APIKey, "libdns-glesys/0.0.2")
	}
	return p.clientCache
}

type recordWithMatchingGlesys struct {
	Record  libdns.Record
	Matches []impl.DNSDomainRecord
}

// getMatchingRecords returns records with the matches to the given records.
// It compares the records in the zone with the records in the input.
// Check for .Matches to be empty or not.
func (p *Provider) getMatchingRecords(ctx context.Context, zone string, records []libdns.Record) ([]recordWithMatchingGlesys, error) {
	zone = cleanZ(zone)
	if debug {
		log.Printf("getMatchingRecords zone=%s", zone)
	}
	existingRecords, err := p.client().DNSDomains.ListRecords(ctx, zone)
	if err != nil {
		return nil, err
	}
	results := []recordWithMatchingGlesys{}
	for _, r := range records {
		rr := r.RR()
		matches := []impl.DNSDomainRecord{}
		for _, dr := range *existingRecords {
			hasMatching := checkParamsMatching(rr, &dr)
			if !hasMatching.all() {
				continue
			}
			matches = append(matches, dr)
		}
		results = append(results, recordWithMatchingGlesys{Record: r, Matches: matches})
	}
	if debug {
		log.Printf("getMatchingRecords result: %+v", results)
	}
	return results, nil
}

type recordWithMatchingLibDNS struct {
	Record  impl.DNSDomainRecord
	Matches []libdns.Record
}

func (p *Provider) getMatchingRecordsLibDNS(ctx context.Context, zone string, records []libdns.Record) ([]recordWithMatchingLibDNS, error) {
	zone = cleanZ(zone)
	if debug {
		log.Printf("getMatchingRecordsLibDNS zone=%s", zone)
	}
	existingRecords, err := p.client().DNSDomains.ListRecords(ctx, zone)
	if err != nil {
		return nil, err
	}
	results := []recordWithMatchingLibDNS{}
	for _, dr := range *existingRecords {
		matches := []libdns.Record{}
		for _, r := range records {
			rr := r.RR()
			hasMatching := checkParamsMatching(rr, &dr)
			if !hasMatching.all() {
				continue
			}
			matches = append(matches, r)
		}
		results = append(results, recordWithMatchingLibDNS{Record: dr, Matches: matches})
	}
	if debug {
		log.Printf("getMatchingRecordsLibDNS result: %+v", results)
	}
	return results, nil
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	zone = cleanZ(zone)
	if debug {
		log.Printf("GetRecords zone=%s", zone)
	}
	drs, err := p.client().DNSDomains.ListRecords(ctx, zone)
	if err != nil {
		return nil, err
	}
	records := make([]libdns.Record, len(*drs))
	for i, dr := range *drs {
		if zone != dr.DomainName {
			return records, fmt.Errorf("unexpected domainname in respose: %v", dr.DomainName)
		}
		r, err := toLibDNS(&dr)
		if err != nil {
			return nil, err
		}
		records[i] = r
	}
	if debug {
		log.Printf("GetRecords result: %+v", records)
	}
	return records, nil

}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	zone = cleanZ(zone)
	if debug {
		log.Printf("AppendRecords zone=%s", zone)
	}
	results := []libdns.Record{}
	for _, r := range records {
		rr := r.RR()
		param := impl.AddRecordParams{
			DomainName: zone,
			Host:       rr.Name,
			Data:       rr.Data,
			TTL:        int(rr.TTL / time.Second),
			Type:       strings.ToUpper(rr.Type),
		}
		// if r.Priority > 0 {
		// 	param.Data = fmt.Sprintf("%d %s", r.Priority, param.Data)
		// }

		dr, err := p.client().DNSDomains.AddRecord(ctx, param)
		if err != nil {
			return results, err
		}
		if rr, err := toLibDNS(dr); err != nil {
			return results, err
		} else {
			results = append(results, rr)
		}
	}
	if debug {
		log.Printf("AppendRecords result: %+v", results)
	}
	return results, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// In RFC 9499 terms, SetRecords appends, modifies, or deletes records in the
// zone so that for each RRset in the input, the records provided in the input
// are the only members of their RRset in the output zone.
// Calls to SetRecords are presumed to be atomic;
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	zone = cleanZ(zone)
	if debug {
		log.Printf("SetRecords zone=%s", zone)
	}

	type updateChange struct {
		From impl.DNSDomainRecord
		To   impl.DNSDomainRecord
	}

	type changes struct {
		updates   []updateChange
		deletes   []impl.DNSDomainRecord
		additions []impl.DNSDomainRecord
	}

	wanted := changes{
		updates:   []updateChange{},
		deletes:   []impl.DNSDomainRecord{},
		additions: []impl.DNSDomainRecord{},
	}
	executed := changes{
		updates:   []updateChange{},
		deletes:   []impl.DNSDomainRecord{},
		additions: []impl.DNSDomainRecord{},
	}

	makeAfter := func(inErr error) ([]libdns.Record, error) {
		if inErr != nil {
			// revert the changes
			for _, dr := range wanted.deletes {
				// add the record back
				param := impl.AddRecordParams{
					DomainName: zone,
					Host:       dr.Host,
					Data:       dr.Data,
					TTL:        dr.TTL,
					Type:       strings.ToUpper(dr.Type),
				}
				_, err := p.client().DNSDomains.AddRecord(ctx, param)
				if err != nil {
					return nil, err
				}
			}
			for _, dr := range wanted.updates {
				// update the record back
				param := impl.UpdateRecordParams{
					RecordID: dr.From.RecordID,
					Host:     dr.From.Host,
					Data:     dr.From.Data,
					TTL:      dr.From.TTL,
					Type:     strings.ToUpper(dr.From.Type),
				}
				_, err := p.client().DNSDomains.UpdateRecord(ctx, param)
				if err != nil {
					return nil, err
				}
			}
			for _, dr := range wanted.additions {
				// delete the record
				err := p.client().DNSDomains.DeleteRecord(ctx, dr.RecordID)
				if err != nil {
					return nil, err
				}
			}
			return nil, inErr
		}
		after := []libdns.Record{}
		for _, dr := range wanted.additions {
			after = append(after, mustToLibDNS(&dr))
		}
		for _, dr := range wanted.updates {
			after = append(after, mustToLibDNS(&dr.To))
		}
		return after, nil
	}

	// find the records that need to be updated or deleted
	matching, err := p.getMatchingRecords(ctx, zone, records)
	if err != nil {
		return makeAfter(err)
	}
	for _, m := range matching {
		if len(m.Matches) == 0 {
			// no matches, record needs to be added
			wanted.additions = append(wanted.additions, impl.DNSDomainRecord{
				DomainName: zone,
				Host:       m.Record.RR().Name,
				Data:       m.Record.RR().Data,
				TTL:        int(m.Record.RR().TTL / time.Second),
				Type:       strings.ToUpper(m.Record.RR().Type),
			})
			continue
		}
		for _, dr := range m.Matches {
			if dr.Type == m.Record.RR().Type && dr.Host == m.Record.RR().Name && dr.Data == m.Record.RR().Data && dr.TTL == int(m.Record.RR().TTL/time.Second) {
				// record already exists, no need to update
				continue
			}
			// record needs to be updated
			wanted.updates = append(wanted.updates, updateChange{
				From: dr,
				To: impl.DNSDomainRecord{
					DomainName: dr.DomainName,
					RecordID:   dr.RecordID,
					Host:       dr.Host,
					Data:       m.Record.RR().Data,
					TTL:        int(m.Record.RR().TTL / time.Second),
					Type:       strings.ToUpper(m.Record.RR().Type),
				},
			})
		}
		// record needs to be deleted
		wanted.deletes = append(wanted.deletes, m.Matches...)

	}

	// changes needs to be atomic so in case of failure we need to
	// revert the changes
	// we need to delete the records that are not in the wanted list

	// delete the records that need to be deleted
	for _, dr := range wanted.deletes {
		err = p.client().DNSDomains.DeleteRecord(ctx, dr.RecordID)
		if err != nil {
			return makeAfter(err)
		}
		executed.deletes = append(executed.deletes, dr)
	}
	// update the records that need to be updated
	for _, dr := range wanted.updates {
		param := impl.UpdateRecordParams{
			RecordID: dr.From.RecordID,
			Host:     dr.To.Host,
			Data:     dr.To.Data,
			TTL:      dr.To.TTL,
			Type:     strings.ToUpper(dr.To.Type),
		}
		_, err = p.client().DNSDomains.UpdateRecord(ctx, param)
		if err != nil {
			return makeAfter(err)
		}
		executed.updates = append(executed.updates, updateChange{
			From: dr.From,
			To:   dr.To,
		})
	}
	// add the records that need to be added
	for _, dr := range wanted.additions {
		param := impl.AddRecordParams{
			DomainName: zone,
			Host:       dr.Host,
			Data:       dr.Data,
			TTL:        dr.TTL,
			Type:       strings.ToUpper(dr.Type),
		}
		dr, err := p.client().DNSDomains.AddRecord(ctx, param)
		if err != nil {
			return makeAfter(err)
		}
		executed.additions = append(executed.additions, *dr)
	}

	return makeAfter(nil)
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	zone = cleanZ(zone)
	if debug {
		log.Printf("DeleteRecords zone=%s", zone)
	}
	matching, err := p.getMatchingRecords(ctx, zone, records)
	if err != nil {
		return nil, err
	}
	results := []libdns.Record{}
	for _, m := range matching {
		for _, dr := range m.Matches {
			err = p.client().DNSDomains.DeleteRecord(ctx, dr.RecordID)
			if err != nil {
				return results, err
			}
			r, err := toLibDNS(&dr)
			if err != nil {
				return results, err
			}
			results = append(results, r)
		}
	}
	if debug {
		log.Printf("DeleteRecords results: %+v", results)
	}
	return results, nil
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
