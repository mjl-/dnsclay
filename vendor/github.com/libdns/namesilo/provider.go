// Package libdnsnamesilo implements a DNS record management client compatible
// with the libdns interfaces for Namesilo.
package namesilo

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

const (
	apiEndpoint = "https://www.namesilo.com/api/"
)

// Provider facilitates DNS record manipulation with Namesilo.
type Provider struct {
	APIToken string `json:"api_token,omitempty"`
}

type reply struct {
	Code   int    `xml:"reply>code"`
	Detail string `xml:"reply>detail"`
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiEndpoint+"dnsListRecords?version=1&type=xml&key="+p.APIToken+"&domain="+zoneToDomain(zone), nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	var response struct {
		reply
		Records []struct {
			ID       string `xml:"record_id"`
			Type     string `xml:"type"`
			Name     string `xml:"host"`
			Value    string `xml:"value"`
			TTL      int    `xml:"ttl"`
			Priority int    `xml:"distance"`
		} `xml:"reply>resource_record"`
	}

	if err := doHttpRequestWithXmlResponse(client, req, &response); err != nil {
		return nil, fmt.Errorf("request failed: %s", err)
	}

	if response.Code != 300 {
		return nil, fmt.Errorf("failed to get records for zone \"%s\": Namesilo API status code %v; %s", zone, response.Code, response.Detail)
	}

	var records []libdns.Record

	for _, record := range response.Records {
		records = append(records, libdns.Record{
			ID:       record.ID,
			Type:     record.Type,
			Name:     record.Name,
			Value:    record.Value,
			TTL:      time.Duration(record.TTL) * time.Second,
			Priority: uint(record.Priority),
		})
	}

	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	client := &http.Client{}

	var appendedRecords []libdns.Record

	for _, record := range records {
		rrttl := ""
		if record.TTL != time.Duration(0) {
			rrttl = fmt.Sprintf("&rrttl=%d", int64(record.TTL/time.Second))
		}

		rrdistance := ""
		if record.Priority != 0 {
			rrdistance = fmt.Sprintf("&rrdistance=%d", record.Priority)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiEndpoint+"dnsAddRecord?version=1&type=xml&key="+p.APIToken+"&domain="+zoneToDomain(zone)+"&rrtype="+record.Type+"&rrhost="+record.Name+"&rrvalue="+record.Value+rrttl+rrdistance, nil)
		if err != nil {
			return appendedRecords, err
		}

		var response struct {
			reply
			ID string `xml:"reply>record_id"`
		}

		if err := doHttpRequestWithXmlResponse(client, req, &response); err != nil {
			return appendedRecords, fmt.Errorf("request failed: %s", err)
		}

		if response.Code != 300 {
			return appendedRecords, fmt.Errorf("failed to append record for zone \"%s\": Namesilo API status code %v; %s", zone, response.Code, response.Detail)
		}

		record.ID = response.ID

		appendedRecords = append(appendedRecords, record)
	}

	return appendedRecords, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	existingRecords, err := p.GetRecords(ctx, zone)
	if err != nil {
		return nil, fmt.Errorf("error to retrieve existing records: %v", err)
	}

	existingRecordsMap := make(map[string]libdns.Record)
	for _, record := range existingRecords {
		existingRecordsMap[record.ID] = record
	}

	var newRecords []libdns.Record
	var changedRecords []libdns.Record
	for _, record := range records {
		if record.ID == "" {
			newRecords = append(newRecords, record)
		} else {
			if _, exists := existingRecordsMap[record.ID]; !exists {
				return nil, fmt.Errorf("record does not exits %v", record)
			}
			changedRecords = append(changedRecords, record)
		}
	}

	resultRecords, err := p.AppendRecords(ctx, zone, newRecords)
	if err != nil {
		return resultRecords, err
	}

	client := &http.Client{}
	for _, record := range changedRecords {
		rrttl := ""
		if record.TTL != time.Duration(0) {
			rrttl = fmt.Sprintf("&rrttl=%d", int64(record.TTL/time.Second))
		}

		rrdistance := ""
		if record.Priority != 0 {
			rrdistance = fmt.Sprintf("&rrdistance=%d", record.Priority)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiEndpoint+"dnsUpdateRecord?version=1&type=xml&key="+p.APIToken+"&domain="+zoneToDomain(zone)+"&rrid="+record.ID+"&rrhost="+record.Name+"&rrvalue="+record.Value+rrttl+rrdistance, nil)
		if err != nil {
			return resultRecords, err
		}

		var response struct {
			reply
			ID string `xml:"reply>record_id"`
		}

		if err := doHttpRequestWithXmlResponse(client, req, &response); err != nil {
			return resultRecords, fmt.Errorf("request failed: %s", err)
		}

		if response.Code != 300 {
			return resultRecords, fmt.Errorf("failed to append record for zone \"%s\": Namesilo API status code %v; %s", zone, response.Code, response.Detail)
		}

		record.ID = response.ID

		resultRecords = append(resultRecords, record)
	}

	return resultRecords, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	client := &http.Client{}

	var deletedRecords []libdns.Record

	for _, record := range records {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiEndpoint+"dnsDeleteRecord?version=1&type=xml&key="+p.APIToken+"&domain="+zoneToDomain(zone)+"&rrid="+record.ID, nil)
		if err != nil {
			return deletedRecords, err
		}

		var response reply

		if err := doHttpRequestWithXmlResponse(client, req, &response); err != nil {
			return deletedRecords, fmt.Errorf("request to delete record failed: %s", err)
		}

		if response.Code != 300 {
			return deletedRecords, fmt.Errorf("failed to delete record for zone \"%s\": Namesilo API status code %v; %s", zone, response.Code, response.Detail)
		}

		deletedRecords = append(deletedRecords, record)
	}

	return deletedRecords, nil
}

func doHttpRequestWithXmlResponse(client *http.Client, req *http.Request, resp interface{}) error {
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(response.Body)
		return fmt.Errorf("response with unexpected status code %v; %s", response.StatusCode, string(respBody))
	}

	result, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("reading response failed: %v", err)
	}

	if err := xml.Unmarshal(result, resp); err != nil {
		return fmt.Errorf("unmarshaling of xml failed: %v", err)
	}

	return nil
}

func zoneToDomain(zone string) string {
	return strings.TrimSuffix(zone, ".")
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
