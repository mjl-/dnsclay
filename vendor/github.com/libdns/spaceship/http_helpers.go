package libdnsspaceship

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// getHTTPClient returns the HTTP client to use for API requests
func (p *Provider) getHTTPClient() *http.Client {
	if p.HTTPClient != nil {
		return p.HTTPClient
	}
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}

// getBaseURL returns the base URL for API requests
func (p *Provider) getBaseURL() string {
	if p.BaseURL != "" {
		return strings.TrimSuffix(p.BaseURL, "/")
	}
	// Default from the OpenAPI servers
	return "https://spaceship.dev/api"
}

// doRequest performs an HTTP request to the Spaceship API and returns response body and status code
func (p *Provider) doRequest(ctx context.Context, method, endpoint string, body interface{}) ([]byte, int, error) {
	url := p.getBaseURL() + endpoint
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Use API key/secret headers as described in the OpenAPI spec
	if p.APIKey != "" {
		req.Header.Set("X-API-Key", p.APIKey)
	}
	if p.APISecret != "" {
		req.Header.Set("X-API-Secret", p.APISecret)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	client := p.getHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return respBody, resp.StatusCode, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, resp.StatusCode, nil
}
