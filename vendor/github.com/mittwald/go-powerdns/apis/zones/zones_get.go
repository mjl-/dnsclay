package zones

import (
	"context"
	"fmt"
	"github.com/mittwald/go-powerdns/pdnshttp"
	"net/http"
	"net/url"
)

type GetZoneOption interface {
	ApplyToGetZoneRequest(req *http.Request) error
}

type getZoneOptionFunc func(req *http.Request) error

func (g getZoneOptionFunc) ApplyToGetZoneRequest(req *http.Request) error {
	return g(req)
}

func WithoutResourceRecordSets() GetZoneOption {
	return getZoneOptionFunc(func(req *http.Request) error {
		_ = pdnshttp.WithQueryValue("rrsets", "false")(req)
		return nil
	})
}

func WithResourceRecordSetFilter(name, recordType string) GetZoneOption {
	return getZoneOptionFunc(func(req *http.Request) error {
		_ = pdnshttp.WithQueryValue("rrset_name", name)(req)
		_ = pdnshttp.WithQueryValue("rrset_type", recordType)(req)
		return nil
	})
}

func (c *client) GetZone(ctx context.Context, serverID, zoneID string, opts ...GetZoneOption) (*Zone, error) {
	zone := Zone{}
	path := fmt.Sprintf("/servers/%s/zones/%s", url.PathEscape(serverID), url.PathEscape(zoneID))

	req, err := c.httpClient.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		if err := opt.ApplyToGetZoneRequest(req); err != nil {
			return nil, err
		}
	}

	if err := c.httpClient.Do(ctx, req, &zone); err != nil {
		if e, ok := err.(pdnshttp.ErrUnexpectedStatus); ok {
			if e.StatusCode == http.StatusUnprocessableEntity {
				return nil, pdnshttp.ErrNotFound{}
			}
		}

		return nil, err
	}

	return &zone, nil
}
