package dreamhost

import (
	"context"
	"sync"

	"github.com/adamantal/go-dreamhost/api"
	"github.com/libdns/libdns"
)

// Provider facilitates DNS record manipulation with Dreamhost.
type Provider struct {
	APIKey string `json:"api_key,omitempty"`
	client api.Client
	once   sync.Once
	mutex  sync.Mutex
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	err := p.init()
	if err != nil {
		return nil, err
	}

	var records []libdns.Record

	apiRecords, err := p.client.ListDNSRecords(ctx)
	if err != nil {
		return nil, err
	}

	// translate each Dreamhost Domain Record to a libdns Record
	for _, rec := range apiRecords {
		if rec.Zone == zone {
			records = append(records, recordFromApiDnsRecord(rec))
		}

	}

	return records, nil
}

func (p *Provider) addDNSRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	var createdRecords []libdns.Record

	for _, record := range records {
		apiInputRecord := apiDnsRecordInputFromRecord(record, zone)
		err := p.client.AddDNSRecord(ctx, apiInputRecord)
		if err == nil {
			createdRecords = append(createdRecords, record)
		}
	}

	return createdRecords, nil
}

func (p *Provider) removeDNSRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	var deletedRecords []libdns.Record

	for _, record := range records {
		apiInputRecord := apiDnsRecordInputFromRecord(record, zone)
		err := p.client.RemoveDNSRecord(ctx, apiInputRecord)
		if err == nil {
			deletedRecords = append(deletedRecords, record)
		}
	}

	return deletedRecords, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	err := p.init()
	if err != nil {
		return nil, err
	}

	return p.addDNSRecords(ctx, zone, records)
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	err := p.init()
	if err != nil {
		return nil, err
	}

	return p.addDNSRecords(ctx, zone, records)
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	err := p.init()
	if err != nil {
		return nil, err
	}

	return p.removeDNSRecords(ctx, zone, records)
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
