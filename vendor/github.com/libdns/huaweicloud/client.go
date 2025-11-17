package huaweicloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strings"

	"github.com/libdns/libdns"
)

type Client struct {
	accessKeyId     string
	secretAccessKey string
	region          string
	singer          *Signer
}

// NewClient creates a new Huawei Cloud DNS client.
func NewClient(accessKeyId, secretAccessKey, region string) *Client {
	if region == "" {
		region = "cn-south-1"
	}

	client := &Client{
		accessKeyId:     accessKeyId,
		secretAccessKey: secretAccessKey,
		region:          region,
		singer: &Signer{
			Key:    accessKeyId,
			Secret: secretAccessKey,
		},
	}

	return client
}

func (c *Client) GetRecords(ctx context.Context, zone string) ([]RecordSet, error) {
	zoneId, err := c.getZoneId(ctx, zone)
	if err != nil {
		return nil, err
	}

	url := c.getBaseURL()
	url = url.JoinPath("zones", zoneId, "recordsets")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)

	resp := new(ListRecordsResponse)
	if err = c.doAPIRequest(req, resp); err != nil {
		return nil, err
	}

	return resp.RecordSets, nil
}

func (c *Client) AppendRecord(ctx context.Context, zone string, record RecordSet) (*RecordSet, error) {
	zoneId, err := c.getZoneId(ctx, zone)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}

	url := c.getBaseURL()
	url = url.JoinPath("zones", zoneId, "recordsets")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp := new(RecordSet)
	if err = c.doAPIRequest(req, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) UpdateRecord(ctx context.Context, zone string, record RecordSet) (*RecordSet, error) {
	zoneId, err := c.getZoneId(ctx, zone)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}

	url := c.getBaseURL()
	url = url.JoinPath("zones", zoneId, "recordsets", record.Id)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp := new(RecordSet)
	if err = c.doAPIRequest(req, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) DeleteRecord(ctx context.Context, zone string, recordId string) (*RecordSet, error) {
	zoneId, err := c.getZoneId(ctx, zone)
	if err != nil {
		return nil, err
	}

	url := c.getBaseURL()
	url = url.JoinPath("zones", zoneId, "recordsets", recordId)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
	if err != nil {
		return nil, err
	}

	resp := new(RecordSet)
	if err = c.doAPIRequest(req, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) GetRecordId(ctx context.Context, zone, recName, recType string, recVal ...string) (string, error) {
	zoneId, err := c.getZoneId(ctx, zone)
	if err != nil {
		return "", err
	}

	url := c.getBaseURL()
	url = url.JoinPath("zones", zoneId, "recordsets")
	query := url.Query()
	query.Set("search_mode", "equal")
	query.Set("type", recType)
	query.Set("name", libdns.AbsoluteName(recName, zone))
	url.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)

	resp := new(ListRecordsResponse)
	if err = c.doAPIRequest(req, resp); err != nil {
		return "", err
	}

	if len(resp.RecordSets) == 0 {
		return "", fmt.Errorf("record %q not found", recName)
	}
	if len(resp.RecordSets) != 1 {
		return "", fmt.Errorf("returned more than one record for %q, expected one, actual %d", recName, len(resp.RecordSets))
	}

	if len(recVal) > 0 && recVal[0] != "" {
		rec, err := resp.RecordSets[0].libdnsRecord(zone)
		if err != nil {
			return "", err
		}
		for _, r := range rec {
			rr := r.RR()
			if rr.Data == recVal[0] {
				return resp.RecordSets[0].Id, nil
			}
		}
	}

	return resp.RecordSets[0].Id, nil
}

func (c *Client) getZoneId(ctx context.Context, zone string) (string, error) {
	zone = strings.TrimSuffix(zone, ".")

	url := c.getBaseURL()
	url = url.JoinPath("zones")
	query := url.Query()
	query.Set("name", strings.TrimSuffix(zone, "."))
	query.Set("search_mode", "equal")
	url.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)

	resp := new(ListZonesResponse)
	if err = c.doAPIRequest(req, resp); err != nil {
		return "", err
	}

	if len(resp.Zones) == 0 {
		return "", fmt.Errorf("zone %q not found", zone)
	}
	if len(resp.Zones) != 1 {
		return "", fmt.Errorf("returned more than one zone for %q, expected one, actual %d", zone, len(resp.Zones))
	}

	return resp.Zones[0].Id, nil
}

func (c *Client) getBaseURL() *neturl.URL {
	baseURL, _ := neturl.Parse("https://dns." + c.region + ".myhuaweicloud.com/v2")
	return baseURL
}

func (c *Client) doAPIRequest(req *http.Request, result any) error {
	if err := c.singer.Sign(req); err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("got error status: HTTP %d: %+v", resp.StatusCode, string(body))
	}

	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	return nil
}
