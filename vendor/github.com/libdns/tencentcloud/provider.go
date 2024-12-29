package tencentcloud

import (
	"context"

	"github.com/libdns/libdns"
)

// Provider is a libdns provider for Tencent Cloud DNS
type Provider struct {
	// SecretId is the secret ID for Tencent Cloud DNS
	SecretId string
	// SecretKey is the secret key for Tencent Cloud DNS
	SecretKey string
}

// GetRecords gets the records for a zone
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	return p.describeRecordList(ctx, zone)
}

// AppendRecords appends records to a zone
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	for k, record := range records {
		if id, err := p.createRecord(ctx, zone, record); err != nil {
			return records, err
		} else {
			records[k].ID = id
		}
	}
	return records, nil
}

// SetRecords sets the records for a zone
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	for i, record := range records {
		if record.ID == "" {
			newRecord, err := p.AppendRecords(ctx, zone, []libdns.Record{record})
			if err != nil {
				return nil, err
			}
			records[i].ID = newRecord[0].ID
		} else {
			if err := p.modifyRecord(ctx, zone, record); err != nil {
				return nil, err
			}
		}
	}
	return records, nil
}

// DeleteRecords deletes the records for a zone
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	for _, record := range records {
		if err := p.deleteRecord(ctx, zone, record); err != nil {
			return nil, err
		}
	}
	return records, nil
}

// Interface guards

var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
