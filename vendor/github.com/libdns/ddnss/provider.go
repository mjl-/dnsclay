// Package libdnstemplate implements a DNS record management client compatible
// with the libdns interfaces for ddnss.
package ddnss

import (
	"context"
	"sync"

	"github.com/libdns/libdns"
)

// Provider facilitates DNS record manipulation with ddnss.
type Provider struct {
	APIToken string `json:"api_token"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`

	mutex sync.Mutex
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	domains, err := p.getDomainsFromWebinterface()
	if err != nil {
		return nil, err
	}

	records := []libdns.Record{}
	for _, domain := range domains {
		records = append(records, libdns.Record{
			ID:    domain,
			Name:  domain,
			Type:  domain,
			Value: domain,
		})
	}

	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {

	var appendedRecords []libdns.Record

	for _, rec := range records {
		err := p.setRecord(ctx, zone, rec, false)
		if err != nil {
			return nil, err
		}
		appendedRecords = append(appendedRecords, rec)
	}

	return appendedRecords, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {

	var setRecords []libdns.Record

	for _, rec := range records {
		err := p.setRecord(ctx, zone, rec, false)
		if err != nil {
			return nil, err
		}
		setRecords = append(setRecords, rec)
	}

	return setRecords, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {

	var deletedRecords []libdns.Record

	for _, rec := range records {
		err := p.setRecord(ctx, zone, rec, true)
		if err != nil {
			return nil, err
		}
		deletedRecords = append(deletedRecords, rec)
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
