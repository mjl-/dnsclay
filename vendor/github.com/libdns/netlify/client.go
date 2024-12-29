package netlify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/libdns/libdns"
)

// createRecord creates a DNS record in the specified zone. It returns the DNS
// record created
func (p *Provider) createRecord(ctx context.Context, zoneInfo netlifyZone, record libdns.Record) (netlifyDNSRecord, error) {
	jsonBytes, err := json.Marshal(netlifyRecord(record))
	if err != nil {
		return netlifyDNSRecord{}, err
	}
	reqURL := fmt.Sprintf("%s/dns_zones/%s/dns_records", baseURL, zoneInfo.ID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return netlifyDNSRecord{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	var result netlifyDNSRecord
	err = p.doAPIRequest(req, false, false, false, true, &result)
	if err != nil {
		return netlifyDNSRecord{}, err
	}

	return result, nil
}

// updateRecord updates a DNS record. oldRec must have both an ID and zone ID.
// Only the non-empty fields in newRec will be changed.
func (p *Provider) updateRecord(ctx context.Context, oldRec netlifyDNSRecord, newRec netlifyDNSRecord) (netlifyDNSRecord, error) {
	zoneID := oldRec.DNSZoneID
	// Temporary fix as currently the only way to update a dns record is to delete the previous one and to recreate one with the new specifications
	reqURL := fmt.Sprintf("%s/dns_zones/%s/dns_records/%s", baseURL, zoneID, oldRec.ID)
	jsonBytes, err := json.Marshal(newRec)
	if err != nil {
		return netlifyDNSRecord{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return netlifyDNSRecord{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	var result netlifyDNSRecord
	err = p.doAPIRequest(req, false, false, false, true, &result)
	if err != nil {
		return netlifyDNSRecord{}, err
	}

	jsonEditBytes, err := json.Marshal(newRec)
	if err != nil {
		return netlifyDNSRecord{}, err
	}
	reqEditURL := fmt.Sprintf("%s/dns_zones/%s/dns_records", baseURL, zoneID)
	reqEdit, err := http.NewRequestWithContext(ctx, http.MethodPost, reqEditURL, bytes.NewReader(jsonEditBytes))
	if err != nil {
		return netlifyDNSRecord{}, err
	}
	reqEdit.Header.Set("Content-Type", "application/json")

	var resultEdit netlifyDNSRecord
	err = p.doAPIRequest(reqEdit, false, false, false, true, &resultEdit)
	if err != nil {
		return netlifyDNSRecord{}, err
	}

	return resultEdit, err
}

// getDNSRecords gets all record in a zone. It returns an array of the records
// in the zone. It may return an empty slice and a nil error.
func (p *Provider) getDNSRecords(ctx context.Context, zoneInfo netlifyZone, rec libdns.Record, matchContent bool) ([]netlifyDNSRecord, error) {
	qs := make(url.Values)
	qs.Set("type", rec.Type)
	qs.Set("name", libdns.AbsoluteName(rec.Name, zoneInfo.Name))
	if matchContent {
		qs.Set("content", rec.Value)
	}

	reqURL := fmt.Sprintf("%s/dns_zones/%s/dns_records", baseURL, zoneInfo.ID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	var results []netlifyDNSRecord
	err = p.doAPIRequest(req, false, false, true, false, &results)
	var rest_to_return []netlifyDNSRecord
	for _, res := range results {
		if res.Hostname == libdns.AbsoluteName(rec.Name, zoneInfo.Name) && res.Type == rec.Type {
			rest_to_return = append(rest_to_return, res)
		}
	}
	if err != nil {
		return nil, err
	}
	return rest_to_return, nil
}

// getZoneInfo get the information from a DNS zone. It returns the dns zone
func (p *Provider) getZoneInfo(ctx context.Context, zoneName string) (netlifyZone, error) {
	p.zonesMu.Lock()
	defer p.zonesMu.Unlock()

	zoneName = strings.TrimRight(zoneName, ".")

	// if we already got the zone info, reuse it
	if p.zones == nil {
		p.zones = make(map[string]netlifyZone)
	}
	if zone, ok := p.zones[zoneName]; ok {
		return zone, nil
	}

	qs := make(url.Values)
	qs.Set("name", zoneName)
	reqURL := fmt.Sprintf("%s/dns_zones?%s", baseURL, qs.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return netlifyZone{}, err
	}

	var zones []netlifyZone
	err = p.doAPIRequest(req, true, false, true, false, &zones)
	if err != nil {
		return netlifyZone{}, err
	}
	if len(zones) != 1 {
		return netlifyZone{}, fmt.Errorf("expected 1 zone, got %d for %s", len(zones), zoneName)
	}

	// cache this zone for possible reuse
	p.zones[zoneName] = zones[0]

	return zones[0], nil
}

// doAPIRequest authenticates the request req and does the round trip. It returns
// nil if there was no error, the error otherwise. The decoded content is passed
// to the calling function by the result variable
func (p *Provider) doAPIRequest(req *http.Request, isZone bool, isDel bool, isGet bool, isSolo bool, result interface{}) error {
	req.Header.Set("Authorization", "Bearer "+p.PersonalAccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("got error status: HTTP %d: %+v", resp.StatusCode, string(body))
	}

	// delete DNS record
	if isDel && !isZone {
		if len(body) > 0 {
			var err netlifyDNSDeleteError
			json.Unmarshal(body, &result)
			return fmt.Errorf(err.Message)
		}
		return err
	}

	// get zone info
	if isZone && isGet {
		err = json.Unmarshal(body, &result)
		if err != nil {
			return err
		}
		return err
	}

	// get DNS records
	if !isZone && isGet && !isSolo {
		err = json.Unmarshal(body, &result)
		if err != nil {
			return err
		}
		return err
	}

	// get DNS record
	if !isZone && isGet && isSolo {
		err = json.Unmarshal(body, &result)
		if err != nil {
			return err
		}
		return err
	}

	// update DNS record
	if !isZone && isSolo && !isGet {
		err = json.Unmarshal(body, &result)
		if err != nil {
			return nil
		}
		return err
	}

	return err
}

const baseURL = "https://api.netlify.com/api/v1"
