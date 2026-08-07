// Package linode implements a DNS record management client compatible
// with the libdns interfaces for Linode.
package linode

import (
	"context"
	"fmt"
	"sync"

	"github.com/libdns/libdns"
	"github.com/linode/linodego"
)

// Provider facilitates DNS record manipulation with Linode.
type Provider struct {
	// APIToken is the Linode Personal Access Token, see https://cloud.linode.com/profile/tokens.
	APIToken string `json:"api_token,omitempty"`
	// APIURL is the Linode API hostname, i.e. "api.linode.com".
	APIURL string `json:"api_url,omitempty"`
	// APIVersion is the Linode API version, i.e. "v4".
	APIVersion string `json:"api_version,omitempty"`
	client     linodego.Client
	once       sync.Once
	mutex      sync.Mutex
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.init(ctx)
	domainID, err := p.getDomainIDByZone(ctx, zone)
	if err != nil {
		return nil, fmt.Errorf("could not find domain ID for zone: %s: %v", zone, err)
	}
	records, err := p.listDomainRecords(ctx, zone, domainID)
	if err != nil {
		return nil, err
	}
	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.init(ctx)
	domainID, err := p.getDomainIDByZone(ctx, zone)
	if err != nil {
		return nil, fmt.Errorf("could not find domain ID for zone: %s: %v", zone, err)
	}
	addedRecords := make([]libdns.Record, 0, len(records))
	for _, record := range records {
		addedRecord, err := p.createDomainRecord(ctx, zone, domainID, &record)
		if err != nil {
			return nil, err
		}
		addedRecords = append(addedRecords, *addedRecord)
	}
	return addedRecords, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.init(ctx)
	domainID, err := p.getDomainIDByZone(ctx, zone)
	if err != nil {
		return nil, fmt.Errorf("could not find domain ID for zone: %s: %v", zone, err)
	}
	updatedRecords := make([]libdns.Record, 0, len(records))
	for _, record := range records {
		updatedRecord, err := p.createOrUpdateDomainRecord(ctx, zone, domainID, &record)
		if err != nil {
			return nil, err
		}
		updatedRecords = append(updatedRecords, *updatedRecord)
	}
	return updatedRecords, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.init(ctx)
	domainID, err := p.getDomainIDByZone(ctx, zone)
	if err != nil {
		return nil, fmt.Errorf("could not find domain ID for zone: %s: %v", zone, err)
	}
	deletedRecords := make([]libdns.Record, 0, len(records))
	for _, record := range records {
		err := p.deleteDomainRecord(ctx, domainID, &record)
		if err != nil {
			return nil, err
		}
		deletedRecords = append(deletedRecords, record)
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
