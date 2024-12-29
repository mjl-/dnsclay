package cryptokeys

import "github.com/mittwald/go-powerdns/pdnshttp"

type client struct {
	httpClient *pdnshttp.Client
}

// New returns a new HTTP API Client
func New(hc *pdnshttp.Client) Client {
	return &client{
		httpClient: hc,
	}
}
