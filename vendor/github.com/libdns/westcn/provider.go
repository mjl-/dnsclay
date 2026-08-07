// Package westcn implements a DNS record management client compatible
// with the libdns interfaces for west.cn.
package westcn

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

// Provider facilitates DNS record manipulation with west.cn.
type Provider struct {
	// Username is your username for west.cn, see https://www.west.cn/CustomerCenter/doc/apiv2.html#12u3001u8eabu4efdu9a8cu8bc10a3ca20id3d12u3001u8eabu4efdu9a8cu8bc13e203ca3e
	Username string `json:"username,omitempty"`
	// APIPassword is your API password for west.cn, see https://www.west.cn/CustomerCenter/doc/apiv2.html#12u3001u8eabu4efdu9a8cu8bc10a3ca20id3d12u3001u8eabu4efdu9a8cu8bc13e203ca3e
	APIPassword string `json:"api_password,omitempty"`
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}

	items, err := client.GetRecords(ctx, strings.TrimSuffix(zone, "."))
	if err != nil {
		return nil, err
	}

	records := make([]libdns.Record, len(items))
	for i, item := range items {
		records[i] = libdns.Record{
			ID:       strconv.Itoa(item.ID),
			Type:     item.Type,
			Name:     item.Host,
			Value:    item.Value,
			TTL:      time.Duration(item.TTL) * time.Second,
			Priority: uint(item.Level),
		}
	}

	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}

	for i, record := range records {
		if record.TTL <= 0 {
			record.TTL = 10 * time.Minute
		}
		if record.Priority <= 0 {
			record.Priority = 10
		}
		id, err := client.AddRecord(ctx, Record{
			Domain: strings.TrimSuffix(zone, "."),
			Host:   record.Name,
			Type:   record.Type,
			Value:  record.Value,
			TTL:    int(record.TTL.Seconds()),
			Level:  int(record.Priority),
		})
		if err != nil {
			return nil, err
		}
		records[i].ID = strconv.Itoa(id)
	}

	return records, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}

	for i, record := range records {
		if record.TTL <= 0 {
			record.TTL = 10 * time.Minute
		}
		if record.Priority <= 0 {
			record.Priority = 10
		}
		if record.ID == "" {
			id, err := p.getRecordId(ctx, zone, record.Name, record.Type, record.Value)
			if err == nil {
				id2, _ := strconv.Atoi(id)
				if err = client.DeleteRecord(ctx, strings.TrimSuffix(zone, "."), id2); err != nil {
					return nil, err
				}
			}
		} else {
			id, err := strconv.Atoi(record.ID)
			if err != nil {
				return nil, fmt.Errorf("invalid record ID %q: %w", record.ID, err)
			}
			if err = client.DeleteRecord(ctx, strings.TrimSuffix(zone, "."), id); err != nil {
				return nil, err
			}
		}
		records[i].ID = ""
		id, err := client.AddRecord(ctx, Record{
			Domain: strings.TrimSuffix(zone, "."),
			Host:   record.Name,
			Type:   record.Type,
			Value:  record.Value,
			TTL:    int(record.TTL.Seconds()),
			Level:  10,
		})
		if err != nil {
			return nil, err
		}
		records[i].ID = strconv.Itoa(id)
	}

	return records, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		if record.ID == "" {
			id, err := p.getRecordId(ctx, zone, record.Name, record.Type, record.Value)
			if err != nil {
				return nil, err
			}
			record.ID = id
		}
		id, err := strconv.Atoi(record.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid record ID %q: %w", record.ID, err)
		}
		if err = client.DeleteRecord(ctx, strings.TrimSuffix(zone, "."), id); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (p *Provider) getRecordId(ctx context.Context, zone, recName, recType string, recVal ...string) (string, error) {
	records, err := p.GetRecords(ctx, zone)
	if err != nil {
		return "", err
	}

	for _, record := range records {
		if recName == record.Name && recType == record.Type {
			if len(recVal) > 0 && recVal[0] != "" && record.Value != recVal[0] {
				continue
			}
			return record.ID, nil
		}
	}

	return "", fmt.Errorf("record %q not found", recName)
}

func (p *Provider) getClient() (*Client, error) {
	return NewClient(p.Username, p.APIPassword)
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
