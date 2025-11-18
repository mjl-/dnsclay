// Package libdnsspaceship implements a DNS record management client compatible
// with the libdns interfaces for Spaceship. This package allows you to manage
// DNS records using the Spaceship DNS API.
package libdnsspaceship

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/libdns/libdns"
)

// convertToLibdnsRecord moved to conversions.go

// convertFromLibdnsRecord moved to conversions.go

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	if err := p.validateCredentials(); err != nil {
		return nil, err
	}

	// Clean zone name
	zone = strings.TrimSuffix(zone, ".")

	var records []libdns.Record
	// API requires pagination parameters 'take' and 'skip'. We'll page through all records.
	take := 100
	if p.PageSize > 0 {
		take = p.PageSize
	}
	skip := 0
	for {
		endpoint := fmt.Sprintf("/v1/dns/records/%s?take=%d&skip=%d", zone, take, skip)
		body, _, err := p.doRequest(ctx, "GET", endpoint, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get records: %w", err)
		}
		var lr listResponse
		if err := json.Unmarshal(body, &lr); err != nil {
			return nil, fmt.Errorf("failed to unmarshal records response: %w", err)
		}
		for _, sr := range lr.Items {
			if record := p.toLibdnsRR(sr, zone); record != nil {
				records = append(records, record)
			}
		}
		if skip+len(lr.Items) >= lr.Total {
			break
		}
		skip += take
	}

	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	if err := p.validateCredentials(); err != nil {
		return nil, err
	}

	// Clean zone name
	zone = strings.TrimSuffix(zone, ".")

	var items []spaceshipRecordUnion
	for _, r := range records {
		if item := p.fromLibdnsRR(r, zone); item != nil {
			items = append(items, *item)
		}
	}

	payload := map[string]interface{}{
		"force": false,
		"items": items,
	}

	endpoint := fmt.Sprintf("/v1/dns/records/%s", zone)
	_, status, err := p.doRequest(ctx, "PUT", endpoint, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to append records: %w", err)
	}
	if status != 204 {
		// In case API returns body with created data we could parse it; but it should be 204
		// Fall back to returning the input records
	}

	// Return records converted from the request payload as the representation of what was created
	var added []libdns.Record
	for _, it := range items {
		if record := p.toLibdnsRR(it, zone); record != nil {
			added = append(added, record)
		}
	}
	return added, nil
}

// SetRecords sets the records in the zone by saving the provided records (force update).
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	if err := p.validateCredentials(); err != nil {
		return nil, err
	}

	zone = strings.TrimSuffix(zone, ".")
	var items []spaceshipRecordUnion
	for _, r := range records {
		if item := p.fromLibdnsRR(r, zone); item != nil {
			items = append(items, *item)
		}
	}
	payload := map[string]interface{}{
		"force": true,
		"items": items,
	}
	endpoint := fmt.Sprintf("/v1/dns/records/%s", zone)
	_, status, err := p.doRequest(ctx, "PUT", endpoint, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to set records: %w", err)
	}
	if status != 204 {
		// API should return 204. If not, still return input records as best-effort.
	}
	var updated []libdns.Record
	for _, it := range items {
		if record := p.toLibdnsRR(it, zone); record != nil {
			updated = append(updated, record)
		}
	}
	return updated, nil
}

// DeleteRecords deletes the specified records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	if err := p.validateCredentials(); err != nil {
		return nil, err
	}

	zone = strings.TrimSuffix(zone, ".")
	var items []spaceshipRecordUnion
	for _, rec := range records {
		item := p.fromLibdnsRR(rec, zone)
		if item == nil {
			rr := rec.RR()
			return nil, fmt.Errorf("unsupported record type for deletion: %s", rr.Type)
		}
		items = append(items, *item)
	}
	endpoint := fmt.Sprintf("/v1/dns/records/%s", zone)
	_, status, err := p.doRequest(ctx, "DELETE", endpoint, items)
	if err != nil {
		return nil, fmt.Errorf("failed to delete records: %w", err)
	}
	if status != 204 {
		// API should return 204. If not, proceed anyway.
	}
	return records, nil
}


// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
