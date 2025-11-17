// Helper functions for record conversion
package autodns

import (
	"fmt"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

// ToLibDNS converts an AutoDNS ZoneRecord to a libdns Record
func ToLibDNS(zr ZoneRecord) (libdns.Record, error) {
	switch strings.ToUpper(zr.Type) {
	case "A", "AAAA":
		ip, err := netip.ParseAddr(zr.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid %s record value: %s", zr.Type, zr.Value)
		}
		return &libdns.Address{
			Name: zr.Name,
			IP:   ip,
			TTL:  time.Duration(zr.TTL) * time.Second,
		}, nil
	case "CNAME":
		return &libdns.CNAME{
			Name:   zr.Name,
			Target: zr.Value,
			TTL:    time.Duration(zr.TTL) * time.Second,
		}, nil
	case "TXT":
		return &libdns.TXT{
			Name: zr.Name,
			Text: zr.Value,
			TTL:  time.Duration(zr.TTL) * time.Second,
		}, nil
	case "MX":
		// MX record format: "priority target"
		parts := strings.Fields(zr.Value)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid MX record format: %s", zr.Value)
		}
		preference, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid MX preference: %s", parts[0])
		}
		return &libdns.MX{
			Name:       zr.Name,
			Preference: uint16(preference),
			Target:     parts[1],
			TTL:        time.Duration(zr.TTL) * time.Second,
		}, nil
	case "NS":
		return &libdns.NS{
			Name:   zr.Name,
			Target: zr.Value,
			TTL:    time.Duration(zr.TTL) * time.Second,
		}, nil
	case "SRV":
		// SRV record format: "priority weight port target"
		parts := strings.Fields(zr.Value)
		if len(parts) != 4 {
			return nil, fmt.Errorf("invalid SRV record format: %s", zr.Value)
		}
		priority, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid SRV priority: %s", parts[0])
		}
		weight, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid SRV weight: %s", parts[1])
		}
		port, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid SRV port: %s", parts[2])
		}
		return &libdns.SRV{
			Name:     zr.Name,
			Priority: uint16(priority),
			Weight:   uint16(weight),
			Port:     uint16(port),
			Target:   parts[3],
			TTL:      time.Duration(zr.TTL) * time.Second,
		}, nil
	default:
		// For unsupported record types, use the generic RR type
		return &libdns.RR{
			Name: zr.Name,
			Type: zr.Type,
			Data: zr.Value,
			TTL:  time.Duration(zr.TTL) * time.Second,
		}, nil
	}
}

// ToAutoDNS converts a libdns Record to an AutoDNS ZoneRecord
func ToAutoDNS(record libdns.Record) (ZoneRecord, error) {
	switch r := record.(type) {
	case *libdns.Address:
		recordType := "A"
		if r.IP.Is6() {
			recordType = "AAAA"
		}
		return ZoneRecord{
			Name:  r.Name,
			Type:  recordType,
			Value: r.IP.String(),
			TTL:   int(r.TTL.Seconds()),
		}, nil
	case *libdns.CNAME:
		return ZoneRecord{
			Name:  r.Name,
			Type:  "CNAME",
			Value: r.Target,
			TTL:   int(r.TTL.Seconds()),
		}, nil
	case *libdns.TXT:
		return ZoneRecord{
			Name:  r.Name,
			Type:  "TXT",
			Value: r.Text,
			TTL:   int(r.TTL.Seconds()),
		}, nil
	case *libdns.MX:
		return ZoneRecord{
			Name:  r.Name,
			Type:  "MX",
			Value: fmt.Sprintf("%d %s", r.Preference, r.Target),
			TTL:   int(r.TTL.Seconds()),
		}, nil
	case *libdns.NS:
		return ZoneRecord{
			Name:  r.Name,
			Type:  "NS",
			Value: r.Target,
			TTL:   int(r.TTL.Seconds()),
		}, nil
	case *libdns.SRV:
		return ZoneRecord{
			Name:  r.Name,
			Type:  "SRV",
			Value: fmt.Sprintf("%d %d %d %s", r.Priority, r.Weight, r.Port, r.Target),
			TTL:   int(r.TTL.Seconds()),
		}, nil
	case *libdns.RR:
		return ZoneRecord{
			Name:  r.Name,
			Type:  r.Type,
			Value: r.Data,
			TTL:   int(r.TTL.Seconds()),
		}, nil
	default:
		// Fallback: try to get the RR representation
		rr := record.RR()
		return ZoneRecord{
			Name:  rr.Name,
			Type:  rr.Type,
			Value: rr.Data,
			TTL:   int(rr.TTL.Seconds()),
		}, nil
	}
}
