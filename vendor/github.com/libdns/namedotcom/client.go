package namedotcom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/libdns/libdns"
)

// nameClient extends the namedotcom api and request handler to the provider..
type nameClient struct {
	client *nameDotCom
	mutex  sync.Mutex
}

// getClient initiates a new nameClient and assigns it to the provider..
func (p *Provider) getClient(ctx context.Context) error {
	newNameclient, err := NewNameDotComClient(ctx, p.Token, p.User, p.Server)
	if err != nil {
		return err
	}
	p.client = newNameclient
	return nil
}

// listAllRecords returns all records for the given zone .. GET /v4/domains/{ domainName }/records
func (p *Provider) listAllRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	var (
		records []libdns.Record

		/*** 'zone' args that are passed in using compliant zone formats have the FQDN '.' suffix qualifier
		and in order to use the zone arg as a domainName reference to name.com's api we must remove the '.' suffix.
		otherwise the api will not recognize the domain.. ***/
		unFQDNzone = strings.TrimSuffix(zone, ".")

		method  = "GET"
		body    io.Reader
		resp    = &listRecordsResponse{}
		reqPage = 1

		err error
	)

	if err = p.getClient(ctx); err != nil {
		return []libdns.Record{}, err
	}

	// handle pagination, in case domain has more records then the default of 1000 per page
	for reqPage > 0 {
		if reqPage != 0 {
			endpoint := fmt.Sprintf("/v4/domains/%s/records?page=%d", unFQDNzone, reqPage)

			if body, err = p.client.doRequest(ctx, method, endpoint, nil); err != nil {
				return []libdns.Record{}, fmt.Errorf("request failed:  %w", err)
			}

			if err = json.NewDecoder(body).Decode(resp); err != nil {
				return []libdns.Record{}, fmt.Errorf("could not decode name.com's response:  %w", err)
			}

			for _, record := range resp.Records {
				records = append(records, record.toLibDNSRecord())
			}

			reqPage = int(resp.NextPage)
		}
	}

	return records, nil
}

//  deleteRecord  DELETE /v4/domains/{ domainName }/records/{ record.ID }
func (p *Provider) deleteRecord(ctx context.Context, zone string, record libdns.Record) (libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	var (
		shouldDelete nameDotComRecord
		unFQDNzone   = strings.TrimSuffix(zone, ".")

		method   = "DELETE"
		endpoint = fmt.Sprintf("/v4/domains/%s/records/%s", unFQDNzone, record.ID)
		body     io.Reader
		post     = &bytes.Buffer{}

		err error
	)

	shouldDelete.fromLibDNSRecord(record, unFQDNzone)

	if err = p.getClient(ctx); err != nil {
		return libdns.Record{}, err
	}

	if err = json.NewEncoder(post).Encode(shouldDelete); err != nil {
		return libdns.Record{}, fmt.Errorf("could not encode form data for request:  %w", err)
	}

	if body, err = p.client.doRequest(ctx, method, endpoint, post); err != nil {
		return libdns.Record{}, fmt.Errorf("request to delete the record was not successful:  %w", err)
	}

	if err = json.NewDecoder(body).Decode(&shouldDelete); err != nil {
		return libdns.Record{}, fmt.Errorf("could not decode the response from name.com:  %w", err)
	}

	return shouldDelete.toLibDNSRecord(), nil
}

// upsertRecord  PUT || POST /v4/domains/{ domainName }/records/{ record.ID }
func (p *Provider) upsertRecord(ctx context.Context, zone string, canidateRecord libdns.Record) (libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var (
		shouldUpsert nameDotComRecord
		unFQDNzone   = strings.TrimSuffix(zone, ".")

		method   = "PUT"
		endpoint = fmt.Sprintf("/v4/domains/%s/records/%s", unFQDNzone, canidateRecord.ID)
		body     io.Reader
		post     = &bytes.Buffer{}

		err error
	)

	if canidateRecord.ID == "" {
		method = "POST"
		endpoint = fmt.Sprintf("/v4/domains/%s/records", unFQDNzone)
	}

	shouldUpsert.fromLibDNSRecord(canidateRecord, unFQDNzone)

	if err = p.getClient(ctx); err != nil {
		return libdns.Record{}, err
	}

	if err = json.NewEncoder(post).Encode(shouldUpsert); err != nil {
		return libdns.Record{}, fmt.Errorf("could not encode the form data for the request:  %w", err)
	}

	if body, err = p.client.doRequest(ctx, method, endpoint, post); err != nil {
		if strings.Contains(err.Error(), "Duplicate Record") {
			err = fmt.Errorf("name.com will not allow an update to a record that has identical values to an existing record: %w", err)
		}

		return libdns.Record{}, fmt.Errorf("request to update the record was not successful:  %w", err)
	}

	if err = json.NewDecoder(body).Decode(&shouldUpsert); err != nil {
		return libdns.Record{}, fmt.Errorf("could not decode name.com's response:  %w", err)
	}

	return shouldUpsert.toLibDNSRecord(), nil
}
