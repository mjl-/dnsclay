package infomaniak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// Client that abstracts and calls infomaniak API
type Client struct {
	// infomaniak API token
	Token string

	// http client used for requests
	HttpClient *http.Client
}

// GetDnsRecordsForZone loads all dns records for a given zone
func (c *Client) GetDnsRecordsForZone(ctx context.Context, zone string) ([]IkRecord, error) {
	var dnsRecords []IkRecord
	_, err := c.doRequest(ctx, http.MethodGet, getRecordsEndpointUrl(zone), nil, &dnsRecords)
	if err != nil {
		return nil, err
	}
	unescapeTargets(dnsRecords)
	return dnsRecords, nil
}

// CreateOrUpdateRecord creates a record if its Id property is not set, otherwise it updates the record
func (c *Client) CreateOrUpdateRecord(ctx context.Context, zone string, record IkRecord) (*IkRecord, error) {
	rawJson, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}

	isNew := record.ID == 0
	var method = http.MethodPost
	var endpoint = getRecordsEndpointUrl(zone)

	if !isNew {
		method = http.MethodPut
		endpoint = getRecordEndpointUrl(zone, fmt.Sprint(record.ID), false)
	}

	var updatedRecord IkRecord
	_, err = c.doRequest(ctx, method, endpoint, bytes.NewBuffer(rawJson), &updatedRecord)
	if err != nil {
		return nil, err
	}
	unescapeTarget(&updatedRecord)
	return &updatedRecord, nil
}

// DeleteRecord deletes an existing dns record for a given zone
func (c *Client) DeleteRecord(ctx context.Context, zone string, recordId string) error {
	_, err := c.doRequest(ctx, http.MethodDelete, getRecordEndpointUrl(zone, recordId, true), nil, nil)
	return err
}

// GetFqdnOfZoneForDomain returns the FQDN of the zone managed by infomaniak
func (c *Client) GetFqdnOfZoneForDomain(ctx context.Context, domain string) (string, error) {
	var zones []IkZone
	_, err := c.doRequest(ctx, http.MethodGet, getZonesEndpointUrl(domain), nil, &zones)
	if err != nil {
		return "", err
	}

	sort.Slice(zones, func(a int, b int) bool {
		return len(zones[a].Fqdn) > len(zones[b].Fqdn)
	})

	for _, managedZone := range zones {
		if strings.HasSuffix(domain, managedZone.Fqdn) {
			return managedZone.Fqdn, nil
		}
	}

	return "", fmt.Errorf("could not find the zone managed by infomaniak for %s", domain)
}

// doRequest performs the API call for the given parameters and parses the response's data to the given responseData struct - if the parameter is not nil
func (c *Client) doRequest(ctx context.Context, method, url string, requestBody io.Reader, responseData any) (*IkResponse, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, requestBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	rawResp, err := c.HttpClient.Do(req)

	if err != nil {
		return nil, err
	}
	defer rawResp.Body.Close()

	var resp IkResponse
	err = json.NewDecoder(rawResp.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}

	if rawResp.StatusCode >= 400 || resp.Result != "success" {
		return nil, fmt.Errorf("got errors: HTTP %d: %+v", rawResp.StatusCode, string(resp.Error))
	}

	if responseData != nil {
		err = json.Unmarshal(resp.Data, responseData)
		if err != nil {
			return nil, err
		}
	}

	return &resp, nil
}

// unescapeTargets makes sure all record's target value conforms to *unescaped* standard zone file syntax
func unescapeTargets(recs []IkRecord) {
	for i := range recs {
		unescapeTarget(&recs[i])
	}
}

// unescapeTarget makes sure record target value conforms to *unescaped* standard zone file syntax
func unescapeTarget(rec *IkRecord) {
	unquoted, err := strconv.Unquote(rec.Target)
	if err == nil {
		rec.Target = unquoted
	}
}

const apiBaseUrl = "https://api.infomaniak.com"
const recordsPath = apiBaseUrl + "/2/zones/%s/records%s"
const recordDetailParam = "with=records_description"

// getRecordEndpointUrl returns API endpoint for a specific, already existing record
func getRecordEndpointUrl(zone string, recordId string, isDelete bool) string {
	param := "/" + recordId
	if !isDelete {
		param = param + "?" + recordDetailParam
	}
	return fmt.Sprintf(recordsPath, zone, param)
}

// getRecordsEndpointUrl returns API endpoint for all records of a zone
func getRecordsEndpointUrl(zone string) string {
	return fmt.Sprintf(recordsPath, zone, "?"+recordDetailParam)
}

// getRecordsEndpointUrl returns API endpoint for all records of a zone
func getZonesEndpointUrl(domain string) string {
	return fmt.Sprintf("%s/2/domains/%s/zones", apiBaseUrl, domain)
}
