package netlify

import (
	"fmt"
	"net/netip"
	"time"

	"github.com/libdns/libdns"
	"github.com/netlify/open-api/v2/go/models"
)

type netlifyZone struct {
	*models.DNSZone
}

type netlifyDNSRecord struct {
	*models.DNSRecord
}

func (r netlifyDNSRecord) libdnsRecord(zone string) (libdns.Record, error) {
	name := libdns.RelativeName(r.Hostname, zone)
	ttl := time.Duration(r.TTL) * time.Second
	switch r.Type {
	case "A", "AAAA":
		addr, err := netip.ParseAddr(r.Value)
		if err != nil {
			return libdns.Address{}, fmt.Errorf("invalid IP address %q: %v", r.Value, err)
		}
		return libdns.Address{
			Name: name,
			TTL:  ttl,
			IP:   addr,
		}, nil
	case "CNAME":
		return libdns.CNAME{
			Name:   name,
			TTL:    ttl,
			Target: r.Value,
		}, nil
	case "MX":
		return libdns.MX{
			Name:       name,
			TTL:        ttl,
			Preference: uint16(r.Priority),
			Target:     r.Value,
		}, nil
	case "NS":
		return libdns.NS{
			Name:   name,
			TTL:    ttl,
			Target: r.Value,
		}, nil
	case "TXT":
		return libdns.TXT{
			Name: name,
			TTL:  ttl,
			Text: r.Value,
		}, nil
	default:
		return libdns.RR{
			Name: name,
			TTL:  ttl,
			Type: r.Type,
			Data: r.Value,
		}.Parse()
	}
}

func netlifyRecord(r libdns.Record) netlifyDNSRecord {
	rr := r.RR()

	netlifyRec := netlifyDNSRecord{
		&models.DNSRecord{
			Type:     rr.Type,
			Hostname: rr.Name,
			TTL:      int64(rr.TTL),
			Value:    rr.Data,
		},
	}

	switch rec := r.(type) {
	case libdns.MX:
		netlifyRec.Priority = int64(rec.Preference)
	case libdns.NS:
		netlifyRec.Value = rec.Target
	case libdns.CNAME:
		netlifyRec.Value = rec.Target
	case libdns.TXT:
		netlifyRec.Value = rec.Text
	}

	return netlifyRec
}

type netlifyDNSDeleteError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}
