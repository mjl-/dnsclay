// Package googleclouddns implements a DNS record management client compatible
// with the libdns interfaces for Google Cloud DNS.
package googleclouddns

import (
	"context"
	"sync"
	"time"

	"github.com/libdns/libdns"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/googleapi"
)

// Provider facilitates DNS record manipulation with Google Cloud DNS.
type Provider struct {
	Project            string `json:"gcp_project,omitempty"`
	ServiceAccountJSON string `json:"gcp_application_default,omitempty"`
	service            *dns.Service
	zoneMap            map[string]string
	zoneMapLastUpdated time.Time
	mutex              sync.Mutex
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.getCloudDNSRecords(ctx, zone)
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	processedRecords := make(libdnsRecords, 0)
	recordsToAppend := libdnsRecords(records)
	for recordData, recordsToPost := range recordsToAppend.groupRecordsByType() {
		existingRecords, err := p.getCloudDNSRecord(ctx, zone, recordData.name, recordData.recordType)
		if err != nil {
			if gErr, ok := err.(*googleapi.Error); !ok || gErr.Code != 404 {
				return processedRecords, err
			}
		}
		verifiedNewRecords := make(libdnsRecords, 0)
		for _, newRecord := range recordsToPost { // Make sure that we do not append a record that already exists
			if existingRecords.doesNotHaveRecord(newRecord) {
				verifiedNewRecords = append(verifiedNewRecords, newRecord)
			}
		}
		if len(verifiedNewRecords) == 0 {
			continue
		}
		submittedRecords, err := p.postCloudDNSRecord(ctx, zone, append(existingRecords, verifiedNewRecords...))
		if err != nil {
			return processedRecords, err
		}
		// Let's generate an exact list of appended records based on the returned results
		for _, updatedRecord := range submittedRecords {
			if verifiedNewRecords.hasRecord(updatedRecord) {
				processedRecords = append(processedRecords, updatedRecord)
			}
		}
	}
	return processedRecords, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	processedRecords := make(libdnsRecords, 0)
	recordsToSet := libdnsRecords(records)
	for _, recordsToPost := range recordsToSet.groupRecordsByType() {
		submittedRecords, err := p.postCloudDNSRecord(ctx, zone, recordsToPost)
		if err != nil {
			return processedRecords, err
		}
		processedRecords = append(processedRecords, submittedRecords...)
	}
	return processedRecords, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	recordsToDelete := libdnsRecords(records)
	deletedRecords := make(libdnsRecords, 0)
	for recordData, recordsToDelete := range recordsToDelete.groupRecordsByType() {
		existingRecords, err := p.getCloudDNSRecord(ctx, zone, recordData.name, recordData.recordType)
		if err != nil { // If the entry does not exist, nothing to delete so skip this set
			if gErr, ok := err.(*googleapi.Error); !ok || gErr.Code != 404 {
				return deletedRecords, err
			}
			continue
		}
		verifiedRecords := make(libdnsRecords, 0)
		for _, recordToDelete := range recordsToDelete { // Make sure the requested records exist in the Cloud DNS record
			if existingRecords.hasRecord(recordToDelete) {
				verifiedRecords = append(verifiedRecords, recordToDelete)
			}
		}
		if len(verifiedRecords) == 0 { // The Cloud DNS entry does not have these records so skip this set
			continue
		}
		processedRecords, err := p.deleteCloudDNSRecord(
			ctx, zone, recordData.name, recordData.recordType, verifiedRecords, existingRecords)
		if err != nil {
			return deletedRecords, err
		}
		deletedRecords = append(deletedRecords, processedRecords...)
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
