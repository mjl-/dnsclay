package directadmin

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

type daZone struct {
	Records                  []daRecord `json:"records"`
	Dnssec                   string     `json:"dnssec,omitempty"`
	UserDnssecControl        string     `json:"user_dnssec_control,omitempty"`
	DNSNs                    string     `json:"dns_ns,omitempty"`
	DNSPtr                   string     `json:"dns_ptr,omitempty"`
	DNSSpf                   string     `json:"dns_spf,omitempty"`
	DNSTTL                   string     `json:"dns_ttl,omitempty"`
	DNSAffectPointersDefault string     `json:"DNS_AFFECT_POINTERS_DEFAULT,omitempty"`
	DNSTLSa                  string     `json:"dns_tlsa,omitempty"`
	DNSCaa                   string     `json:"dns_caa,omitempty"`
	AllowDNSUnderscore       string     `json:"allow_dns_underscore,omitempty"`
	FullMxRecords            string     `json:"full_mx_records,omitempty"`
	DefaultTTL               string     `json:"default_ttl,omitempty"`
	AllowTTLOverride         string     `json:"allow_ttl_override,omitempty"`
	TTLIsOverridden          string     `json:"ttl_is_overridden,omitempty"`
	TTL                      string     `json:"ttl,omitempty"`
	TTLValue                 string     `json:"ttl_value,omitempty"`
}

type daRecord struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	Combined string `json:"combined"`
	TTL      string `json:"ttl,omitempty"`
}

var ErrUnsupported = errors.New("unsupported record type")

func (r daRecord) libdnsRecord(zone string) (libdns.Record, error) {
	var ttl time.Duration
	if len(r.TTL) > 0 {
		ttlVal, err := strconv.Atoi(r.TTL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse TTL for %v: %v", r.Name, err)
		}
		ttl = time.Duration(ttlVal) * time.Second
	}

	rr := libdns.RR{
		Name: r.Name,
		Type: r.Type,
		TTL:  ttl,
		Data: r.Value,
	}

	switch r.Type {
	case "MX":
		splits := strings.Split(r.Value, " ")
		if len(splits) >= 2 {
			priority, err := strconv.Atoi(splits[0])
			if err == nil {
				return &libdns.MX{
					Name:       r.Name,
					TTL:        ttl,
					Preference: uint16(priority),
					Target:     fmt.Sprintf("%v.%v", splits[1], zone),
				}, nil
			}
		}
		return &rr, nil
	case "SRV":
		// Parse SRV record format: "priority weight port target"
		splits := strings.Split(r.Value, " ")
		if len(splits) >= 4 {
			priority, err1 := strconv.Atoi(splits[0])
			weight, err2 := strconv.Atoi(splits[1])
			port, err3 := strconv.Atoi(splits[2])
			if err1 == nil && err2 == nil && err3 == nil {
				// Extract service and transport from the name
				// SRV names are typically in format: _service._transport.name
				nameParts := strings.Split(r.Name, ".")
				if len(nameParts) >= 3 && strings.HasPrefix(nameParts[0], "_") && strings.HasPrefix(nameParts[1], "_") {
					service := strings.TrimPrefix(nameParts[0], "_")
					transport := strings.TrimPrefix(nameParts[1], "_")
					// Reconstruct the base name (without service and transport prefixes)
					baseName := strings.Join(nameParts[2:], ".")

					return &libdns.SRV{
						Service:   service,
						Transport: transport,
						Name:      baseName,
						TTL:       ttl,
						Priority:  uint16(priority),
						Weight:    uint16(weight),
						Port:      uint16(port),
						Target:    splits[3],
					}, nil
				}
			}
		}
		// Fall back to generic RR if parsing fails
		return &rr, nil
	case "URI":
		return nil, ErrUnsupported
	default:
		return &rr, nil
	}
}

type daResponse struct {
	Error   string `json:"error,omitempty"`
	Success string `json:"success,omitempty"`
	Result  string `json:"result,omitempty"`
}

type daDomainList []string
