package bunny

import (
	"context"
	"strings"
	"sync"

	"github.com/libdns/libdns"
)

// Provider facilitates DNS record manipulation with Bunny.net
type Provider struct {
	// AccessKey is the Bunny.net API key - see https://docs.bunny.net/reference/bunnynet-api-overview
	AccessKey string                        `json:"access_key"`
	Debug     bool                          `json:"debug"`
	Logger    func(string, []libdns.Record) `json:"-"`

	zones   map[string]bunnyZone `json:"-"`
	zonesMu sync.Mutex
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, domain string) ([]libdns.Record, error) {
	zone, err := p.getZone(ctx, unFQDN(domain))
	if err != nil {
		return nil, err
	}

	records, err := p.getAllRecords(ctx, zone)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, domain string, records []libdns.Record) ([]libdns.Record, error) {
	zone, err := p.getZone(ctx, unFQDN(domain))
	if err != nil {
		return nil, err
	}

	var appendedRecords []libdns.Record
	for _, record := range records {
		newRecord, err := p.createRecord(ctx, zone, record)
		if err != nil {
			return nil, err
		}
		appendedRecords = append(appendedRecords, newRecord)
	}

	return appendedRecords, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, domain string, records []libdns.Record) ([]libdns.Record, error) {
	zone, err := p.getZone(ctx, unFQDN(domain))
	if err != nil {
		return nil, err
	}

	var setRecords []libdns.Record
	for _, record := range records {
		setRecord, err := p.createOrUpdateRecord(ctx, zone, record)
		if err != nil {
			return setRecords, err
		}
		setRecords = append(setRecords, setRecord)
	}

	return setRecords, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, domain string, records []libdns.Record) ([]libdns.Record, error) {
	zone, err := p.getZone(ctx, unFQDN(domain))
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		err := p.deleteRecord(ctx, zone, record)
		if err != nil {
			return nil, err
		}
	}

	return records, nil
}

// ListZones returns the list of available DNS zones.
func (p *Provider) ListZones(ctx context.Context) ([]libdns.Zone, error) {
	bunnyZones, err := p.getAllZones(ctx)
	if err != nil {
		return nil, err
	}

	zones := make([]libdns.Zone, len(bunnyZones))
	for i, bunnyZone := range bunnyZones {
		zones[i] = libdns.Zone{
			Name: bunnyZone.Domain,
		}
	}

	return zones, nil
}

// unFQDN trims any trailing "." from fqdn. Bunny.net's API does not use FQDNs.
func unFQDN(fqdn string) string {
	return strings.TrimSuffix(fqdn, ".")
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
	_ libdns.ZoneLister     = (*Provider)(nil)
)
