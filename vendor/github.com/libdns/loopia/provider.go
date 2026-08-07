// Package libdns-loopia implements a DNS record management client compatible
// with the libdns interfaces for Loopia.
package loopia

import (
	"context"
	"strings"

	"github.com/libdns/libdns"
)

// Provider facilitates DNS record manipulation with Loopia.
type Provider struct {
	client
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Customer string `json:"customer,omitempty"`
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	n, z := loopify("", zone)
	result, err := p.getZoneRecords(ctx, z)
	if err != nil {
		return result, err
	}
	if n != "" {
		filtered := []libdns.Record{}
		for _, r := range result {
			if strings.HasSuffix(r.Name, n) {
				unLoopifyName(n, &r)
				filtered = append(filtered, r)
			}
		}
		return filtered, err
	}
	return result, err
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	n, z := loopifyRecords(zone, records)
	result, err := p.addDNSEntries(ctx, z, records)
	unLoopifyRecords(n, result)
	return result, err
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	hostSuffix, z := loopifyRecords(zone, records)
	result, err := p.setDNSEntries(ctx, z, records)
	unLoopifyRecords(hostSuffix, result)
	return result, err
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	hostSuffix, z := loopifyRecords(zone, records)
	result, err := p.removeDNSEntries(ctx, z, records)
	unLoopifyRecords(hostSuffix, result)
	return result, err
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
