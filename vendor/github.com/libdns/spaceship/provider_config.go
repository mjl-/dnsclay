package libdnsspaceship

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Provider facilitates DNS record manipulation with Spaceship.
type Provider struct {
	// APIKey is the Spaceship API key for authentication
	APIKey string `json:"api_key,omitempty"`

	// APISecret is the Spaceship API secret for authentication
	APISecret string `json:"api_secret,omitempty"`

	// BaseURL is the base URL for the Spaceship API (defaults to https://spaceship.dev/api)
	BaseURL string `json:"base_url,omitempty"`

	// HTTPClient allows customization of the HTTP client used for API requests
	HTTPClient *http.Client `json:"-"`

	// PageSize controls pagination size used by GetRecords (defaults to 100)
	PageSize int `json:"page_size,omitempty"`
}

// listResponse models the GET /v1/dns/records/{domain} response
type listResponse struct {
	Items []spaceshipRecordUnion `json:"items"`
	Total int                    `json:"total"`
}

// NewProviderFromEnv constructs a Provider using environment variables.
// Recognized environment variables:
// - LIBDNS_SPACESHIP_APIKEY: API key (required for API calls)
// - LIBDNS_SPACESHIP_APISECRET: API secret (required for API calls)
// - LIBDNS_SPACESHIP_BASEURL: optional base URL override
// - LIBDNS_SPACESHIP_PAGESIZE: optional page size for list operations
// - LIBDNS_SPACESHIP_TIMEOUT: optional HTTP client timeout in seconds
func NewProviderFromEnv() *Provider {
	p := &Provider{}
	p.PopulateFromEnv()
	return p
}

// PopulateFromEnv fills unset Provider fields from environment variables.
func (p *Provider) PopulateFromEnv() {
	if p.APIKey == "" {
		p.APIKey = os.Getenv("LIBDNS_SPACESHIP_APIKEY")
	}
	if p.APISecret == "" {
		p.APISecret = os.Getenv("LIBDNS_SPACESHIP_APISECRET")
	}
	if p.BaseURL == "" {
		if v := os.Getenv("LIBDNS_SPACESHIP_BASEURL"); v != "" {
			p.BaseURL = v
		}
	}
	if p.PageSize == 0 {
		if v := os.Getenv("LIBDNS_SPACESHIP_PAGESIZE"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				p.PageSize = n
			}
		}
	}
	// If user hasn't provided HTTPClient and a timeout env var is present, set a client
	if p.HTTPClient == nil {
		if v := os.Getenv("LIBDNS_SPACESHIP_TIMEOUT"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				p.HTTPClient = &http.Client{Timeout: time.Duration(n) * time.Second}
			}
		}
	}
}

// validateCredentials checks if the required API credentials are set
func (p *Provider) validateCredentials() error {
	if p.APIKey == "" || p.APISecret == "" {
		return fmt.Errorf("API key and secret are required")
	}
	return nil
}
