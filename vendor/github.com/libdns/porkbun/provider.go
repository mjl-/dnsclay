// Package porkbun implements a DNS record management client compatible
// with the libdns interfaces for Porkbun.
package porkbun

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/libdns/libdns"
)

// Provider facilitates DNS record manipulation with Porkbun.
type Provider struct {
	APIKey       string `json:"api_key,omitempty"`
	APISecretKey string `json:"api_secret_key,omitempty"`
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(_ context.Context, zone string) ([]libdns.Record, error) {
	trimmedZone := LibdnsZoneToPorkbunDomain(zone)

	credentialJson, err := json.Marshal(p.getCredentials())
	if err != nil {
		return nil, err
	}
	response, err := MakeApiRequest("/dns/retrieve/"+trimmedZone, bytes.NewReader(credentialJson), pkbnRecordsResponse{})

	if err != nil {
		return nil, err
	}

	if response.Status != "SUCCESS" {
		return nil, errors.New(fmt.Sprintf("Invalid response status %s", response.Status))
	}

	recs := make([]libdns.Record, 0, len(response.Records))
	for _, rec := range response.Records {
		recs = append(recs, rec.toLibdnsRecord(zone))
	}
	return recs, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(_ context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	credentials := p.getCredentials()
	trimmedZone := LibdnsZoneToPorkbunDomain(zone)

	var createdRecords []libdns.Record

	for _, record := range records {
		if record.TTL/time.Second < 600 {
			record.TTL = 600 * time.Second
		}
		ttlInSeconds := int(record.TTL / time.Second)
		relativeName := libdns.RelativeName(record.Name, zone)
		trimmedName := relativeName
		if relativeName == "@" {
			trimmedName = ""
		}

		reqBody := pkbnRecordPayload{&credentials, record.Value, trimmedName, strconv.Itoa(ttlInSeconds), record.Type}
		reqJson, err := json.Marshal(reqBody)
		if err != nil {
			return createdRecords, err
		}

		response, err := MakeApiRequest(fmt.Sprintf("/dns/create/%s", trimmedZone), bytes.NewReader(reqJson), pkbnCreateResponse{})

		if err != nil {
			return createdRecords, err
		}

		if response.Status != "SUCCESS" {
			return createdRecords, errors.New(fmt.Sprintf("Invalid response status %s", response.Status))
		}

		// TODO contact support endpoint isn't returning the ID despite it being in their docs. Fetch as a workaround
		created, err := p.getMatchingRecord(record, zone)
		if err == nil && len(created) == 1 {
			record.ID = created[0].ID
		}
		createdRecords = append(createdRecords, record)
	}

	return createdRecords, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	var updates []libdns.Record
	var creates []libdns.Record
	var results []libdns.Record
	for _, r := range records {
		if r.ID == "" {
			// Try fetch record in case we are just missing the ID
			matches, err := p.getMatchingRecord(r, zone)
			if err != nil {
				return nil, err
			}

			if len(matches) == 0 {
				creates = append(creates, r)
				continue
			}

			if len(matches) > 1 {
				return nil, fmt.Errorf("unexpectedly found more than 1 record for %v", r)
			}

			r.ID = matches[0].ID
			updates = append(updates, r)
		} else {
			updates = append(updates, r)
		}
	}

	created, err := p.AppendRecords(ctx, zone, creates)
	if err != nil {
		return nil, err
	}
	updated, err := p.updateRecords(ctx, zone, updates)
	if err != nil {
		return nil, err
	}

	results = append(results, created...)
	results = append(results, updated...)
	return results, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(_ context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	credentials := p.getCredentials()
	trimmedZone := LibdnsZoneToPorkbunDomain(zone)

	var deletedRecords []libdns.Record

	for _, record := range records {
		var queuedDeletes []libdns.Record
		if record.ID == "" {
			// Try fetch record in case we are just missing the ID
			matches, err := p.getMatchingRecord(record, zone)
			if err != nil {
				return deletedRecords, err
			}
			for _, rec := range matches {
				queuedDeletes = append(queuedDeletes, rec)
			}
		} else {
			queuedDeletes = append(queuedDeletes, record)
		}

		reqJson, err := json.Marshal(credentials)
		if err != nil {
			return nil, err
		}

		for _, recordToDelete := range queuedDeletes {
			_, err = MakeApiRequest(fmt.Sprintf("/dns/delete/%s/%s", trimmedZone, recordToDelete.ID), bytes.NewReader(reqJson), pkbnResponseStatus{})
			if err != nil {
				return deletedRecords, err
			}
			deletedRecords = append(deletedRecords, recordToDelete)
		}
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
