package api

import (
	"net/http"

	"github.com/pkg/errors"
)

const (
	apiURL = "https://api.dreamhost.com/"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

type requestFormat string

const (
	jsonFormat requestFormat = "json"
	csvFormat  requestFormat = "csv" // nolint:varcheck,deadcode
)

func NewClient(apiKey string, httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	client := Client{
		apiKey:     apiKey,
		httpClient: httpClient,
	}
	if err := client.validate(); err != nil {
		return nil, err
	}
	return &client, nil
}

func (c *Client) validate() error {
	if c.apiKey == "" {
		return errors.New("empty API key provided")
	}
	if c.httpClient == nil {
		return errors.New("nil http client")
	}
	return nil
}
