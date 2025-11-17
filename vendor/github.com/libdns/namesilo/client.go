package namesilo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

func (p *Provider) defaultValues() url.Values {
	qs := make(url.Values)
	qs.Set("version", "1")
	qs.Set("type", "json")
	qs.Set("key", p.APIToken)
	return qs
}

func (p *Provider) getDNSRecords(ctx context.Context, zone string) ([]record, error) {
	qs := p.defaultValues()
	qs.Set("domain", zoneToDomain(zone))

	reqURL := fmt.Sprintf("%s/dnsListRecords?%s", baseURL, qs.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	var results []record
	_, err = p.doAPIRequest(req, &results)
	return results, err
}

func (p *Provider) deleteDNSRecord(ctx context.Context, zone string, recordId string) error {
	qs := p.defaultValues()
	qs.Set("domain", zoneToDomain(zone))
	qs.Set("rrid", recordId)

	reqURL := fmt.Sprintf("%s/dnsDeleteRecord?%s", baseURL, qs.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}

	_, err = p.doAPIRequest(req, nil)
	return err
}

func (p *Provider) addDNSRecord(ctx context.Context, zone string, record record) error {
	qs := p.defaultValues()
	qs.Set("domain", zoneToDomain(zone))
	qs.Set("rrtype", record.Type)
	qs.Set("rrhost", record.Host)
	qs.Set("rrvalue", record.Value)
	if record.Distance != 0 {
		qs.Set("rrdistance", strconv.Itoa(int(record.Distance)))
	}
	if record.TTL != 0 {
		qs.Set("rrttl", strconv.Itoa(int(record.TTL)))
	}

	reqURL := fmt.Sprintf("%s/dnsAddRecord?%s", baseURL, qs.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}

	_, err = p.doAPIRequest(req, nil)
	return err
}

func (p *Provider) updateDNSRecord(ctx context.Context, zone string, recordId string, record record) error {
	qs := p.defaultValues()
	qs.Set("domain", zoneToDomain(zone))
	qs.Set("rrid", recordId)
	qs.Set("rrtype", record.Type)
	qs.Set("rrhost", record.Host)
	qs.Set("rrvalue", record.Value)
	if record.Distance != 0 {
		qs.Set("rrdistance", strconv.Itoa(int(record.Distance)))
	}
	if record.TTL != 0 {
		qs.Set("rrttl", strconv.Itoa(int(record.TTL)))
	}

	reqURL := fmt.Sprintf("%s/dnsUpdateRecord?%s", baseURL, qs.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}

	_, err = p.doAPIRequest(req, nil)
	return err
}

func (p *Provider) doAPIRequest(req *http.Request, result any) (namesiloResponse, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return namesiloResponse{}, err
	}
	defer resp.Body.Close()

	var respData namesiloResponse
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return namesiloResponse{}, err
	}

	if resp.StatusCode >= 400 {
		return namesiloResponse{}, fmt.Errorf("got error status: HTTP %d: %s", resp.StatusCode, respData.Reply.Detail)
	}

	if respData.Reply.Code != 300 {
		return namesiloResponse{}, fmt.Errorf("failed to perform API call: Namesilo API status code %d; %s", respData.Reply.Code, respData.Reply.Detail)
	}

	if len(respData.Reply.ResourceRecord) > 0 && result != nil {
		if err = json.Unmarshal(respData.Reply.ResourceRecord, result); err != nil {
			return respData, err
		}
	}

	return respData, nil
}

const baseURL = "https://www.namesilo.com/api"
