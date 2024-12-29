package dynu

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
)

const defaultBaseURL = "https://api.dynu.com/v2"

type Client struct {
	baseURL    *url.URL
	HTTPClient *http.Client
	APIToken   string

	mutex sync.Mutex
}

func NewClient(APIToken string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
		APIToken:   APIToken,
	}
}

func (c *Client) joinUrlPath(elem ...string) *url.URL {
	return c.baseURL.JoinPath(elem...)
}

func (c *Client) GetRootDomain(ctx context.Context, hostname string) (*DNSHostname, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	endpoint := c.joinUrlPath("dns", "getroot", hostname)
	apiResponse := DNSHostname{}
	apiException := APIException{}
	err := c.doRetryWithCustomError(ctx, http.MethodGet, endpoint.String(), nil, &apiResponse, &apiException)
	if err != nil {
		return nil, err
	}

	if apiResponse.StatusCode != 200 {
		return nil, fmt.Errorf("API error: %w", apiException)
	}

	return &apiResponse, nil

}

func (c *Client) GetRecords(ctx context.Context, hostnameId int64) ([]DNSRecord, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	endpoint := c.joinUrlPath("dns", fmt.Sprint(hostnameId), "record")

	apiResponse := RecordsResponse{}
	apiException := APIException{}
	err := c.doRetryWithCustomError(ctx, http.MethodGet, endpoint.String(), nil, &apiResponse, &apiException)
	if err != nil {
		return nil, err
	}

	if apiResponse.StatusCode != 200 {
		return nil, fmt.Errorf("API error: %w", apiException)
	}

	return apiResponse.DNSRecords, nil
}

func (c *Client) AddOrUpdateRecord(ctx context.Context, hostnameId int64, record DNSRecord, ignoreRecordId bool) (*DNSRecord, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	urlPaths := []string{"dns", fmt.Sprint(hostnameId), "record"}
	if record.ID != 0 && !ignoreRecordId {
		urlPaths = append(urlPaths, fmt.Sprint(record.ID))
	}

	endpoint := c.joinUrlPath(urlPaths...)

	reqBody, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to create request JSON body: %w", err)
	}

	apiResponse := DNSRecord{}
	apiException := APIException{}
	err = c.doRetryWithCustomError(ctx, http.MethodPost, endpoint.String(), reqBody, &apiResponse, &apiException)
	if err != nil {
		return nil, err
	}

	if apiResponse.StatusCode != 200 {
		return nil, fmt.Errorf("API error: %w", apiException)
	}

	return &apiResponse, nil
}

func (c *Client) DeleteRecord(ctx context.Context, hostnameId int64, dnsRecordId string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	endpoint := c.joinUrlPath("dns", fmt.Sprint(hostnameId), "record", dnsRecordId)

	apiResponse := DeleteResponse{}
	apiException := APIException{}
	err := c.doRetryWithCustomError(ctx, http.MethodDelete, endpoint.String(), nil, &apiResponse, &apiException)
	if err != nil {
		return err
	}

	if apiResponse.StatusCode != 200 {
		return fmt.Errorf("API error: %w", apiException)
	}

	return nil
}

// retry with exponential backoff on EOF
func (c *Client) doRetryWithCustomError(ctx context.Context, method, uri string, body []byte, result any, errorResult any) error {
	operation := func() error {
		return c.doWithCustomError(ctx, method, uri, body, result, errorResult)
	}

	notify := func(err error, duration time.Duration) {
		log.Printf("client retrying in %s because of %v", duration, err)
	}

	bo := backoff.NewExponentialBackOff(backoff.WithInitialInterval(1 * time.Second))

	err := backoff.RetryNotify(operation, bo, notify)
	return err
}

// exception fields are at the top level of json rather than nested under exception object; parse json again as custom exception object for error logging
func (c *Client) doWithCustomError(ctx context.Context, method, uri string, body []byte, result any, errorResult any) error {
	var reqBody io.Reader
	if len(body) > 0 {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, uri, reqBody)
	if err != nil {
		return backoff.Permanent(fmt.Errorf("unable to create request: %w", err))
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("API-Key", c.APIToken)

	resp, err := c.HTTPClient.Do(req)
	if errors.Is(err, io.EOF) {
		return err
	}

	if err != nil {
		return backoff.Permanent(err)
	}

	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return backoff.Permanent(err)
	}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return backoff.Permanent(err)
	}

	if errorResult != nil {
		_ = json.Unmarshal(raw, errorResult)
	}

	return nil
}
