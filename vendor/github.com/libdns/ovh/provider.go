package ovh

import (
	"context"

	"github.com/libdns/libdns"
)

// Provider implements the libdns interfaces for OVH.
type Provider struct {
	Endpoint          string `json:"endpoint,omitempty"`
	ApplicationKey    string `json:"application_key,omitempty"`
	ApplicationSecret string `json:"application_secret,omitempty"`
	ConsumerKey       string `json:"consumer_key,omitempty"`
	client            Client
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	records, err := p.getRecords(ctx, unFQDN(zone))
	if err != nil {
		return nil, err
	}

	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	var appended []libdns.Record

	zone = unFQDN(zone)
	for _, record := range records {
		rec, _, err := p.addRecord(ctx, zone, record)
		if err != nil {
			return nil, err
		}
		appended = append(appended, rec)
	}

	if len(appended) > 0 {
		if err := p.refreshZone(ctx, zone); err != nil {
			return nil, err
		}
	}

	return appended, nil
}

// SetRecords sets the records in the zone by updating existing records or creating new ones.
// It returns the records that were added during the operation.
//
// Since OVH does not support batch operations, this implementation attempts to simulate atomic behavior.
// If any record creation fails, the function attempts to roll back by deleting all records
// that were successfully added during the operation.
//
// If the rollback succeeds, an [AtomicErr] is returned to indicate that the zone remains in a consistent state.
// If the rollback itself fails (e.g., some added records could not be deleted), a non-atomic error is returned,
// and the zone may be left in an inconsistent state.
//
// Similarly, after all record creations have succeeded, any obsolete records are deleted to match the desired state.
// If these deletions fail, a non-atomic error is returned to indicate partial success,
// and the caller should assume the zone may be inconsistent.
//
// This implementation ensures that a nil error is returned **only if** all intended changes
// (additions and deletions) were successfully applied.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	zone = unFQDN(zone)

	added, err := p.setRecords(ctx, zone, records)
	if err != nil {
		return nil, err
	}

	if len(added) > 0 {
		if err := p.refreshZone(ctx, zone); err != nil {
			return nil, err
		}
	}

	return added, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	var deleted []libdns.Record

	zone = unFQDN(zone)
	for _, record := range records {
		recs, err := p.deleteRecords(ctx, zone, record)
		if err != nil {
			return nil, err
		}
		deleted = append(deleted, recs...)
	}

	if len(deleted) > 0 {
		if err := p.refreshZone(ctx, zone); err != nil {
			return nil, err
		}
	}

	return deleted, nil
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
