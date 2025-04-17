package alidns

import (
	"context"

	"github.com/libdns/libdns"
)

// Provider implements the libdns interfaces for Alicloud.
type Provider struct {
	client mClient
	// The API Key ID Required by Aliyun's for accessing the Aliyun's API
	AccKeyID string `json:"access_key_id"`
	// The API Key Secret Required by Aliyun's for accessing the Aliyun's API
	AccKeySecret string `json:"access_key_secret"`
	// Optional for identifing the region of the Aliyun's Service,The default is zh-hangzhou
	RegionID string `json:"region_id,omitempty"`
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	var rls []libdns.Record
	for _, rec := range recs {
		ar := alidnsRecordWithZone(rec, zone)
		rid, err := p.addDomainRecord(ctx, ar)
		if err != nil {
			return nil, err
		}
		ar.RecID = rid
		rls = append(rls, ar.LibdnsRecord())
	}
	return rls, nil
}

// DeleteRecords deletes the records from the zone. If a record does not have an ID,
// it will be looked up. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	var rls []libdns.Record
	for _, rec := range recs {
		ar := alidnsRecordWithZone(rec, zone)
		_, err := p.delDomainRecord(ctx, ar)
		if err != nil {
			return nil, err
		}
		rls = append(rls, ar.LibdnsRecord())
	}
	return rls, nil
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	var rls []libdns.Record
	recs, err := p.queryDomainRecords(ctx, zone)
	if err != nil {
		return nil, err
	}
	for _, rec := range recs {
		rls = append(rls, rec.LibdnsRecord())
	}
	return rls, nil
}

// SetRecords sets the records in the zone, either by updating existing records
// or creating new ones. It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	var rls []libdns.Record
	var err error
	for _, rec := range recs {
		ar := alidnsRecordWithZone(rec, zone)
		if ar.RecID == "" {
			ar.RecID, err = p.addDomainRecord(ctx, ar)
			if err != nil {
				return nil, err
			}
		} else {
			_, err = p.setDomainRecord(ctx, ar)
			if err != nil {
				return nil, err
			}
		}
		rls = append(rls, ar.LibdnsRecord())
	}
	return rls, nil
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
