// Package libdnstemplate implements a DNS record management client compatible
// with the libdns interfaces for Domainnameshop.
package domainnameshop

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/libdns/libdns"
)

// Provider facilitates DNS record manipulation with Domainnameshop
// https://api.domeneshop.no/docs/#section/Authentication
type Provider struct {
	APIToken  string `json:"api_token"`
	APISecret string `json:"api_secret"`

	zones   map[string]dsZone
	zonesMu sync.Mutex

	knownRecords   map[string][]dsDNSRecord
	knownRecordsMu sync.Mutex
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	zoneinfo, err := p.getAllDomainRecords(ctx, p.APIToken, p.APISecret, zone)
	if err != nil {
		return nil, err
	}
	recs := make([]libdns.Record, 0, len(zoneinfo))
	for _, rec := range zoneinfo {
		libdnsRec, err := rec.libdnsRecord()
		if err != nil {
			return nil, fmt.Errorf("parsing Domainnameshop DNS record %+v: %v", rec, err)
		}
		recs = append(recs, libdnsRec)
	}
	log.Printf("GOT RECORDS: %#v", recs)

	return recs, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {

	var created []libdns.Record
	for _, rec := range records {
		dsrr, err := libdnsRecordTodsDNSRecord(rec)
		if err != nil {
			return nil, err
		}

		result, err := p.createDNSRecord(ctx, p.APIToken, p.APISecret, zone, dsrr)
		if err != nil {
			return nil, err
		}

		libdnsRec, err := result.libdnsRecord()
		if err != nil {
			return nil, fmt.Errorf("parsing Domainnameshop DNS record %+v: %v", rec, err)
		}

		created = append(created, libdnsRec)
	}

	return created, nil
}

// DeleteRecords deletes the records from the zone.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {

	for _, record := range records {
		dsrr, converr := libdnsRecordTodsDNSRecord(record)
		if converr != nil {
			return nil, converr
		}

		err := p.deleteDNSRecord(ctx, p.APIToken, p.APISecret, zone, dsrr)
		if err != nil {
			return nil, err
		}
	}

	return records, nil
}

// SetRecords sets the records in the zone, either by updating existing records
// or creating new ones. It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	var appendedRecords []dsDNSRecord
	for _, record := range records {
		dsrr, converr := libdnsRecordTodsDNSRecord(record)
		if converr != nil {
			return nil, converr
		}

		newRecord, err := p.createOrUpdateDNSRecord(ctx, p.APIToken, p.APISecret, zone, dsrr)
		if err != nil {
			return nil, err
		}
		appendedRecords = append(appendedRecords, newRecord)
	}

	recs := make([]libdns.Record, 0, len(appendedRecords))
	for _, rec := range appendedRecords {
		libdnsRec, err := rec.libdnsRecord()
		if err != nil {
			return nil, fmt.Errorf("parsing Domainnameshop DNS record %+v: %v", rec, err)
		}
		recs = append(recs, libdnsRec)
	}
	log.Printf("GOT RECORDS: %#v", recs)

	return recs, nil
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
