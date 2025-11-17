// Package westcn implements a DNS record management client compatible
// with the libdns interfaces for west.cn.
package westcn

import (
	"context"
	"fmt"
	"sync"

	"github.com/libdns/libdns"
)

// Provider facilitates DNS record manipulation with west.cn.
type Provider struct {
	// Username is your username for west.cn, see https://www.west.cn/CustomerCenter/doc/apiv2.html#12u3001u8eabu4efdu9a8cu8bc10a3ca20id3d12u3001u8eabu4efdu9a8cu8bc13e203ca3e
	Username string `json:"username,omitempty"`
	// APIPassword is your API password for west.cn, see https://www.west.cn/CustomerCenter/doc/apiv2.html#12u3001u8eabu4efdu9a8cu8bc10a3ca20id3d12u3001u8eabu4efdu9a8cu8bc13e203ca3e
	APIPassword string `json:"api_password,omitempty"`
	// once is used to ensure the client is initialized only once.
	once sync.Once
	//  client is the west.cn API client.
	client *Client
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	client := p.getClient()

	records, err := client.GetRecords(ctx, zone)
	if err != nil {
		return nil, err
	}

	results := make([]libdns.Record, 0, len(records))
	for _, rec := range records {
		libdnsRec, err := rec.libdnsRecord(zone)
		if err != nil {
			return nil, fmt.Errorf("parsing West.cn DNS record %+v: %v", rec, err)
		}
		results = append(results, libdnsRec)
	}

	return results, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
// NOTE: This implementation is NOT atomic.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	client := p.getClient()

	var results []libdns.Record
	for _, rec := range records {
		westcnRec, err := westcnRecord(zone, rec)
		if err != nil {
			return nil, fmt.Errorf("parsing libdns record %+v: %v", rec, err)
		}
		if _, err = client.AppendRecord(ctx, zone, westcnRec); err != nil {
			return nil, err
		}
		libdnsRec, err := westcnRec.libdnsRecord(zone)
		if err != nil {
			return nil, fmt.Errorf("parsing West.cn DNS record %+v: %v", rec, err)
		}
		results = append(results, libdnsRec)
	}

	return results, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
// NOTE: This implementation is NOT atomic.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	client := p.getClient()

	var results []libdns.Record
	for _, rec := range records {
		// West.cn does not support updating records, so we need to delete the existing record first
		rr := rec.RR()
		id, err := p.getRecordId(ctx, zone, rr.Name, rr.Type, rr.Data)
		if err == nil {
			if err = client.DeleteRecord(ctx, zone, id); err != nil {
				return nil, err
			}
		}
		// Now we can add the record
		westcnRec, err := westcnRecord(zone, rec)
		if err != nil {
			return nil, fmt.Errorf("parsing libdns record %+v: %v", rec, err)
		}
		_, err = client.AppendRecord(ctx, zone, westcnRec)
		if err != nil {
			return nil, err
		}
		// Add the record to the results
		libdnsRec, err := westcnRec.libdnsRecord(zone)
		if err != nil {
			return nil, fmt.Errorf("parsing West.cn DNS record %+v: %v", rec, err)
		}
		results = append(results, libdnsRec)
	}

	return results, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
// NOTE: This implementation is NOT atomic.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	client := p.getClient()

	for _, record := range records {
		rr := record.RR()
		id, err := p.getRecordId(ctx, zone, rr.Name, rr.Type, rr.Data)
		if err != nil {
			return nil, err
		}
		if err = client.DeleteRecord(ctx, zone, id); err != nil {
			return nil, err
		}
	}

	return records, nil
}

func (p *Provider) getRecordId(ctx context.Context, zone, recName, recType string, recVal ...string) (int, error) {
	client := p.getClient()

	records, err := client.GetRecords(ctx, zone)
	if err != nil {
		return 0, err
	}

	for _, rec := range records {
		if recName == rec.Item && recType == rec.Type {
			if len(recVal) > 0 && recVal[0] != "" && rec.Value != recVal[0] {
				continue
			}
			return rec.ID, nil
		}
	}

	return 0, fmt.Errorf("record %q not found", recName)
}

func (p *Provider) getClient() *Client {
	p.once.Do(func() {
		if p.Username == "" || p.APIPassword == "" {
			panic("westcn: credentials missing")
		}
		p.client = NewClient(p.Username, p.APIPassword)
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
