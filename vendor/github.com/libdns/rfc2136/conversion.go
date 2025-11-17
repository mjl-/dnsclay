package rfc2136

import (
	"fmt"
	"strings"
	"time"

	"github.com/libdns/libdns"
	"github.com/miekg/dns"
)

func recordToRR(rec libdns.Record, zone string) (dns.RR, error) {
	libdnsRR := rec.RR()

	if !strings.HasSuffix(zone, ".") {
		zone += "."
	}

	data := libdnsRR.Data
	if libdnsRR.Type == "TXT" {
		// Chunk the data into 255 character segments
		runes := []rune(data)
		data = ""
		for i := 0; i < len(runes); i += 255 {
			data += fmt.Sprintf("%q ", string(runes[i:min(i+255, len(runes))]))
		}
	}

	zoneFile := fmt.Sprintf(
		"$ORIGIN %s\n%s %d IN %s %s",
		zone,
		libdnsRR.Name,
		int(libdnsRR.TTL.Seconds()),
		libdnsRR.Type,
		data,
	)

	dnsRR, err := dns.NewRR(zoneFile)
	if err != nil {
		return nil, err
	}

	return dnsRR, nil
}

func recordFromRR(rr dns.RR, zone string) libdns.RR {
	hdr := rr.Header()

	rec := libdns.RR{
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

	// Get the value from the record string representation.
	if txt, ok := rr.(*dns.TXT); ok {
		rec.Data = strings.Join(txt.Txt, "")
	} else {
		rec.Data = strings.TrimPrefix(rr.String(), hdrStr)
	}

	return rec
}
