package timeweb

import (
	"context"
	"fmt"
	"github.com/libdns/libdns"
	"net/http"
)

type Provider struct {
	ApiURL   string
	ApiToken string
}

func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	reqURL := fmt.Sprintf("%s/domains/%s/dns-records", p.ApiURL, zone)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	var result RecordsResponse
	err = p.doAPIRequest(req, &result)

	recs := make([]libdns.Record, 0, len(result.DNSRecords))
	for _, r := range result.DNSRecords {
		recs = append(recs, r.libDNSRecord(zone))
	}

	return recs, err
}

func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	var created []libdns.Record
	for _, record := range records {
		result, err := p.createRecord(ctx, zone, record)
		if err != nil {
			return nil, err
		}
		created = append(created, result.libDNSRecord(zone))
	}

	return created, nil
}

func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	var deleted []libdns.Record
	for _, record := range records {
		err := p.deleteRecord(ctx, zone, record)
		if err != nil {
			return nil, err
		}
		deleted = append(deleted, record)
	}

	return deleted, nil
}

func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	zoneRecords, err := p.GetRecords(ctx, zone)
	if err != nil {
		return nil, err
	}

	var results []libdns.Record
	var resultErr error
	for _, libRecord := range records {
		exists := isRecordExists(zoneRecords, libRecord)
		if exists {
			record, err := p.updateRecord(ctx, zone, libRecord)
			if err != nil {
				resultErr = err
			}
			results = append(results, record.libDNSRecord(zone))
		} else {
			record, err := p.createRecord(ctx, zone, libRecord)
			if err != nil {
				resultErr = err
			}
			results = append(results, record.libDNSRecord(zone))
		}
	}

	return results, resultErr
}

var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
