package luadns

import (
	"strings"
	"time"

	"github.com/libdns/libdns"
	"github.com/luadns/luadns-go"
)

// unFQDN trims any trailing "." from fqdn name.
func unFQDN(fqdn string) string {
	return strings.TrimSuffix(fqdn, ".")
}

// toFQDN converts name in FQDN format.
func toFQDN(name string) string {
	if !strings.HasSuffix(name, ".") {
		return name + "."
	}
	return name
}

// newLibRecord builds a libdns Record from the provider Record.
func newLibRecord(in *luadns.Record, zone string) (libdns.Record, error) {
	rr := libdns.RR{
		Type: in.Type,
		Name: libdns.RelativeName(in.Name, zone),
		Data: in.Content,
		TTL:  time.Duration(in.TTL) * time.Second,
	}

	// Convert RR type to concrete struct.
	r, err := rr.Parse()
	if err != nil {
		return nil, err
	}

	// Set ProviderData for concrete struct types.
	switch r := r.(type) {
	case libdns.Address:
		r.ProviderData = in.ID
		return r, nil
	case libdns.CAA:
		r.ProviderData = in.ID
		return r, nil
	case libdns.CNAME:
		r.ProviderData = in.ID
		return r, nil
	case libdns.ServiceBinding:
		r.ProviderData = in.ID
		return r, nil
	case libdns.MX:
		r.ProviderData = in.ID
		return r, nil
	case libdns.NS:
		r.ProviderData = in.ID
		return r, nil
	case libdns.SRV:
		r.ProviderData = in.ID
		return r, nil
	case libdns.TXT:
		r.ProviderData = in.ID
		return r, nil
	default:
	}

	return r, nil
}

// toLibdnsRecord converts LuaDNS records to libdns records.
func toLibdnsRecord(records []*luadns.Record, zone string) ([]libdns.Record, error) {
	result := []libdns.Record{}
	for _, t := range records {
		r, err := newLibRecord(t, zone)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

// toLibdnsZone converts LuaDNS zones to libdns zones.
func toLibdnsZone(zones []*luadns.Zone) ([]libdns.Zone, error) {
	result := []libdns.Zone{}
	for _, z := range zones {
		result = append(result, libdns.Zone{Name: z.Name})
	}
	return result, nil
}

// toLuaRR converts libdns records to LuaDNS records.
func toLuaRR(records []libdns.Record, zone string) []*luadns.RR {
	result := []*luadns.RR{}
	for _, t := range records {
		rr := t.RR()
		r := &luadns.RR{
			Type:    rr.Type,
			Name:    toFQDN(libdns.AbsoluteName(rr.Name, zone)),
			Content: rr.Data,
			TTL:     uint32(rr.TTL.Seconds()),
		}
		result = append(result, r)
	}
	return result
}
