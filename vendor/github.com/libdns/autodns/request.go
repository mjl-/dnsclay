package autodns

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// buildRequest prepares the request with authentication headers and optional payload
func (p *Provider) buildRequest(ctx context.Context, method, url string, payload any) (req *http.Request, err error) {
	if p.Username == "" || p.Password == "" {
		err = fmt.Errorf("missing username and/or password")
		return
	}

	req, err = http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		err = fmt.Errorf("error creating request: %v", err)
		return
	}

	if payload != nil {
		buf := new(bytes.Buffer)
		if err = json.NewEncoder(buf).Encode(payload); err != nil {
			err = fmt.Errorf("Error encoding JSON: %v", err)
			return
		}
		req.Body = io.NopCloser(buf)
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("X-Domainrobot-Context", p.getAutoDNSContext())
	req.Header.Set("User-Agent", "libdns-autodns/x.y (+https://github.com/till/libdns-autodns)")
	req.SetBasicAuth(p.Username, p.Password)
	return
}

// buildURL prepends the endpoint to the requested API URL.
func (p *Provider) buildURL(path string) string {
	if p.Endpoint == "" {
		return autoDNSendpoint + "/" + path
	}

	return p.Endpoint + "/" + path
}

// getAutoDNSContext returns the provider / API context of the account.
func (p *Provider) getAutoDNSContext() string {
	if p.Context == "" {
		return autoDNScontext
	}

	return p.Context
}

// makeRequest executes the request.
func (p *Provider) makeRequest(req *http.Request) (*http.Response, error) {
	var client *http.Client
	if p.HttpClient == nil {
		client = p.HttpClient
	} else {
		client = p.HttpClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, err
}

// parseResponse parses the response into the struct.
func (p *Provider) parseResponse(resp *http.Response, into any) error {
	return json.NewDecoder(resp.Body).Decode(&into)
}
