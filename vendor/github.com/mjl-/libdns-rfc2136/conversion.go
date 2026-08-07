package rfc2136

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/libdns/libdns"
	"github.com/miekg/dns"
)

func recordToRR(rec libdns.Record, zone string) (dns.RR, error) {
	name := rec.Name
	if name == "" {
		name = "@"
	}

	str := fmt.Sprintf("%s %d IN %s %s", name,
		int(rec.TTL.Seconds()), rec.Type, rec.Value)
	zp := dns.NewZoneParser(strings.NewReader(str), zone, "")
	rr, _ := zp.Next()
	return rr, zp.Err()
}

func recordFromRR(rr dns.RR, zone string) libdns.Record {
	hdr := rr.Header()

	// The record value is the full record string representation with the header string
	// prefix stripped. Package dns represents private-use and unknown records as
	// RFC3597 records. When those are formatted as string, their header prefix has a
	// different form than when the header is formatted separately, so account for
	// that.
	hdrStr := hdr.String()
	typ := dns.TypeToString[hdr.Rrtype]
	if _, ok := rr.(*dns.RFC3597); ok {
		name := strings.SplitN(hdrStr, "\t", 2)[0]
		typ = fmt.Sprintf("TYPE%d", hdr.Rrtype)
		hdrStr = fmt.Sprintf("%s\t%d\tCLASS%d\t%s\t", name, hdr.Ttl, hdr.Class, typ)
	}

	rec := libdns.Record{
		Type:  typ,
		Name:  libdns.RelativeName(hdr.Name, zone),
		TTL:   time.Duration(hdr.Ttl) * time.Second,
		Value: strings.TrimPrefix(rr.String(), hdrStr),
	}

	// Parse priority, weight, and target from the record value. We're using
	// reflection so that we can automatically support any new RR types that
	// define these fields.
	priority := reflect.ValueOf(rr).Elem().FieldByName("Priority")
	if !priority.IsValid() {
		priority = reflect.ValueOf(rr).Elem().FieldByName("Preference")
	}
	if priority.IsValid() {
		priority := priority.Uint()
		rec.Priority = uint(priority)
	}

	weight := reflect.ValueOf(rr).Elem().FieldByName("Weight")
	if weight.IsValid() {
		weight := weight.Uint()
		rec.Weight = uint(weight)
	}

	target := reflect.ValueOf(rr).Elem().FieldByName("Target")
	if target.IsValid() && typ != "SRV" {
		rec.Target = target.String()
	}

	return rec
}
