package inwx

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/libdns/libdns"
)

// Provider facilitates DNS record manipulation with INWX.
type Provider struct {
	// Username of your INWX account.
	Username string `json:"username,omitempty"`

	// Password of your INWX account.
	Password string `json:"password,omitempty"`

	// The shared secret is used to generate a TAN if you have activated "Mobile TAN" for your INWX account.
	SharedSecret string `json:"shared_secret,omitempty"`

	// URL of the JSON-RPC API endpoint. It defaults to the production endpoint.
	EndpointURL string `json:"endpoint_url,omitempty"`

	client   *client
	clientMu sync.Mutex
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	client, err := p.getClient(ctx)
	defer p.removeClient(ctx)

	if err != nil {
		return nil, err
	}

	records, err := client.getRecords(ctx, getDomain(zone))

	if err != nil {
		return nil, err
	}

	results := make([]libdns.Record, 0, len(records))

	for _, inwxRecord := range records {
		results = append(results, libdnsRecord(inwxRecord, zone))
	}

	return results, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	client, err := p.getClient(ctx)
	defer p.removeClient(ctx)

	if err != nil {
		return nil, err
	}

	var results []libdns.Record

	for _, record := range records {
		var recordId, err = client.createRecord(ctx, inwxRecord(record), getDomain(zone))

		if err != nil {
			return nil, err
		}

		record.ID = strconv.Itoa(recordId)

		results = append(results, record)
	}

	return results, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	client, err := p.getClient(ctx)
	defer p.removeClient(ctx)

	if err != nil {
		return nil, err
	}

	var results []libdns.Record

	for _, record := range records {
		if record.ID == "" {
			matches, err := p.client.findRecords(ctx, inwxRecord(record), getDomain(zone), false)

			if err != nil {
				return nil, err
			}

			if len(matches) == 1 {
				record.ID = strconv.Itoa(matches[0].ID)
			}

			if len(matches) == 0 {
				recordId, err := client.createRecord(ctx, inwxRecord(record), getDomain(zone))

				if err != nil {
					return nil, err
				}

				record.ID = strconv.Itoa(recordId)

				results = append(results, record)

				continue
			}

			if len(matches) > 1 {
				return nil, fmt.Errorf("found more than one DNS record for %v", record)
			}
		}

		err := client.updateRecord(ctx, inwxRecord(record))

		if err != nil {
			return nil, err
		}

		results = append(results, record)

	}

	return results, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	client, err := p.getClient(ctx)
	defer p.removeClient(ctx)

	if err != nil {
		return nil, err
	}

	var results []libdns.Record

	for _, record := range records {
		delRecord := record

		if record.ID == "" {
			matches, err := p.client.findRecords(ctx, inwxRecord(record), getDomain(zone), true)

			if err != nil {
				return nil, err
			}

			if len(matches) == 1 {
				delRecord = libdnsRecord(matches[0], zone)
			}

			if len(matches) > 1 {
				return nil, fmt.Errorf("found more than one DNS record for %v", record)
			}
		}

		err := client.deleteRecord(ctx, inwxRecord(delRecord))

		if err != nil {
			return nil, err
		}

		results = append(results, delRecord)
	}

	return results, nil
}

func (p *Provider) getClient(ctx context.Context) (*client, error) {
	p.clientMu.Lock()
	defer p.clientMu.Unlock()

	if p.client == nil {
		client, err := newClient(p.getEndpointURL())
		p.client = client

		if err != nil {
			return nil, err
		}

		err = client.login(ctx, p.Username, p.Password, p.SharedSecret)

		if err != nil {
			return nil, err
		}
	}

	return p.client, nil
}

func (p *Provider) removeClient(ctx context.Context) {
	p.clientMu.Lock()
	defer p.clientMu.Unlock()

	if p.client == nil {
		return
	}

	p.client.logout(ctx)

	p.client = nil
}

func (p *Provider) getEndpointURL() string {
	if p.EndpointURL != "" {
		return p.EndpointURL
	}

	return endpointURL
}

func getDomain(zone string) string {
	return strings.TrimSuffix(zone, ".")
}

func libdnsRecord(record nameserverRecord, zone string) libdns.Record {
	return libdns.Record{
		ID:       strconv.Itoa(record.ID),
		Type:     record.Type,
		Name:     libdns.RelativeName(record.Name, getDomain(zone)),
		Value:    record.Content,
		TTL:      time.Duration(record.TTL) * time.Second,
		Priority: record.Priority,
	}
}

func inwxRecord(record libdns.Record) nameserverRecord {
	recordId, _ := strconv.Atoi(record.ID)

	return nameserverRecord{
		ID:       recordId,
		Name:     record.Name,
		Type:     record.Type,
		Content:  record.Value,
		TTL:      int(record.TTL.Seconds()),
		Priority: record.Priority,
	}
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
