// Package libdnsnamesilo implements a DNS record management client compatible
// with the libdns interfaces for Namesilo.
package namesilo

import (
	"context"
	"fmt"
	"strings"

	"github.com/libdns/libdns"
)

// Provider facilitates DNS record manipulation with Namesilo.
type Provider struct {
	APIToken string `json:"api_token,omitempty"`
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	namesiloRecords, err := p.getDNSRecords(ctx, zone)
	if err != nil {
		return nil, err
	}

	libdnsRecords := make([]libdns.Record, 0, len(namesiloRecords))
	for _, record := range namesiloRecords {
		libdnsRecord, err := record.toLibDNS(zone)
		if err != nil {
			return libdnsRecords, fmt.Errorf("parsing Namesilo DNS record %+v: %v", record, err)
		}
		libdnsRecords = append(libdnsRecords, libdnsRecord)
	}

	return libdnsRecords, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	appendedRecords := make([]libdns.Record, 0, len(records))

	for _, record := range records {
		rec, err := namesiloRecord(zone, record)
		if err != nil {
			return appendedRecords, err
		}

		if err := p.addDNSRecord(ctx, zone, rec); err != nil {
			return appendedRecords, err
		}

		appendedRecords = append(appendedRecords, record)
	}

	return appendedRecords, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	setRecords := make([]libdns.Record, 0, len(records))

	for _, record := range records {
		rr := record.RR()
		recordId, err := p.findRecordId(ctx, zone, rr.Name, rr.Type)
		if err != nil {
			return setRecords, err
		}

		rec, err := namesiloRecord(zone, record)
		if err != nil {
			return setRecords, err
		}

		if recordId == "" {
			if err := p.addDNSRecord(ctx, zone, rec); err != nil {
				return setRecords, err
			}
		} else {
			if err := p.updateDNSRecord(ctx, zone, recordId, rec); err != nil {
				return setRecords, err
			}
		}

		setRecords = append(setRecords, record)
	}

	return setRecords, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	deletedRecords := make([]libdns.Record, 0, len(records))

	for _, record := range records {
		rr := record.RR()
		recordId, err := p.findRecordId(ctx, zone, rr.Name, rr.Type, rr.Data)
		if err != nil {
			return deletedRecords, err
		}

		if recordId == "" {
			return deletedRecords, fmt.Errorf("record not found: %+v", rr)
		}

		if err := p.deleteDNSRecord(ctx, zone, recordId); err != nil {
			return deletedRecords, err
		}

		deletedRecords = append(deletedRecords, record)
	}

	return deletedRecords, nil
}

func (p *Provider) findRecordId(ctx context.Context, zone string, recordName string, recordType string, recordValue ...string) (string, error) {
	records, err := p.getDNSRecords(ctx, zone)
	if err != nil {
		return "", err
	}

	for _, rec := range records {
		libdnsRec, err := rec.toLibDNS(zone)
		if err != nil {
			return "", err
		}

		if recordType == libdnsRec.RR().Type && recordName == libdnsRec.RR().Name {
			if len(recordValue) > 0 && recordValue[0] != "" && recordValue[0] != libdnsRec.RR().Data {
				continue
			}
			return rec.ID, nil
		}
	}

	return "", nil
}

func zoneToDomain(zone string) string {
	return strings.TrimSuffix(zone, ".")
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
