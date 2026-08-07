// Package leaseweb implements a DNS record management client compatible
// with the libdns interfaces for Leaseweb.
// Upstream documentation found at:
// https://developer.leaseweb.com/api-docs/domains_v2.html
package leaseweb

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/libdns/libdns"
)

const (
	LeasewebApiKeyHeader = "X-LSW-Auth"
)

// Provider facilitates DNS record manipulation with Leaseweb.
type Provider struct {
	// Leasewebs API key. Generate one in the Leaseweb customer portal -> Administration -> API Key
	APIKey string `json:"api_token,omitempty"`
	mutex  sync.Mutex
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	domainName := strings.TrimSuffix(zone, ".")

	recordSets, err := p.listRecordSets(domainName)
	if err != nil {
		return nil, err
	}

	records := fromLeaseweb(recordSets)

	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	recordSets, err := fromLibdns(zone, records)
	if err != nil {
		return nil, err
	}

	for _, recordSet := range recordSets {
		_, err := p.createRecordSet(zone, recordSet)
		if err != nil {
			return nil, err
		}
	}

	// TODO: Ideally should check which records are actually POSTed.
	// For now we can assume all if we reach this point with no errors.
	var addedRecords = records
	return addedRecords, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	domainName := strings.TrimSuffix(zone, ".")
	existingRecordSets, err := p.listRecordSets(domainName)
	if err != nil {
		return nil, err
	}

	recordSets, err := fromLibdns(zone, records)
	if err != nil {
		return nil, err
	}

	existingRecords := fromLeaseweb(existingRecordSets)

	var updatedRecords []libdns.Record
	for _, recordSet := range recordSets {
		var hasExisting = false
		for _, existingRecord := range existingRecords {
			if existingRecord.Name == recordSet.Name && existingRecord.Type == recordSet.Type {
				hasExisting = true
			}
		}

		if hasExisting {
			updatedRecordResponse, err := p.updateRecordSet(zone, recordSet)
			if err != nil {
				return nil, err
			}

			for _, updatedRecord := range fromLeaseweb(updatedRecordResponse) {
				updatedRecords = append(updatedRecords, updatedRecord)
			}
		} else {
			_, err := p.createRecordSet(zone, recordSet)
			if err != nil {
				return nil, err
			}
		}
	}

	return updatedRecords, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
// Leaseweb specifics:
// - Well-formatted DELETE requests will always succeed, even for non-existing records.
// TODO: Extract HTTP logic into function in client.go.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	client := &http.Client{}

	var domainName = strings.TrimSuffix(zone, ".")

	recordSets, err := fromLibdns(zone, records)

	for _, recordSet := range recordSets {
		if err != nil {
			return nil, err
		}

		// https://developer.leaseweb.com/api-docs/domains_v2.html#operation/delete/domains/{domainName}/resourceRecordSets/{name}/{type}
		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("https://api.leaseweb.com/hosting/v2/domains/%s/resourceRecordSets/%s/%s", domainName, recordSet.Name, recordSet.Type), nil)
		if err != nil {
			return nil, err
		}

		req.Header.Add(LeasewebApiKeyHeader, p.APIKey)

		res, err := client.Do(req)
		defer res.Body.Close()
		if err != nil {
			return nil, err
		}
		err = handleLeasewebHttpError(res, "delete")
		if err != nil {
			return nil, err
		}
	}

	// TODO: Ideally should check which records are actually POSTed.
	// For now we can assume all if we reach this point with no errors.
	var deletedRecords = records
	return deletedRecords, nil
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
