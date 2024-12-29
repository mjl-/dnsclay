package dnsimple

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/libdns/libdns"
)

// Provider facilitates DNS record manipulation with DNSimple.
type Provider struct {
	APIAccessToken string `json:"api_access_token,omitempty"`
	AccountID      string `json:"account_id,omitempty"`
	APIURL         string `json:"api_url,omitempty"`

	client dnsimple.Client
	once   sync.Once
	mutex  sync.Mutex
}

// initClient will initialize the DNSimple API client with the provided access token and
// store the client in the Provider struct, along with setting the API URL and Account ID.
func (p *Provider) initClient(ctx context.Context) {
	p.once.Do(func() {
		// Create new DNSimple client using the provided access token.
		tc := dnsimple.StaticTokenHTTPClient(ctx, p.APIAccessToken)
		c := dnsimple.NewClient(tc)
		// Set the API URL if using a non-default API hostname (e.g. sandbox).
		if p.APIURL != "" {
			c.BaseURL = p.APIURL
		}
		// If no Account ID is provided, we can call the API to get the corresponding
		// account id for the provided access token.
		if p.AccountID == "" {
			resp, _ := c.Identity.Whoami(context.Background())
			accountID := strconv.FormatInt(resp.Data.Account.ID, 10)
			p.AccountID = accountID
		}

		p.client = *c
	})
}

// Internal helper function to fetch records from the provider, note that this function assumes
// the called is holding a lock on the mutex and has already initialized the client.
func (p *Provider) getRecordsFromProvider(ctx context.Context, zone string) ([]libdns.Record, error) {
	var records []libdns.Record

	resp, err := p.client.Zones.ListRecords(ctx, p.AccountID, unFQDN(zone), &dnsimple.ZoneRecordListOptions{})
	if err != nil {
		return nil, err
	}
	for _, r := range resp.Data {
		record := libdns.Record{
			ID:       strconv.FormatInt(r.ID, 10),
			Type:     r.Type,
			Name:     r.Name,
			Value:    r.Content,
			TTL:      time.Duration(r.TTL * int(time.Second)),
			Priority: uint(r.Priority),
		}
		records = append(records, record)
	}

	return records, nil
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.initClient(ctx)

	return p.getRecordsFromProvider(ctx, zone)
}

// Internal helper function that actually creates the records, does not hold a lock since the called is
// assumed to be holding a lock on the mutex and is in charge of making sure the client is initialized.
func (p *Provider) createRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	var createdRecords []libdns.Record

	// Get the Zone ID from zone name
	resp, err := p.client.Zones.GetZone(ctx, p.AccountID, unFQDN(zone))
	if err != nil {
		return nil, err
	}
	zoneID := strconv.FormatInt(resp.Data.ID, 10)

	for _, r := range records {
		attrs := dnsimple.ZoneRecordAttributes{
			ZoneID:   zoneID,
			Type:     r.Type,
			Name:     &r.Name,
			Content:  r.Value,
			TTL:      int(r.TTL.Seconds()),
			Priority: int(r.Priority),
		}
		resp, err := p.client.Zones.CreateRecord(ctx, p.AccountID, unFQDN(zone), attrs)
		if err != nil {
			return nil, err
		}
		// See https://developer.dnsimple.com/v2/zones/records/#createZoneRecord
		if resp.HTTPResponse.StatusCode == http.StatusCreated {
			r.ID = strconv.FormatInt(resp.Data.ID, 10)
			createdRecords = append(createdRecords, r)
		} else {
			return nil, fmt.Errorf("error creating record: %s, error: %s", r.Name, resp.HTTPResponse.Status)
		}
	}
	return createdRecords, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.initClient(ctx)

	return p.createRecords(ctx, zone, records)
}

// Internal helper function to get the lists of records to create and update respectively
func (p *Provider) getRecordsToCreateAndUpdate(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, []libdns.Record, error) {
	existingRecords, err := p.getRecordsFromProvider(ctx, zone)
	if err != nil {
		return nil, nil, err
	}
	var recordsToUpdate []libdns.Record

	updateMap := make(map[libdns.Record]bool)
	var recordsToCreate []libdns.Record

	// Figure out which records exist and need to be updated
	for _, r := range records {
		updateMap[r] = true
		for _, er := range existingRecords {
			if r.Name != er.Name {
				continue
			}
			if r.ID == "0" || r.ID == "" {
				r.ID = er.ID
			}
			recordsToUpdate = append(recordsToUpdate, r)
			updateMap[r] = false
		}
	}
	// If the record is not updating an existing record, we want to create it
	for r, updating := range updateMap {
		if updating {
			recordsToCreate = append(recordsToCreate, r)
		}
	}

	return recordsToCreate, recordsToUpdate, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.initClient(ctx)

	var setRecords []libdns.Record

	recordsToCreate, recordsToUpdate, err := p.getRecordsToCreateAndUpdate(ctx, zone, records)
	if err != nil {
		return nil, err
	}

	// Create new records and append them to 'setRecords'
	createdRecords, err := p.createRecords(ctx, zone, recordsToCreate)
	if err != nil {
		return nil, err
	}
	for _, r := range createdRecords {
		setRecords = append(setRecords, r)
	}

	// Get the Zone ID from zone name
	resp, err := p.client.Zones.GetZone(ctx, p.AccountID, unFQDN(zone))
	if err != nil {
		return nil, err
	}
	zoneID := strconv.FormatInt(resp.Data.ID, 10)

	// Update existing records and append them to 'SetRecords'
	for _, r := range recordsToUpdate {
		attrs := dnsimple.ZoneRecordAttributes{
			ZoneID:   zoneID,
			Type:     r.Type,
			Name:     &r.Name,
			Content:  r.Value,
			TTL:      int(r.TTL.Seconds()),
			Priority: int(r.Priority),
		}
		id, err := strconv.ParseInt(r.ID, 10, 64)
		if err != nil {
			return nil, err
		}
		resp, err := p.client.Zones.UpdateRecord(ctx, p.AccountID, unFQDN(zone), id, attrs)
		if err != nil {
			return nil, err
		}
		// https://developer.dnsimple.com/v2/zones/records/#updateZoneRecord
		if resp.HTTPResponse.StatusCode == http.StatusOK {
			r.ID = strconv.FormatInt(resp.Data.ID, 10)
			setRecords = append(setRecords, r)
		} else {
			return nil, fmt.Errorf("error updating record: %s", resp.HTTPResponse.Status)
		}
	}
	return setRecords, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.initClient(ctx)

	var deleted []libdns.Record
	var noID []libdns.Record

	for _, r := range records {
		// If the record does not have an ID, we'll try to find it by calling the API later
		// and extrapolating its ID based on the record name, but continue for now.
		if r.ID == "0" || r.ID == "" {
			noID = append(noID, r)
			continue
		}

		id, err := strconv.ParseInt(r.ID, 10, 64)
		if err != nil {
			return deleted, err
		}

		resp, err := p.client.Zones.DeleteRecord(ctx, p.AccountID, unFQDN(zone), id)
		if err != nil {
			return deleted, err
		}
		// See https://developer.dnsimple.com/v2/zones/records/#deleteZoneRecord for API response codes
		if resp.HTTPResponse.StatusCode == http.StatusNoContent {
			deleted = append(deleted, r)
		} else {
			return nil, fmt.Errorf("error deleting record: %s, error: %s", r.Name, resp.HTTPResponse.Status)
		}
	}

	// Return early if there are no records we need to try and find IDs for
	if len(noID) == 0 {
		return deleted, nil
	}

	// If we received records without an ID earlier, we're going to try and figure out the ID by calling
	// GetRecords and comparing the record name. If we're able to find it, we'll delete it, otherwise
	// we'll append it to our list of failed to delete records.
	existingRecords, err := p.getRecordsFromProvider(ctx, zone)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing records: %s", err.Error())
	}
	for _, r := range noID {
		for _, fr := range existingRecords {
			if r.Name != fr.Name {
				continue
			}
			id, err := strconv.ParseInt(fr.ID, 10, 64)
			if err != nil {
				return nil, err
			}
			resp, err := p.client.Zones.DeleteRecord(ctx, p.AccountID, unFQDN(zone), id)
			if err != nil {
				return nil, err
			}
			// See https://developer.dnsimple.com/v2/zones/records/#deleteZoneRecord for API response codes
			if resp.HTTPResponse.StatusCode == http.StatusNoContent {
				deleted = append(deleted, r)
			} else {
				return nil, fmt.Errorf("error deleting record: %s, error: %s", r.Name, resp.HTTPResponse.Status)
			}
			break
		}
	}
	return deleted, nil
}

// unFQDN trims any trailing "." from fqdn. dnsimple's API does not use FQDNs.
func unFQDN(fqdn string) string {
	return strings.TrimSuffix(fqdn, ".")
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
