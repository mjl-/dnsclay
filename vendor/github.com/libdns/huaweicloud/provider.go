package huaweicloud

import (
	"context"
	"fmt"
	"sync"

	"github.com/libdns/libdns"
)

// Provider facilitates DNS record manipulation with Huawei Cloud
type Provider struct {
	// AccessKeyId is required by the Huawei Cloud API for authentication.
	AccessKeyId string `json:"access_key_id,omitempty"`
	// SecretAccessKey is required by the Huawei Cloud API for authentication.
	SecretAccessKey string `json:"secret_access_key,omitempty"`
	// RegionId is optional and defaults to "cn-south-1".
	RegionId string `json:"region_id,omitempty"`
	// once is used to ensure the client is initialized only once.
	once sync.Once
	//  client is the Huawei Cloud DNS client.
	client *Client
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	client := p.getClient()

	records, err := client.GetRecords(ctx, zone)
	if err != nil {
		return nil, err
	}

	var results []libdns.Record
	for _, record := range records {
		rec, err := record.libdnsRecord(zone)
		if err != nil {
			return nil, fmt.Errorf("parsing Huawei Cloud DNS record %+v: %v", record, err)
		}
		results = append(results, rec...)
	}

	return results, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
// NOTE: This implementation is NOT atomic.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	client := p.getClient()

	var results []libdns.Record
	for _, rec := range records {
		hwRec, err := hwRecord(zone, rec)
		if err != nil {
			return nil, fmt.Errorf("parsing libdns record %+v: %v", rec, err)
		}
		resp, err := client.AppendRecord(ctx, zone, hwRec)
		if err != nil {
			return nil, err
		}
		libdnsRecs, err := resp.libdnsRecord(zone)
		if err != nil {
			return nil, fmt.Errorf("parsing Huawei Cloud DNS record %+v: %v", resp, err)
		}
		results = append(results, libdnsRecs...)
	}

	return results, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
// NOTE: This implementation is NOT atomic.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	client := p.getClient()

	var results []libdns.Record
	for _, record := range records {
		rr := record.RR()
		id, err := client.GetRecordId(ctx, zone, rr.Name, rr.Type, rr.Data)
		if err != nil {
			// No existing record found, create a new one
			hwRec, err := hwRecord(zone, record)
			if err != nil {
				return nil, fmt.Errorf("parsing libdns record %+v: %v", record, err)
			}
			resp, err := client.AppendRecord(ctx, zone, hwRec)
			if err != nil {
				return nil, err
			}
			libdnsRecs, err := resp.libdnsRecord(zone)
			if err != nil {
				return nil, fmt.Errorf("parsing Huawei Cloud DNS record %+v: %v", resp, err)
			}
			results = append(results, libdnsRecs...)
		} else {
			// Existing record found, update it
			hwRec, err := hwRecord(zone, record)
			if err != nil {
				return nil, fmt.Errorf("parsing libdns record %+v: %v", record, err)
			}
			hwRec.Id = id
			hwRec.Ttl = int32(rr.TTL.Seconds())
			resp, err := client.UpdateRecord(ctx, zone, hwRec)
			if err != nil {
				return nil, err
			}
			libdnsRecs, err := resp.libdnsRecord(zone)
			if err != nil {
				return nil, fmt.Errorf("parsing Huawei Cloud DNS record %+v: %v", resp, err)
			}
			results = append(results, libdnsRecs...)
		}
	}

	return results, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
// NOTE: This implementation is NOT atomic.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	client := p.getClient()

	var results []libdns.Record
	for _, record := range records {
		rr := record.RR()
		id, err := client.GetRecordId(ctx, zone, rr.Name, rr.Type, rr.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to get record ID for %s: %v", rr.Name, err)
		}
		resp, err := client.DeleteRecord(ctx, zone, id)
		if err != nil {
			return nil, fmt.Errorf("failed to delete record %s: %v", rr.Name, err)
		}
		libdnsRecs, err := resp.libdnsRecord(zone)
		if err != nil {
			return nil, fmt.Errorf("parsing Huawei Cloud DNS record %+v: %v", resp, err)
		}
		results = append(results, libdnsRecs...)
	}

	return results, nil
}

// getClient initializes the client for the provider.
func (p *Provider) getClient() *Client {
	p.once.Do(func() {
		if p.AccessKeyId == "" || p.SecretAccessKey == "" {
			panic("huaweicloud: credentials missing")
		}
		p.client = NewClient(p.AccessKeyId, p.SecretAccessKey, p.RegionId)
	})
	return p.client
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
