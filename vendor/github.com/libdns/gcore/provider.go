// Package gcore implements the libdns interfaces for GCore DNS
package gcore

import (
	"context"
	"fmt"
	"strings"
	"time"

	gcoreSDK "github.com/G-Core/gcore-dns-sdk-go"
	"github.com/libdns/libdns"
)

// qualityRecordNames takes a libdns.Record and a zone, and returns a new record with a name that is fully qualified
// (i.e. it includes the zone name). If the record name does not end with the zone name and a '.', the zone name and '.'
// are appended to the record name. Otherwise the record name is left unchanged.
func qualityRecordNames(record libdns.Record, zone string) libdns.Record {
	return libdns.Record{
		Name:  libdns.AbsoluteName(record.Name, zone),
		Type:  record.Type,
		TTL:   record.TTL,
		Value: record.Value,
	}
}

// unqualifyRecordNames takes a libdns.Record and a zone, and returns a new record with a name that is unqualified
// (i.e. it does not include the zone name). If the record name ends with the zone name and a '.', it is replaced with
// an empty string. Otherwise the record name is left unchanged.
func unqualifyRecordNames(record libdns.Record, zone string) libdns.Record {
	return libdns.Record{
		Name:  libdns.RelativeName(record.Name, zone),
		Type:  record.Type,
		TTL:   record.TTL,
		Value: record.Value,
	}
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
		rrSets, err := cli.RRSet(ctx, zone, gcoreRecord.Name, gcoreRecord.Type)
		if err != nil {
			return nil, err
		}
		for _, rrSet := range rrSets.Records {
			records[i] = libdns.Record{
				Name:  gcoreRecord.Name,
				Type:  gcoreRecord.Type,
				TTL:   time.Duration(gcoreRecord.TTL) * time.Second,
				Value: rrSet.ContentToString(),
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
		recordsByType[record.Type] = append(recordsByType[record.Type], record)
	}

	var addedRecords []libdns.Record

	for recordType, records := range recordsByType {
		for _, record := range records {
			rrSet, err := cli.RRSet(ctx, zone, record.Name, recordType)
			if err != nil {
				if strings.Contains(err.Error(), "404: record is not found") {
					rrSet = gcoreSDK.RRSet{
						Type: recordType,
						TTL:  int(record.TTL.Seconds()),
						Records: []gcoreSDK.ResourceRecord{
							{
								Content: []any{record.Value},
								Enabled: true,
							},
						},
					}
					if err := cli.UpdateRRSet(ctx, zone, record.Name, recordType, rrSet); err != nil {
						return nil, err
					}
					addedRecords = append(addedRecords, record)
					continue
				}
				return nil, err
			}

			for _, rr := range rrSet.Records {
				if rr.ContentToString() == record.Value {
					continue
				}

				rrSet.Records = append(rrSet.Records, gcoreSDK.ResourceRecord{
					Content: []any{record.Value},
					Enabled: true,
				})
			}

			if err := cli.UpdateRRSet(ctx, zone, record.Name, recordType, rrSet); err != nil {
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
		rrSet, err := cli.RRSet(ctx, zone, record.Name, record.Type)
		if err != nil {
			return nil, err
		}

		for _, rr := range rrSet.Records {
			if rr.ContentToString() == record.Value {
				continue
			}

			rrSet.Records = append(rrSet.Records, gcoreSDK.ResourceRecord{
				Content: []any{record.Value},
				Enabled: true,
			})
		}

		if err := cli.UpdateRRSet(ctx, zone, record.Name, record.Type, rrSet); err != nil {
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
		if cli.DeleteRRSetRecord(ctx, zone, record.Name, record.Type, record.Value) != nil {
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
