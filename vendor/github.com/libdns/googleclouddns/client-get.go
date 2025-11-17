package googleclouddns

import (
	"context"
	"fmt"

	"github.com/libdns/libdns"
	"google.golang.org/api/dns/v1"
)

// getCloudDNSRecords returns all the records for the specified zone. It breaks up a single Google Record
// with multiple Values into separate libdns.Records.
func (p *Provider) getCloudDNSRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	if err := p.newService(ctx); err != nil {
		return nil, err
	}

	gcdZone, err := p.getCloudDNSZone(zone)
	if err != nil {
		return nil, err
	}
	rrsReq := p.service.ResourceRecordSets.List(p.Project, gcdZone)
	records := make([]libdns.Record, 0)
	if err := rrsReq.Pages(ctx, func(page *dns.ResourceRecordSetsListResponse) error {
		for _, googleRecord := range page.Rrsets {
			convertedRecords, err := convertToLibDNS(googleRecord, zone)
			if err != nil {
				return fmt.Errorf("error converting to libdns records: %w", err)
			}
			records = append(records, convertedRecords...)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return records, nil
}

// getCloudDNSRecord returns the record for the specified zone, name, and type. It breaks up a single Cloud DNS Record
// with multiple Values into separate libdns.Records.
func (p *Provider) getCloudDNSRecord(ctx context.Context, zone, name, recordType string) (libdnsRecords, error) {
	if err := p.newService(ctx); err != nil {
		return nil, err
	}
	gcdZone, err := p.getCloudDNSZone(zone)
	if err != nil {
		return nil, err
	}
	fullName := libdns.AbsoluteName(name, zone)
	rrs, err := p.service.ResourceRecordSets.Get(p.Project, gcdZone, fullName, recordType).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	return convertToLibDNS(rrs, zone)
}
