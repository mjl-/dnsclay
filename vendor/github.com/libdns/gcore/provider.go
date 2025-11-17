// Package gcore implements the libdns interfaces for GCore DNS
package gcore

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	gcoreSDK "github.com/G-Core/gcore-dns-sdk-go"
	"github.com/libdns/libdns"
)

// qualityRecordNames takes a libdns.Record and a zone, and returns a new record with a name that is fully qualified
// (i.e. it includes the zone name). If the record name does not end with the zone name and a '.', the zone name and '.'
// are appended to the record name. Otherwise the record name is left unchanged.
func qualityRecordNames(record libdns.Record, zone string) libdns.Record {
	rr, err := record.RR().Parse()
	if err != nil {
		log.Printf("error parsing record: %v", err)
		return record
	}

	if addr, isAddress := rr.(libdns.Address); isAddress {
		return libdns.Address{
			Name: libdns.AbsoluteName(addr.Name, zone),
			TTL:  addr.TTL,
			IP:   addr.IP,
		}
	} else if cname, isCNAME := rr.(libdns.CNAME); isCNAME {
		return libdns.CNAME{
			Name:   libdns.AbsoluteName(cname.Name, zone),
			TTL:    cname.TTL,
			Target: cname.Target,
		}
	} else if txt, isPTR := rr.(libdns.TXT); isPTR {
		return libdns.TXT{
			Name: libdns.AbsoluteName(txt.Name, zone),
			TTL:  txt.TTL,
			Text: txt.Text,
		}
	} else if mx, isMX := rr.(libdns.MX); isMX {
		return libdns.MX{
			Name:       libdns.AbsoluteName(mx.Name, zone),
			TTL:        mx.TTL,
			Preference: mx.Preference,
			Target:     mx.Target,
		}
	} else if ns, isNS := rr.(libdns.NS); isNS {
		return libdns.NS{
			Name:   libdns.AbsoluteName(ns.Name, zone),
			TTL:    ns.TTL,
			Target: ns.Target,
		}
	} else if srv, isSRV := rr.(libdns.SRV); isSRV {
		return libdns.SRV{
			Name:      libdns.AbsoluteName(srv.Name, zone),
			TTL:       srv.TTL,
			Service:   srv.Service,
			Transport: srv.Transport,
			Priority:  srv.Priority,
			Weight:    srv.Weight,
			Target:    srv.Target,
			Port:      srv.Port,
		}
	}
	log.Printf("[qualifyRecordNames] type: %s name: %s", record.RR().Type, record.RR().Name)

	return record
}

// unqualifyRecordNames takes a libdns.Record and a zone, and returns a new record with a name that is unqualified
// (i.e. it does not include the zone name). If the record name ends with the zone name and a '.', it is replaced with
// an empty string. Otherwise the record name is left unchanged.
func unqualifyRecordNames(record libdns.Record, zone string) libdns.Record {
	rr, err := record.RR().Parse()
	if err != nil {
		log.Printf("error parsing record: %v", err)
		return record
	}

	if addr, isAddress := rr.(libdns.Address); isAddress {
		return libdns.Address{
			Name: libdns.RelativeName(addr.Name, zone),
			TTL:  addr.TTL,
			IP:   addr.IP,
		}
	} else if cname, isCNAME := rr.(libdns.CNAME); isCNAME {
		return libdns.CNAME{
			Name:   libdns.RelativeName(cname.Name, zone),
			TTL:    cname.TTL,
			Target: cname.Target,
		}
	} else if txt, isPTR := rr.(libdns.TXT); isPTR {
		return libdns.TXT{
			Name: libdns.RelativeName(txt.Name, zone),
			TTL:  txt.TTL,
			Text: txt.Text,
		}
	} else if mx, isMX := rr.(libdns.MX); isMX {
		return libdns.MX{
			Name:       libdns.RelativeName(mx.Name, zone),
			TTL:        mx.TTL,
			Preference: mx.Preference,
			Target:     mx.Target,
		}
	} else if ns, isNS := rr.(libdns.NS); isNS {
		return libdns.NS{
			Name:   libdns.RelativeName(ns.Name, zone),
			TTL:    ns.TTL,
			Target: ns.Target,
		}
	} else if srv, isSRV := rr.(libdns.SRV); isSRV {
		return libdns.SRV{
			Name:      libdns.RelativeName(srv.Name, zone),
			TTL:       srv.TTL,
			Service:   srv.Service,
			Transport: srv.Transport,
			Priority:  srv.Priority,
			Weight:    srv.Weight,
			Target:    srv.Target,
			Port:      srv.Port,
		}
	}
	log.Printf("[unqualifyRecordNames] type: %s name: %s", record.RR().Type, record.RR().Name)

	return record
}

// Provider facilitates DNS record manipulation with GCore DNS.
type Provider struct {
	APIKey string `json:"api_key,omitempty"`
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	cli := gcoreSDK.NewClient(gcoreSDK.PermanentAPIKeyAuth(p.APIKey))

	// Get records for zone and convert to libdns records
	gcoreZone, err := cli.Zone(ctx, zone)
	if err != nil {
		return nil, err
	}

	records := make([]libdns.Record, len(gcoreZone.Records))
	for i, gcoreRecord := range gcoreZone.Records {
		rrSets, err := cli.RRSet(ctx, zone, gcoreRecord.Name, gcoreRecord.Type, -1, 0)
		if err != nil {
			return nil, err
		}
		for _, rrSet := range rrSets.Records {
			records[i] = libdns.RR{
				Name: gcoreRecord.Name,
				Type: gcoreRecord.Type,
				TTL:  time.Duration(gcoreRecord.TTL) * time.Second,
				Data: rrSet.ContentToString(),
			}
		}
	}

	for i, record := range records {
		records[i] = unqualifyRecordNames(record, zone)
	}

	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	cli := gcoreSDK.NewClient(gcoreSDK.PermanentAPIKeyAuth(p.APIKey))

	for i, record := range records {
		records[i] = qualityRecordNames(record, zone)
	}

	recordsByType := make(map[string][]libdns.Record)
	for _, record := range records {
		recordsByType[record.RR().Type] = append(recordsByType[record.RR().Type], record)
	}

	var addedRecords []libdns.Record

	for recordType, records := range recordsByType {
		for _, record := range records {
			rrSet, err := cli.RRSet(ctx, zone, record.RR().Name, recordType, -1, 0)
			if err != nil {
				if strings.Contains(err.Error(), "404: record is not found") {
					rrSet = gcoreSDK.RRSet{
						Type: recordType,
						TTL:  int(record.RR().TTL.Seconds()),
						Records: []gcoreSDK.ResourceRecord{
							{
								Content: []any{record.RR().Data},
								Enabled: true,
							},
						},
					}
					if err := cli.UpdateRRSet(ctx, zone, record.RR().Name, recordType, rrSet); err != nil {
						return nil, err
					}
					addedRecords = append(addedRecords, record)
					continue
				}
				return nil, err
			}

			for _, rr := range rrSet.Records {
				if rr.ContentToString() == record.RR().Data {
					continue
				}

				rrSet.Records = append(rrSet.Records, gcoreSDK.ResourceRecord{
					Content: []any{record.RR().Data},
					Enabled: true,
				})
			}

			if err := cli.UpdateRRSet(ctx, zone, record.RR().Name, recordType, rrSet); err != nil {
				return nil, err
			}
			addedRecords = append(addedRecords, record)
		}
	}

	for i, record := range addedRecords {
		addedRecords[i] = unqualifyRecordNames(record, zone)
	}

	return addedRecords, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	cli := gcoreSDK.NewClient(gcoreSDK.PermanentAPIKeyAuth(p.APIKey))

	for i, record := range records {
		records[i] = qualityRecordNames(record, zone)
	}

	var updatedRecords []libdns.Record

	for _, record := range records {
		rrSet, err := cli.RRSet(ctx, zone, record.RR().Name, record.RR().Type, -1, 0)
		if err != nil {
			return nil, err
		}

		for _, rr := range rrSet.Records {
			if rr.ContentToString() == record.RR().Data {
				continue
			}

			rrSet.Records = append(rrSet.Records, gcoreSDK.ResourceRecord{
				Content: []any{record.RR().Data},
				Enabled: true,
			})
		}

		if err := cli.UpdateRRSet(ctx, zone, record.RR().Name, record.RR().Type, rrSet); err != nil {
			return nil, err
		}

		updatedRecords = append(updatedRecords, record)
	}

	for i, record := range updatedRecords {
		updatedRecords[i] = unqualifyRecordNames(record, zone)
	}

	return updatedRecords, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	cli := gcoreSDK.NewClient(gcoreSDK.PermanentAPIKeyAuth(p.APIKey))

	for i, record := range records {
		records[i] = qualityRecordNames(record, zone)
	}

	var deletedRecords []libdns.Record

	for _, record := range records {
		if cli.DeleteRRSetRecord(ctx, zone, record.RR().Name, record.RR().Type, record.RR().Data) != nil {
			return nil, fmt.Errorf("failed to delete record %v", record)
		}
		deletedRecords = append(deletedRecords, record)
	}

	for i, record := range deletedRecords {
		deletedRecords[i] = unqualifyRecordNames(record, zone)
	}

	return deletedRecords, nil
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
