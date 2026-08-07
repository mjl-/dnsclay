package timeweb

import (
	"fmt"
	"github.com/libdns/libdns"
)

type RecordResponse struct {
	Data struct {
		Priority uint   `json:"priority,omitempty"`
		Value    string `json:"value"`
	} `json:"data"`
	ID   uint   `json:"id"`
	Type string `json:"type"`
	Fqdn string `json:"fqdn"`
}

type RecordsResponse struct {
	Meta struct {
		Total int `json:"total"`
	} `json:"meta"`
	DNSRecords []RecordResponse `json:"dns_records"`
}

type Record struct {
	Subdomain string `json:"subdomain"`
	Type      string `json:"type"`
	Value     string `json:"value"`
}

type SavedRecord struct {
	DNSRecord struct {
		ID   uint   `json:"id"`
		Type string `json:"type"`
		Data struct {
			Priority  uint   `json:"priority,omitempty"`
			Value     string `json:"value"`
			Subdomain string `json:"subdomain"`
		}
	} `json:"dns_record"`
}

func (r *SavedRecord) libDNSRecord(zone string) libdns.Record {
	return libdns.Record{
		ID:    fmt.Sprintf("%d", r.DNSRecord.ID),
		Name:  libdns.RelativeName(r.DNSRecord.Data.Subdomain, zone),
		Type:  r.DNSRecord.Type,
		Value: r.DNSRecord.Data.Value,
	}
}

func (r *RecordResponse) libDNSRecord(zone string) libdns.Record {
	return libdns.Record{
		ID:       fmt.Sprintf("%d", r.ID),
		Name:     libdns.RelativeName(r.Fqdn, zone),
		Type:     r.Type,
		Value:    r.Data.Value,
		Priority: r.Data.Priority,
	}
}

func libdnsToRecord(r libdns.Record) Record {
	return Record{
		Type:      r.Type,
		Value:     r.Value,
		Subdomain: r.Name,
	}
}
