package cloudns

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/libdns/libdns"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	AuthId       string `json:"auth_id"`
	SubAuthId    string `json:"sub_auth_id"`
	AuthPassword string `json:"auth_password"`
}

var apiBaseUrl, _ = url.Parse("https://api.cloudns.net/dns/")

// UseClient initializes and returns a new Client instance with provided authentication details.
func UseClient(authId, subAuthId, authPassword string) *Client {
	return &Client{
		AuthId:       authId,
		SubAuthId:    subAuthId,
		AuthPassword: authPassword,
	}
}

// GetRecords retrieves DNS records for the specified zone.
// It returns a slice of libdns.Record or an error if the request fails.
func (c *Client) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	recordsEndpoint := apiBaseUrl.JoinPath("records.json")
	resp, err := c.performGetRequest(ctx, recordsEndpoint, map[string]string{
		"domain-name": zone,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var apiResult map[string]ApiDnsRecord
	if err := json.NewDecoder(resp.Body).Decode(&apiResult); err != nil {
		return nil, errors.New("failed to decode response")
	}

	records := make([]libdns.Record, 0, len(apiResult))
	for _, recordData := range apiResult {
		records = append(records, libdns.Record{
			ID:    recordData.Id,
			Type:  recordData.Type,
			Name:  recordData.Host,
			TTL:   parseDuration(recordData.Ttl + "s"),
			Value: recordData.Record,
		})
	}
	return records, nil
}

// GetRecord retrieves a specific DNS record by its ID from the specified zone.
// It returns a pointer to the matching libdns.Record or an error if the record is not found or the retrieval fails.
func (c *Client) GetRecord(ctx context.Context, zone, recordID string) (*libdns.Record, error) {
	records, err := c.GetRecords(ctx, zone)
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		if record.ID == recordID {
			return &record, nil
		}
	}
	return nil, errors.New("record not found")
}

// AddRecord creates a new DNS record in the specified zone with the given properties and returns the created record or an error.
func (c *Client) AddRecord(ctx context.Context, zone string, recordType string, recordHost string, recordValue string, ttl time.Duration) (*libdns.Record, error) {
	endpoint := apiBaseUrl.JoinPath("add-record.json")

	roundedTTL := ttlRounder(ttl)
	roundedTTLStr := strconv.Itoa(roundedTTL)

	resp, err := c.performPostRequest(ctx, endpoint, map[string]string{
		"domain-name": zone,
		"record-type": recordType,
		"host":        recordHost,
		"record":      recordValue,
		"ttl":         roundedTTLStr,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var resultModel ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&resultModel); err != nil {
		return nil, err
	}

	if resultModel.Status != "Success" {
		return nil, errors.New(resultModel.StatusDescription)
	}

	parsedTTL := parseDuration(roundedTTLStr + "s")

	return &libdns.Record{
		ID:    strconv.Itoa(resultModel.Data.Id),
		Type:  recordType,
		Name:  recordHost,
		TTL:   parsedTTL,
		Value: recordValue,
	}, nil
}

// UpdateRecord updates an existing DNS record in the specified zone with the provided values and returns the updated record.
func (c *Client) UpdateRecord(ctx context.Context, zone string, recordID string, host string, recordValue string, ttl time.Duration) (*libdns.Record, error) {
	updateEndpoint := apiBaseUrl.JoinPath("mod-record.json")

	ttlSec := ttlRounder(ttl)
	resp, err := c.performPostRequest(ctx, updateEndpoint, map[string]string{
		"domain-name": zone,
		"record-id":   recordID,
		"host":        host,
		"record":      recordValue,
		"ttl":         strconv.Itoa(ttlSec),
	})

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var resultModel ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&resultModel); err != nil {
		return nil, err
	}

	if resultModel.Status != "Success" {
		return nil, errors.New(resultModel.StatusDescription)
	}

	existingRecord, err := c.GetRecord(ctx, zone, recordID)
	if err != nil {
		return nil, err
	}

	return &libdns.Record{
		ID:    recordID,
		Type:  existingRecord.Type,
		Name:  host,
		TTL:   parseDuration(strconv.Itoa(ttlSec) + "s"),
		Value: recordValue,
	}, nil
}

// DeleteRecord deletes a DNS record identified by its ID in the specified zone.
// It returns the deleted libdns.Record or nil if the record was not found, and an error if the operation fails.
func (c *Client) DeleteRecord(ctx context.Context, zone string, recordId string) (*libdns.Record, error) {
	rInfo, err := c.GetRecord(ctx, zone, recordId)
	if err != nil {
		if err.Error() == "record not found" {
			return nil, nil
		}
		return nil, err
	}

	endpoint := apiBaseUrl.JoinPath("delete-record.json")
	resp, err := c.performPostRequest(ctx, endpoint, map[string]string{
		"domain-name": zone,
		"record-id":   recordId,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var resultModel ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&resultModel); err != nil {
		return nil, err
	}

	if resultModel.Status != "Success" {
		return nil, errors.New(resultModel.StatusDescription)
	}

	return rInfo, nil
}

// performPostRequest sends a POST request to the specified URL with query parameters and returns the HTTP response or an error.
func (c *Client) performPostRequest(ctx context.Context, targetURL *url.URL, params map[string]string) (*http.Response, error) {
	queries := targetURL.Query()
	c.addAuthParams(queries)

	for k, v := range params {
		queries.Set(k, v)
	}

	targetURL.RawQuery = queries.Encode()
	req, requestErr := http.NewRequestWithContext(ctx, http.MethodPost, targetURL.String(), nil)
	if requestErr != nil {
		return nil, requestErr
	}

	return http.DefaultClient.Do(req)
}

// addAuthParams adds authentication parameters to the provided query values based on the client's credentials.
func (c *Client) addAuthParams(queries url.Values) {
	if c.SubAuthId != "" {
		queries.Set("sub-auth-id", c.SubAuthId)
	} else {
		queries.Set("auth-id", c.AuthId)
	}
	queries.Set("auth-password", c.AuthPassword)
}

// performGetRequest sends a GET request to the specified URL with query parameters and returns the HTTP response or an error.
func (c *Client) performGetRequest(ctx context.Context, targetURL *url.URL, params map[string]string) (*http.Response, error) {
	queries := targetURL.Query()
	c.addAuthParams(queries) // Use extracted function instead of duplicating auth logic

	for k, v := range params {
		queries.Set(k, v)
	}

	targetURL.RawQuery = queries.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL.String(), nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}
