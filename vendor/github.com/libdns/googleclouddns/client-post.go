package googleclouddns

import (
	"context"
	"fmt"
	"time"

	"github.com/libdns/libdns"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/googleapi"
)

// postCloudDNSRecord will attempt to create a new Google Cloud DNS record set based on the libdns.Records or patch an existing one.
func (p *Provider) postCloudDNSRecord(ctx context.Context, zone string, recordsToSend libdnsRecords) (libdnsRecords, error) {
	if err := p.newService(ctx); err != nil {
		return nil, err
	}
	gcdZone, err := p.getCloudDNSZone(zone)
	if err != nil {
		return nil, err
	}
	if len(recordsToSend) == 0 {
		return nil, fmt.Errorf("no records available to add to zone %s", zone)
	}
	name := recordsToSend[0].Name
	fullName := libdns.AbsoluteName(name, zone)
	rrs := dns.ResourceRecordSet{
		Name:    fullName,
		Rrdatas: make([]string, 0),
		Ttl:     int64(recordsToSend[0].TTL / time.Second),
		Type:    recordsToSend[0].Type,
	}
	rrs.Rrdatas = recordsToSend.prepValuesForCloudDNS()
	googleRecord, err := p.service.ResourceRecordSets.Create(p.Project, gcdZone, &rrs).Context(ctx).Do()
	if err != nil {
		if gErr, ok := err.(*googleapi.Error); !ok || gErr.Code != 409 {
			return nil, err
		}
		// Record exists and we'd really like to get this libdns.Record into the zone so how about we try patching it instead...
		googleRecord, err = p.service.ResourceRecordSets.Patch(p.Project, gcdZone, rrs.Name, rrs.Type, &rrs).Context(ctx).Do()
		if err != nil {
			return nil, err
		}
	}
	return convertToLibDNS(googleRecord, zone), nil
}
