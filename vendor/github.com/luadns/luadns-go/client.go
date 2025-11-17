package luadns

import (
	"context"
	"encoding/json"
	"fmt"
)

// OptFunc represents a configuration function which are used to configure the REST API client.
type OptFunc func(*Client)

// SetBaseURL sets a custom baseURL for API requests (used in unit tests).
func SetBaseURL(url string) OptFunc {
	return func(c *Client) {
		c.baseURL = url
	}
}

// RestCallFunc represents a call to REST API.
type RestCallFunc func(ctx context.Context) ([]byte, error)

// Client represents a REST API client for LuaDNS API.
type Client struct {
	baseURL string
	client  *JSONClient
}

// NewClient initializes the REST API client and configures authentication.
func NewClient(email, apiKey string, opts ...OptFunc) *Client {
	c := &Client{
		baseURL: baseURL,
		client:  NewAuthJSONClient(email, apiKey),
	}

	// Apply custom options.
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// UserAgent appends product text to User-Agent header used in HTTP requests.
func (c *Client) UserAgent(product string) {
	if product != "" {
		c.client.userAgent += " " + product
	}
}

// endpoint is a helper which builds the endpoint URL.
func (c *Client) endpoint(format string, args ...any) string {
	return c.baseURL + fmt.Sprintf(format, args...)
}

// do executes REST call and serializes the response into `dest` target.
func (c *Client) do(ctx context.Context, fn RestCallFunc, dest any) error {
	data, err := fn(ctx)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &dest)
}
