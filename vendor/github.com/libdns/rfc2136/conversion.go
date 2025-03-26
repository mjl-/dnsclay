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
	rrType := dns.StringToType[rec.Type]
	rrConstructor := dns.TypeToRR[rrType]
	var rr dns.RR
	if rrConstructor == nil {
		rr = new(dns.RFC3597)
	} else {
		rr = rrConstructor()
	}

	// Create a zone file line representing the record. We're using reflection
	// so that we can automatically support any new RR types that define these
	// fields.
	name := rec.Name
	if name == "" {
		name = "@"
	}

	zoneLine := fmt.Sprintf("%s %d IN %s ", name, int(rec.TTL.Seconds()), rec.Type)

	priority := reflect.ValueOf(rr).Elem().FieldByName("Priority")
	if !priority.IsValid() {
		priority = reflect.ValueOf(rr).Elem().FieldByName("Preference")
	}

	if priority.IsValid() {
		zoneLine += fmt.Sprintf("%d ", rec.Priority)
	}

	weight := reflect.ValueOf(rr).Elem().FieldByName("Weight")
	if weight.IsValid() {
		zoneLine += fmt.Sprintf("%d ", rec.Weight)
	}

	target := reflect.ValueOf(rr).Elem().FieldByName("Target")
	if target.IsValid() {
		zoneLine += fmt.Sprintf("%s ", rec.Target)
	}

	if rec.Value != "" {
		zoneLine += rec.Value
	}
	zoneLine = strings.TrimSuffix(zoneLine, " ") + "\n"

	zp := dns.NewZoneParser(strings.NewReader(zoneLine), zone, "")
	rr, _ = zp.Next()
	return rr, zp.Err()
}

func recordFromRR(rr dns.RR, zone string) libdns.Record {
	hdr := rr.Header()

	rec := libdns.Record{
		Name: libdns.RelativeName(hdr.Name, zone),
		TTL:  time.Duration(hdr.Ttl) * time.Second,
	}

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

	rec.Type = typ

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
		hdrStr += fmt.Sprintf("%d ", priority)
	}

	weight := reflect.ValueOf(rr).Elem().FieldByName("Weight")
	if weight.IsValid() {
		weight := weight.Uint()
		rec.Weight = uint(weight)
		hdrStr += fmt.Sprintf("%d ", weight)
	}

	target := reflect.ValueOf(rr).Elem().FieldByName("Target")
	if target.IsValid() && typ != "SRV" {
		rec.Target = target.String()
		hdrStr += fmt.Sprintf("%s ", rec.Target)
	}

	// Get the value from the record string representation.
	rec.Value = strings.TrimPrefix(rr.String(), hdrStr)

	return rec
}
