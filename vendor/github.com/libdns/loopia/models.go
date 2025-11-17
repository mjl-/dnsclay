package loopia

import (
	"strings"
	"time"

	"github.com/libdns/libdns"
)

type loopiaRecord struct {
	ID       int64  `xmlrpc:"record_id"`
	TTL      int    `xmlrpc:"ttl"`
	Type     string `xmlrpc:"type"`
	RData    string `xmlrpc:"rdata"`
	Priority int    `xmlrpc:"priority"`
}

func (r *loopiaRecord) libdnsRecord(subDomain string) (libdns.Record, error) {
	return libdns.RR{
		Name: subDomain,
		Type: r.Type,
		Data: strings.Trim(r.RData, "\""),
		TTL:  time.Duration(r.TTL) * time.Second,
	}.Parse()
}

func (r *loopiaRecord) mustLibdnsRecord(subDomain string) libdns.Record {
	rr, err := r.libdnsRecord(subDomain)
	if err != nil {
		panic(err)
	}
	return rr
}

func toLoopiaRecord(r libdns.Record, id int64) (loopiaRecord, error) {
	rr := r.RR()

	out := loopiaRecord{
		Type:  rr.Type,
		TTL:   int(rr.TTL / time.Second),
		RData: rr.Data,
		ID:    id,
	}

	return out, nil
}

func mustToLoopiaRecord(r libdns.Record, id int64) loopiaRecord {
	lr, err := toLoopiaRecord(r, id)
	if err != nil {
		panic(err)
	}
	return lr
}

// Compare two libdns records as equal
// except TTL values, ovh can override them
func libdnsRecordEqual(r1 libdns.Record, r2 libdns.Record) bool {
	r1rr, r2rr := r1.RR(), r2.RR()
	return r1rr.Name == r2rr.Name && r1rr.Type == r2rr.Type && r1rr.Data == r2rr.Data
}

func libdnsEqualLoopia(r1 libdns.Record, r2 loopiaRecord) bool {
	r2libdns, err := r2.libdnsRecord(r1.RR().Name)
	if err != nil {
		return false
	}
	return libdnsRecordEqual(r1, r2libdns)
}
