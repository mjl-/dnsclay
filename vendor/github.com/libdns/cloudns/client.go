package cloudns

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"slices"

	"github.com/libdns/libdns"
)

const success = "Success"

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

// GetClouDNSRecords returns the raw upstream results from ClouDNS.
// For use when the IDs of the individual records needs to be preserved, which
// cannot be done with the generic libdns.Record interface.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - zone: The DNS zone (domain) to retrieve records from
//
// Returns:
//   - []ApiDnsRecord: Slice of all DNS records in the zone
//   - error: Any error that occurred during the operation
func (c *Client) GetClouDNSRecords(ctx context.Context, zone string) ([]ApiDnsRecord, error) {
	recordsEndpoint := apiBaseUrl.JoinPath("records.json")
	params := map[string]string{
		"domain-name": zone,
	}

	// Perform the API request
	resp, err := c.performGetRequest(ctx, recordsEndpoint, params)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned non-OK status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the API response
	var apiResult map[string]ApiDnsRecord
	if err := json.NewDecoder(resp.Body).Decode(&apiResult); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	return slices.Collect(maps.Values(apiResult)), nil
}

// GetRecords retrieves DNS records for the specified zone.
// It returns a slice of libdns.Record or an error if the request fails.
func (c *Client) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	apiResult, err := c.GetClouDNSRecords(ctx, zone)
	if err != nil {
		return nil, err
	}

	records := make([]libdns.Record, 0, len(apiResult))
	for _, recordData := range apiResult {
		record, err := recordData.toLibdnsRecord()
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	// Log the number of records found
	fmt.Printf("Found %d records in zone %s\n", len(records), zone)

	return records, nil
}

// AddRecord creates a new DNS record in the specified zone with the given properties and returns the created record or an error.
// It handles API communication, response parsing, and error handling.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - zone: The DNS zone (domain) to add the record to
//   - record: The DNS record to add
//
// Returns:
//   - libdns.Record: The created record
//   - error: Any error that occurred during the operation
func (c *Client) AddRecord(ctx context.Context, zone string, record ApiDnsRecord) (libdns.Record, error) {
	endpoint := apiBaseUrl.JoinPath("add-record.json")

	params := record.toParameters()
	params["domain-name"] = zone
	resp, err := c.performPostRequest(ctx, endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned non-OK status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the API response
	var resultModel ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&resultModel); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	// Check if the operation was successful
	if resultModel.Status != success {
		return nil, fmt.Errorf("API operation failed: %s", resultModel.StatusDescription)
	}

	return record.toLibdnsRecord()
}

// UpdateRecord updates an existing DNS record in the specified zone with the provided values and returns the updated record.
// It handles API communication, response parsing, and error handling.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - zone: The DNS zone (domain) containing the record
//   - record: The record to update
//
// Returns:
//   - libdns.Record: The updated record
//   - error: Any error that occurred during the operation
func (c *Client) UpdateRecord(ctx context.Context, zone string, record ApiDnsRecord) (libdns.Record, error) {
	updateEndpoint := apiBaseUrl.JoinPath("mod-record.json")

	params := record.toParameters()
	params["domain-name"] = zone
	resp, err := c.performPostRequest(ctx, updateEndpoint, params)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned non-OK status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the API response
	var resultModel ApiResponse
	if err = json.NewDecoder(resp.Body).Decode(&resultModel); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	// Check if the operation was successful
	if resultModel.Status != success {
		return nil, fmt.Errorf("API operation failed: %s", resultModel.StatusDescription)
	}

	ret, err := record.toLibdnsRecord()
	if err != nil {
		return nil, fmt.Errorf("failed to get existing record details: %w", err)
	}

	return ret, nil
}

// DeleteRecord deletes a DNS record identified by its ID in the specified zone.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - zone: The DNS zone (domain) containing the record
//   - recordId: ID of the record to delete
//
// Returns:
//   - libdns.Record: The deleted record, or nil if the record was not found
//   - error: Any error that occurred during the operation
func (c *Client) DeleteRecord(ctx context.Context, zone string, recordId string) error {
	endpoint := apiBaseUrl.JoinPath("delete-record.json")
	params := map[string]string{
		"domain-name": zone,
		"record-id":   recordId,
	}

	// Perform the API request
	resp, err := c.performPostRequest(ctx, endpoint, params)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned non-OK status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the API response
	var resultModel ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&resultModel); err != nil {
		return fmt.Errorf("failed to decode API response: %w", err)
	}

	// Check if the operation was successful
	if resultModel.Status != success {
		return fmt.Errorf("API operation failed: %s", resultModel.StatusDescription)
	}

	return nil
}

// performPostRequest sends a POST request to the specified URL with query parameters and returns the HTTP response or an error.
// It adds authentication parameters and builds the request with the provided context.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - targetURL: The API endpoint URL
//   - params: Map of query parameters to include in the request
//
// Returns:
//   - *http.Response: The HTTP response from the API
//   - error: Any error that occurred during the request
func (c *Client) performPostRequest(ctx context.Context, targetURL *url.URL, params map[string]string) (*http.Response, error) {
	// Create a copy of the URL to avoid modifying the original
	requestURL := *targetURL

	// Get query parameters and add authentication
	queries := requestURL.Query()
	c.addAuthParams(queries)

	// Add all provided parameters to the query
	for k, v := range params {
		queries.Set(k, v)
	}

	// Encode the query parameters and set them on the URL
	requestURL.RawQuery = queries.Encode()

	// Create a new HTTP request with the provided context
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set appropriate headers
	req.Header.Set("User-Agent", "cloudns-go-client/1.0")
	req.Header.Set("Accept", "application/json")

	// Execute the request
	return http.DefaultClient.Do(req)
}

// addAuthParams adds authentication parameters to the provided query values based on the client's credentials.
// It selects between auth-id and sub-auth-id based on which is provided.
//
// Parameters:
//   - queries: The url.Values to add authentication parameters to
func (c *Client) addAuthParams(queries url.Values) {
	// CloudNS API requires either auth-id or sub-auth-id, but not both
	if c.SubAuthId != "" {
		queries.Set("sub-auth-id", c.SubAuthId)
	} else {
		queries.Set("auth-id", c.AuthId)
	}

	// Always include the auth password
	queries.Set("auth-password", c.AuthPassword)
}

// performGetRequest sends a GET request to the specified URL with query parameters and returns the HTTP response or an error.
// It adds authentication parameters and builds the request with the provided context.
//
// Parameters:
//   - ctx: Context for timeout and cancellation
//   - targetURL: The API endpoint URL
//   - params: Map of query parameters to include in the request
//
// Returns:
//   - *http.Response: The HTTP response from the API
//   - error: Any error that occurred during the request
func (c *Client) performGetRequest(ctx context.Context, targetURL *url.URL, params map[string]string) (*http.Response, error) {
	// Create a copy of the URL to avoid modifying the original
	requestURL := *targetURL

	// Get query parameters and add authentication
	queries := requestURL.Query()
	c.addAuthParams(queries)

	// Add all provided parameters to the query
	for k, v := range params {
		queries.Set(k, v)
	}

	// Encode the query parameters and set them on the URL
	requestURL.RawQuery = queries.Encode()

	// Create a new HTTP request with the provided context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set appropriate headers
	req.Header.Set("User-Agent", "cloudns-go-client/1.0")
	req.Header.Set("Accept", "application/json")

	// Execute the request
	return http.DefaultClient.Do(req)
}
