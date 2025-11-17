// Package nanelo implements a DNS record management client compatible
// with the libdns interfaces for Nanelo
package nanelo

import (
	"context"
	"fmt"
	"encoding/json"
	"net/url"
	"net/http"
	"time"

	"github.com/libdns/libdns"
)

// Provider facilitates DNS record manipulation with Nanelo
type Provider struct {
	APIToken string `json:"api_token,omitempty"`
}

type APIResponse struct {
	OK     bool                    `json:"ok"`
	Error  *string                 `json:"error"`
	Result *map[string]interface{} `json:"result"`
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	baseURL, _ := url.Parse("https://api.nanelo.com/v1")
	baseURL = baseURL.JoinPath(p.APIToken, "dns", "addrecord")

	var added []libdns.Record

	for _, rec := range records {
		rr := rec.RR()

		endpoint := baseURL
		query := endpoint.Query()
		query.Set("domain", zone)
		query.Set("name", rr.Name)
		query.Set("type", rr.Type)
		query.Set("value", rr.Data)
		query.Set("ttl", fmt.Sprintf("%f", rr.TTL.Seconds()))

		// TODO: support setting the "priority"

		endpoint.RawQuery = query.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), http.NoBody)
		if err != nil {
			return nil, err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var apiResponse APIResponse
		err = json.NewDecoder(resp.Body).Decode(&apiResponse)
		if err != nil {
			return nil, err
		}

		if apiResponse.Error != nil {
			return nil, fmt.Errorf(*apiResponse.Error)
		}
		if !apiResponse.OK {
			return nil, fmt.Errorf("unknown error when trying to create the DNS record")
		}

		if rr.TTL == 0 {
			rr.TTL = time.Minute // Default TTL if not set
		}
		added = append(added, rr)
	}

	return added, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	baseURL, _ := url.Parse("https://api.nanelo.com/v1")
	baseURL = baseURL.JoinPath(p.APIToken, "dns", "deleterecord")

	var deleted []libdns.Record

	for _, rec := range records {
		rr := rec.RR()

		endpoint := baseURL
		query := endpoint.Query()
		query.Set("domain", zone)
		query.Set("name", rr.Name)
		query.Set("type", rr.Type)
		query.Set("value", rr.Data)
		query.Set("ttl", fmt.Sprintf("%f", rr.TTL.Seconds()))

		// TODO: support setting the "priority"

		endpoint.RawQuery = query.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), http.NoBody)
		if err != nil {
			return nil, err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var apiResponse APIResponse
		err = json.NewDecoder(resp.Body).Decode(&apiResponse)
		if err != nil {
			return nil, err
		}

		if apiResponse.Error != nil {
			return nil, fmt.Errorf(*apiResponse.Error)
		}
		if !apiResponse.OK {
			return nil, fmt.Errorf("unknown error when trying to delete the DNS record")
		}

		deleted = append(deleted, rec)
	}

	return deleted, nil
}

// Interface guards
var (
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
