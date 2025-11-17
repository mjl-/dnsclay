// Package libdns-loopia implements a DNS record management client compatible
// with the libdns interfaces for Loopia.
package loopia

import (
	"context"

	"github.com/libdns/libdns"
)

// Provider facilitates DNS record manipulation with Loopia.
type Provider struct {
	client
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Customer string `json:"customer,omitempty"`
	logging  bool   // Enable logging
}

func (p *Provider) SetLogger(logger iLogger) {
	defaultLogger = logger
	Log().Info("Logging enabled")
	p.logging = true
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	ctx = addTrace(ctx, "GetRecords")
	result, err := p.getZoneRecords(ctx, zone)
	if err != nil {
		return result, err
	}

	return result, err
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	ctx = addTrace(ctx, "AppendRecords")
	result, err := p.addDNSEntries(ctx, zone, records)

	return result, err
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	ctx = addTrace(ctx, "SetRecords")
	result, err := p.setRecords(ctx, zone, records)

	return result, err
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	ctx = addTrace(ctx, "DeleteRecords")
	result, err := p.deleteRecords(ctx, zone, records)
	return result, err
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
