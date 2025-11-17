package luadns

import (
	"context"
	"fmt"
	"net/url"
)

// ListRecords returns zone records.
//
// See: http://www.luadns.com/api.html#list-records
func (c *Client) ListRecords(ctx context.Context, zone *Zone, options *ListParams, handlers ...HandlerFunc) ([]*Record, error) {
	records := []*Record{}

	req := func(ctx context.Context) ([]byte, error) {
		uri := &url.URL{
			Path:     fmt.Sprintf("/zones/%d/records", zone.ID),
			RawQuery: options.QueryString(),
		}

		return c.client.Get(ctx, c.endpoint(uri.String()), handlers...)
	}

	err := c.do(ctx, req, &records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// CreateRecord creates a zone record using supplied attributes.
//
// See: http://www.luadns.com/api.html#create-a-record
func (c *Client) CreateRecord(ctx context.Context, zone *Zone, attrs *Record) (*Record, error) {
	var record Record

	req := func(ctx context.Context) ([]byte, error) {
		return c.client.Post(ctx, c.endpoint("/zones/%d/records", zone.ID), attrs)
	}

	err := c.do(ctx, req, &record)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

// GetRecord returns a zone record identified by `recordID`.
//
// See: http://www.luadns.com/api.html#get-a-record
func (c *Client) GetRecord(ctx context.Context, zone *Zone, recordID int64) (*Record, error) {
	var record Record

	req := func(ctx context.Context) ([]byte, error) {
		return c.client.Get(ctx, c.endpoint("/zones/%d/records/%d", zone.ID, recordID))
	}

	err := c.do(ctx, req, &record)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

// UpdateRecord updates a zone record identfied by `recordID` using supplied attributes.
//
// See: http://www.luadns.com/api.html#update-a-record
func (c *Client) UpdateRecord(ctx context.Context, zone *Zone, recordID int64, attrs *Record) (*Record, error) {
	var record Record

	req := func(ctx context.Context) ([]byte, error) {
		return c.client.Put(ctx, c.endpoint("/zones/%d/records/%d", zone.ID, recordID), attrs)
	}

	err := c.do(ctx, req, &record)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

// DeleteRecord deletes a zone record identfied by `recordID`.
//
// See: http://www.luadns.com/api.html#delete-a-record
func (c *Client) DeleteRecord(ctx context.Context, zone *Zone, recordID int64) (*Record, error) {
	var record Record

	req := func(ctx context.Context) ([]byte, error) {
		return c.client.Delete(ctx, c.endpoint("/zones/%d/records/%d", zone.ID, recordID))
	}

	err := c.do(ctx, req, &record)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

// CreateManyRecords creates multiple DNS records using supplied RRs.
//
// See: http://www.luadns.com/api.html#create-many-records
func (c *Client) CreateManyRecords(ctx context.Context, zone *Zone, recs []*RR) ([]*Record, error) {
	var records []*Record

	req := func(ctx context.Context) ([]byte, error) {
		return c.client.Post(ctx, c.endpoint("/zones/%d/records/create_many", zone.ID), recs)
	}

	err := c.do(ctx, req, &records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// UpdateManyRecords updates multiple DNS records using supplied RRs.
//
// For any (name, type) pairs in the input it will change zone records with
// supplied input records, other records remains unchanged.
//
// See: http://www.luadns.com/api.html#update-many-records
//
// Example:
//
//	;; Original records
//	example.com. 3600 IN A   1.1.1.1
//	example.com. 3600 IN A   2.2.2.2
//	example.com. 3600 IN TXT "hello world"
//
//	;; Input records
//	example.com. 3600 IN A   1.1.1.1
//	example.com. 3600 IN A   3.3.3.3
//
//	;; Result records
//	example.com. 3600 IN A   1.1.1.1
//	example.com. 3600 IN A   3.3.3.3
//	example.com. 3600 IN TXT "hello world"
func (c *Client) UpdateManyRecords(ctx context.Context, zone *Zone, recs []*RR) ([]*Record, error) {
	var records []*Record

	req := func(ctx context.Context) ([]byte, error) {
		return c.client.Patch(ctx, c.endpoint("/zones/%d/records", zone.ID), recs)
	}

	err := c.do(ctx, req, &records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// DeleteManyRecords deletes multiple zone records matching supplied RRs.
//
// Only records with exact match of input name, type, content and TTL are being deleted.
// For input RRs only name is required, the other fields are optional.
//
// See: http://www.luadns.com/api.html#delete-many-records
//
// Example:
//
//   - &RR{{Name: "example.com."}}		- matches all example.com. records
//   - &RR{{Name: "example.com.", Type: "TXT"}}	- matches all example.com. records of TXT type
func (c *Client) DeleteManyRecords(ctx context.Context, zone *Zone, recs []*RR) ([]*Record, error) {
	var records []*Record

	req := func(ctx context.Context) ([]byte, error) {
		return c.client.Post(ctx, c.endpoint("/zones/%d/records/delete_many", zone.ID), recs)
	}

	err := c.do(ctx, req, &records)
	if err != nil {
		return nil, err
	}

	return records, nil
}
