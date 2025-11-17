package ovh

import (
	"fmt"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

type ovhRecord struct {
	ID        int64  `json:"id,omitempty"`
	FieldType string `json:"fieldType"`
	SubDomain string `json:"subDomain"`
	TTL       int64  `json:"ttl"`
	Target    string `json:"target"`
}

// Convert ovh to libdns record
// ovh handle root domain by empty field
func (r ovhRecord) libdnsRecord() (libdns.Record, error) {
	name := r.SubDomain
	if name == "" {
		name = "@"
	}

	return libdns.RR{
		Name: name,
		TTL:  time.Duration(r.TTL) * time.Second,
		Type: r.FieldType,
		Data: strings.Trim(r.Target, "\""),
	}.Parse()
}

// Convert libdns to ovh record
// ovh handle root domain by empty field
// name is mandatory for CNAME on ovh
func toOvhRecord(record libdns.Record) (ovhRecord, error) {
	rr := record.RR()

	name := rr.Name
	if name == "@" {
		name = ""
	}

	if name == "" && rr.Type == "CNAME" {
		return ovhRecord{}, fmt.Errorf("name is mandatory for CNAME on ovh")
	}

	// OVH don't like unbalanced quotes
	if rr.Type == "TXT" {
		rr.Data = strings.ReplaceAll(rr.Data, `"`, "")
	}

	ttl := int64(rr.TTL.Seconds())
	if ttl < 60 {
		ttl = 60
	}

	return ovhRecord{
		FieldType: rr.Type,
		SubDomain: name,
		TTL:       ttl,
		Target:    rr.Data,
	}, nil
}

// Compare two libdns records as equal
// except TTL values, ovh can override them
func libdnsRecordEqual(r1 libdns.Record, r2 libdns.Record) bool {
	r1rr, r2rr := r1.RR(), r2.RR()
	return r1rr.Name == r2rr.Name && r1rr.Type == r2rr.Type && r1rr.Data == r2rr.Data
}

// Compare two libdns records as match
// ignore TTL values, ovh can override them
// their names must be exact
// their types must be exact and any not empty
// their data must be exact and any not empty
func libdnsRecordMatch(r1 libdns.Record, r2 libdns.Record) bool {
	r1rr, r2rr := r1.RR(), r2.RR()

	if r1rr.Name != r2rr.Name {
		return false
	}

	if r1rr.Type != r2rr.Type && r1rr.Type != "" && r2rr.Type != "" {
		return false
	}

	if r1rr.Data != r2rr.Data && r1rr.Data != "" && r2rr.Data != "" {
		return false
	}

	return true
}
