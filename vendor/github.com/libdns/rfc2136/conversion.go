package rfc2136

import (
	"fmt"
	"github.com/libdns/libdns"
	"github.com/miekg/dns"
	"strings"
	"time"
)

func recordToRR(rec libdns.Record, zone string) (dns.RR, error) {
	str := fmt.Sprintf(`%s %d IN %s %s`, rec.Name,
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
	return libdns.Record{
		Type:  typ,
		Name:  libdns.RelativeName(hdr.Name, zone),
		TTL:   time.Duration(hdr.Ttl) * time.Second,
		Value: strings.TrimPrefix(rr.String(), hdrStr),
	}
}
