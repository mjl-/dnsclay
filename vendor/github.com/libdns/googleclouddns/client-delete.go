package googleclouddns

import (
	"context"
	"time"

	"github.com/libdns/libdns"
	"google.golang.org/api/dns/v1"
)

// deleteCloudDNSRecord will delete the specified record set.
func (p *Provider) deleteCloudDNSRecord(ctx context.Context, zone, name, recordType string, recordsToDelete, existingRecords libdnsRecords) (libdnsRecords, error) {
	if err := p.newService(ctx); err != nil {
		return nil, err
	}
	fullName := libdns.AbsoluteName(name, zone)
	gcdZone, err := p.getCloudDNSZone(zone)
	if err != nil {
		return nil, err
	}
	updatedRecordList := make(libdnsRecords, 0) // a list of records, if any, to keep for the Cloud DNS entry
	for _, record := range existingRecords {
		if recordsToDelete.doesNotHaveRecord(record) {
			updatedRecordList = append(updatedRecordList, record)
			continue
		}
	}
	if len(updatedRecordList) == 0 { // No records left with Cloud DNS entry, delete the whole thing
		_, err = p.service.ResourceRecordSets.Delete(p.Project, gcdZone, fullName, recordType).Context(ctx).Do()
		return recordsToDelete, err
	}
	// Let's patch the existing entry with the records left
	rrs := dns.ResourceRecordSet{
		Name:    fullName,
		Rrdatas: make([]string, 0),
		Ttl:     int64(updatedRecordList[0].TTL / time.Second),
		Type:    recordType,
	}
	rrs.Rrdatas = updatedRecordList.prepValuesForCloudDNS()
	_, err = p.service.ResourceRecordSets.Patch(p.Project, gcdZone, rrs.Name, rrs.Type, &rrs).Context(ctx).Do()
	return recordsToDelete, err
}
