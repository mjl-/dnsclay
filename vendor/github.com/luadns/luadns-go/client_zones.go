package luadns

import (
	"context"
	"net/url"
)

// ListZones returns user zones.
//
// See: http://www.luadns.com/api.html#list-zones
func (c *Client) ListZones(ctx context.Context, options *ListParams, handlers ...HandlerFunc) ([]*Zone, error) {
	zones := []*Zone{}

	req := func(ctx context.Context) ([]byte, error) {
		uri := &url.URL{
			Path:     "/zones",
			RawQuery: options.QueryString(),
		}

		return c.client.Get(ctx, c.endpoint(uri.String()), handlers...)
	}

	err := c.do(ctx, req, &zones)
	if err != nil {
		return nil, err
	}

	return zones, nil
}

// CreateZone creates a new zone using supplied attributes.
//
// See: http://www.luadns.com/api.html#create-a-zone
func (c *Client) CreateZone(ctx context.Context, attrs *Zone) (*Zone, error) {
	var zone Zone

	req := func(ctx context.Context) ([]byte, error) {
		return c.client.Post(ctx, c.endpoint("/zones"), attrs)
	}

	err := c.do(ctx, req, &zone)
	if err != nil {
		return nil, err
	}

	return &zone, nil
}

// GetZone get a specific zone identfied by `zoneID`.
//
// See: http://www.luadns.com/api.html#get-a-zone
func (c *Client) GetZone(ctx context.Context, zoneID int64) (*Zone, error) {
	var zone Zone

	req := func(ctx context.Context) ([]byte, error) {
		return c.client.Get(ctx, c.endpoint("/zones/%d", zoneID))
	}

	err := c.do(ctx, req, &zone)
	if err != nil {
		return nil, err
	}

	return &zone, nil
}

// UpdateZone zone identified by `zoneID` using supplied attributes.
//
// See: http://www.luadns.com/api.html#update-a-zone
func (c *Client) UpdateZone(ctx context.Context, zoneID int64, attrs *Zone) (*Zone, error) {
	var zone Zone

	req := func(ctx context.Context) ([]byte, error) {
		return c.client.Put(ctx, c.endpoint("/zones/%d", zoneID), attrs)
	}

	err := c.do(ctx, req, &zone)
	if err != nil {
		return nil, err
	}

	return &zone, nil
}

// DeleteZone delete specific zone using zone ID.
//
// See: http://www.luadns.com/api.html#delete-a-zone
func (c *Client) DeleteZone(ctx context.Context, zoneID int64) (*Zone, error) {
	var zone Zone

	req := func(ctx context.Context) ([]byte, error) {
		return c.client.Delete(ctx, c.endpoint("/zones/%d", zoneID))
	}

	err := c.do(ctx, req, &zone)
	if err != nil {
		return nil, err
	}

	return &zone, nil
}
