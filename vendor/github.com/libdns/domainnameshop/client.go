package domainnameshop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL string = "https://api.domeneshop.no/v0"

// We set a default ttl that's used if TTL is not specified by other users
// By default domainname.shop uses 1 hour long TTL which might be too long in a lot of usecases
// The api specifies that TTL must be in seconds but also in must multiples of 60
const defaultTtl = time.Duration(2 * time.Minute)

func (p *Provider) doRequest(token string, secret string, request *http.Request, result any) error {
	request.SetBasicAuth(token, secret)
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("got error status: HTTP %d: %+v", response.StatusCode, string(body))
	}

	if result != nil {
		if err = json.NewDecoder(response.Body).Decode(&result); err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) getDomainInfo(ctx context.Context, token string, secret string, zone string) (dsZone, error) {
	p.zonesMu.Lock()
	defer p.zonesMu.Unlock()
	// if we already got the zone info, reuse it
	if p.zones == nil {
		p.zones = make(map[string]dsZone)
	}
	if zone, ok := p.zones[zone]; ok {
		return zone, nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf(defaultBaseURL+"/domains?domain=%s", url.QueryEscape(removeFQDNTrailingDot(zone))), nil)
	if err != nil {
		return dsZone{}, err
	}

	var zones []dsZone
	err = p.doRequest(token, secret, req, &zones)
	if err != nil {
		return dsZone{}, err
	}

	if len(zones) != 1 {
		return dsZone{}, fmt.Errorf("expected 1 zone, got %d for %s", len(zones), zone)
	}
	p.zones[zone] = zones[0]

	return zones[0], nil
}

func (p *Provider) getAllDomainRecords(ctx context.Context, token string, secret string, zone string) ([]dsDNSRecord, error) {
	domain, err := p.getDomainInfo(ctx, token, secret, zone)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf(defaultBaseURL+"/domains/%d/dns", domain.ID), nil)
	if err != nil {
		return nil, err
	}

	var result []dsDNSRecord
	err = p.doRequest(token, secret, req, &result)
	if err != nil {
		return nil, err
	}

	// Save the records for later
	p.knownRecordsMu.Lock()
	defer p.knownRecordsMu.Unlock()
	if p.knownRecords == nil {
		p.knownRecords = make(map[string][]dsDNSRecord)
	}
	p.knownRecords[zone] = result

	return result, nil
}

// Get a dns record from zone
// Retrieving records directly require an ID, since we dont' really have that ahead of time we can only really rely on getting the whole zone
// We try to cache results to reduce the need for queries
func (p *Provider) getDNSRecord(ctx context.Context, token string, secret string, zone string, record dsDNSRecord) (dsDNSRecord, error) {
	// Try to retrieve from our cached records first
	var dsrecord = p.getRecordFromKnownRecords(record, zone)

	// if it's not an emtpy struct we return it
	if (dsDNSRecord{}) != dsrecord {
		return dsrecord, nil
	}

	// Fall back to getting the full zone info
	_, err := p.getAllDomainRecords(ctx, token, secret, zone)
	if err != nil {
		return dsDNSRecord{}, err
	}

	// Try to retrieve again, if it's still empty then we assume nothing was found
	dsrecord = p.getRecordFromKnownRecords(record, zone)
	return dsrecord, nil
}

func (p *Provider) deleteDNSRecord(ctx context.Context, token string, secret string, zone string, record dsDNSRecord) error {
	domain, err := p.getDomainInfo(ctx, token, secret, zone)
	if err != nil {
		return err
	}

	// Try to retrieve from our cached records first
	dsrecord, err := p.getDNSRecord(ctx, token, secret, zone, record)
	if err != nil {
		return err
	}
	// if the result is empty we don't need to delete
	if (dsDNSRecord{}) == dsrecord {
		return nil
	}

	reqURL := fmt.Sprintf(defaultBaseURL+"/domains/%d/dns/%d", domain.ID, dsrecord.ID)
	req, err := http.NewRequestWithContext(ctx, "DELETE", reqURL, nil)
	if err != nil {
		return err
	}

	err = p.doRequest(token, secret, req, nil)
	if err != nil {
		return err
	}
	_ = p.removeRecordFromKnownRecords(dsrecord, zone)
	return nil
}

func (p *Provider) createDNSRecord(ctx context.Context, token string, secret string, zone string, record dsDNSRecord) (dsDNSRecord, error) {
	domain, err := p.getDomainInfo(ctx, token, secret, zone)
	if err != nil {
		return dsDNSRecord{}, err
	}

	record.Host = normalizeRecordName(record.Host, zone)

	reqData := record
	if reqData.TTL == 0 {
		reqData.TTL = int(defaultTtl.Seconds())
	}
	reqBuffer, err := json.Marshal(reqData)
	if err != nil {
		return dsDNSRecord{}, err
	}

	reqURL := fmt.Sprintf(defaultBaseURL+"/domains/%d/dns", domain.ID)
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(reqBuffer))
	if err != nil {
		return dsDNSRecord{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	var result dsDNSRecord
	err = p.doRequest(token, secret, req, &result)
	if err != nil {
		return dsDNSRecord{}, err
	}
	// Add the ID to the incoming record
	record.ID = result.ID

	return record, nil
}

func (p *Provider) updateDNSRecord(ctx context.Context, token string, secret string, zone string, record dsDNSRecord) (dsDNSRecord, error) {
	domain, err := p.getDomainInfo(ctx, token, secret, zone)
	if err != nil {
		return dsDNSRecord{}, err
	}

	reqData := record
	if reqData.TTL == 0 {
		reqData.TTL = int(defaultTtl.Seconds())
	}
	reqBuffer, err := json.Marshal(reqData)
	if err != nil {
		return dsDNSRecord{}, err
	}

	reqURL := fmt.Sprintf(defaultBaseURL+"/domains/%d/dns/%d", domain.ID, record.ID)
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(reqBuffer))
	if err != nil {
		return dsDNSRecord{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	var result dsDNSRecord
	err = p.doRequest(token, secret, req, &result)
	if err != nil {
		return dsDNSRecord{}, err
	}
	// We don't actually get an result from the API so we query for update
	return p.getDNSRecord(ctx, token, secret, zone, record)
}

func (p *Provider) createOrUpdateDNSRecord(ctx context.Context, token string, secret string, zone string, r dsDNSRecord) (dsDNSRecord, error) {
	if r.ID == 0 {
		return p.createDNSRecord(ctx, token, secret, zone, r)
	}

	return p.updateDNSRecord(ctx, token, secret, zone, r)
}

func (p *Provider) getRecordFromKnownRecords(record dsDNSRecord, zone string) dsDNSRecord {
	p.knownRecordsMu.Lock()
	defer p.knownRecordsMu.Unlock()
	if p.knownRecords == nil {
		p.knownRecords = make(map[string][]dsDNSRecord)
	}

	if zoneRecords, ok := p.knownRecords[zone]; ok {
		for _, rec := range zoneRecords {
			if record.ID == rec.ID && record.ID != 0 && rec.ID != 0 {
				return rec
			} else if record.Host == rec.Host && record.Data == rec.Data && rec.ID != 0 {
				return rec
			}
		}
	}
	return dsDNSRecord{}
}

func (p *Provider) removeRecordFromKnownRecords(record dsDNSRecord, zone string) bool {
	p.knownRecordsMu.Lock()
	defer p.knownRecordsMu.Unlock()
	if p.knownRecords == nil {
		p.knownRecords = make(map[string][]dsDNSRecord)
	}

	if zoneRecords, ok := p.knownRecords[zone]; ok {
		for i, rec := range zoneRecords {
			if record.ID == rec.ID && record.ID != 0 {
				p.knownRecords[zone][i].ID = 0
				return true
			}
		}
	}
	return false
}

// func (p *Provider) updateRecordInKnownRecords(dsZoneRecords dsZoneRecord, dsRecord dsDNSRecord, zone string) {
// 	p.knownRecordsMu.Lock()
// 	defer p.knownRecordsMu.Unlock()
// 	if p.knownRecords == nil {
// 		p.knownRecords = make(map[string][]dsDNSRecord)
// 	}
// 	if zoneRecords, ok := p.knownRecords[zone]; ok {
// 		for _, rec := range zoneRecords {
// 			if record.ID == rec.ID {
// 				return rec
// 			} else if record.Host == rec.Host && record.Data == record.Data {
// 				return rec
// 			}
// 		}
// 	}
// 	p.knownRecords[dsZoneRecords] = dsRecord
// }

func removeFQDNTrailingDot(fqdn string) string {
	return strings.TrimSuffix(fqdn, ".")
}

func normalizeRecordName(recordName string, zone string) string {
	normalized := removeFQDNTrailingDot(recordName)
	normalized = strings.TrimSuffix(normalized, removeFQDNTrailingDot(zone))
	if normalized == "" {
		normalized = "@"
	}
	return removeFQDNTrailingDot(normalized)
}
