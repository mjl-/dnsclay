package libdnsspaceship

import (
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

// toLibdnsRR converts a spaceshipRecordUnion (API) to a libdns.Record
func (p *Provider) toLibdnsRR(sr spaceshipRecordUnion, zone string) libdns.Record {
	// normalize name relative to zone
	name := strings.TrimSuffix(sr.Name, "."+zone)
	name = strings.TrimSuffix(name, ".")
	if name == zone || sr.Name == zone {
		name = ""
	}
	ttl := time.Duration(sr.TTL) * time.Second

	switch strings.ToUpper(sr.Type) {
	case "A", "AAAA":
		if sr.Address != "" {
			if ip, err := netip.ParseAddr(sr.Address); err == nil {
				return libdns.Address{Name: name, TTL: ttl, IP: ip, ProviderData: sr}
			}
		}
	case "TXT":
		return libdns.TXT{Name: name, TTL: ttl, Text: sr.Value, ProviderData: sr}
	case "CNAME":
		return libdns.CNAME{Name: name, TTL: ttl, Target: sr.Cname, ProviderData: sr}
	case "MX":
		return libdns.MX{Name: name, TTL: ttl, Target: sr.Exchange, Preference: uint16(sr.Preference), ProviderData: sr}
	case "SRV":
		// extract service/transport from name if present
		service, transport := "", ""
		if sr.Name != "" {
			labels := strings.Split(sr.Name, ".")
			if len(labels) >= 2 {
				service = strings.TrimPrefix(labels[0], "_")
				transport = strings.TrimPrefix(labels[1], "_")
			}
		}
		port := sr.PortInt
		if port == 0 {
			switch pv := sr.Port.(type) {
			case string:
				if v, err := strconv.Atoi(strings.TrimPrefix(pv, "_")); err == nil {
					port = v
				}
			case float64:
				port = int(pv)
			case int:
				port = pv
			}
		}
		return libdns.SRV{Name: name, TTL: ttl, Service: service, Transport: transport, Priority: uint16(sr.Priority), Weight: uint16(sr.Weight), Port: uint16(port), Target: sr.Target, ProviderData: sr}
	case "NS":
		// Use libdns.NS for nameserver records
		return libdns.NS{Name: name, TTL: ttl, Target: sr.Nameserver, ProviderData: sr}
	case "CAA":
		// Use libdns.CAA as the typed representation
		// convert stored union fields into a libdns.CAA value
		flag := 0
		if sr.Flag != nil {
			flag = *sr.Flag
		}
		var f8 uint8
		if flag < 0 {
			f8 = 0
		} else if flag > 255 {
			f8 = 255
		} else {
			f8 = uint8(flag)
		}
		return libdns.CAA{Name: name, TTL: ttl, Flags: f8, Tag: sr.Tag, Value: sr.Value, ProviderData: sr}
	case "HTTPS":
		// Convert to libdns.ServiceBinding with scheme "https"
		var params libdns.SvcParams
		if sr.SvcParams != "" {
			if p, err := libdns.ParseSvcParams(sr.SvcParams); err == nil {
				params = p
			}
		}
		target := sr.SvcTarget
		if target == "" {
			target = sr.TargetName
		}
		return libdns.ServiceBinding{
			Name:     name,
			TTL:      ttl,
			Scheme:   "https",
			Priority: uint16(sr.SvcPriority),
			Target:   target,
			Params:   params,
			ProviderData: sr,
		}
	}
	// Return nil for unsupported record types (including PTR) - they will be filtered out
	return nil
}

// fromLibdnsRR converts a libdns.Record into a spaceshipRecordUnion suitable for create/update
// Returns nil for unsupported record types
func (p *Provider) fromLibdnsRR(lr libdns.Record, zone string) *spaceshipRecordUnion {
	rr := lr.RR()
	name := rr.Name
	if name == "" {
		name = "@"
	} else {
		name = libdns.AbsoluteName(rr.Name, zone)
	}
	rec := spaceshipRecordUnion{ResourceRecordBase: ResourceRecordBase{Name: name, Type: strings.ToUpper(rr.Type), TTL: int(rr.TTL.Seconds())}}

	// MX handled specially
	if mx, ok := lr.(libdns.MX); ok {
		rec.Exchange = mx.Target
		rec.Preference = int(mx.Preference)
		return &rec
	}

	// Handle SRV records
	if srv, ok := lr.(libdns.SRV); ok {
		// map libdns.SRV fields into the spaceship payload
		rec.Service = "_" + strings.TrimPrefix(srv.Service, "_")
		rec.Protocol = "_" + strings.TrimPrefix(srv.Transport, "_")
		rec.Priority = int(srv.Priority)
		rec.Weight = int(srv.Weight)
		rec.Target = srv.Target
		rec.PortInt = int(srv.Port)
		if rec.PortInt != 0 {
			rec.Port = rec.PortInt
		}
		return &rec
	}

	// Handle NS records
	if ns, ok := lr.(libdns.NS); ok {
		rec.Nameserver = ns.Target
		return &rec
	}

	// Handle CAA records
	if caa, ok := lr.(libdns.CAA); ok {
		tmpFlag := new(int)
		*tmpFlag = int(caa.Flags)
		rec.Flag = tmpFlag
		rec.Tag = caa.Tag
		rec.Value = caa.Value
		return &rec
	}

	// Handle ServiceBinding (HTTPS) records
	if svc, ok := lr.(libdns.ServiceBinding); ok {
		// Only handle HTTPS records (ServiceBinding with scheme "https")
		if strings.ToLower(svc.Scheme) == "https" {
			rec.Type = "HTTPS"
			rec.SvcPriority = int(svc.Priority)
			rec.TargetName = svc.Target  // Use TargetName for API compatibility
			rec.SvcParams = svc.Params.String()
			return &rec
		}
		// For non-HTTPS ServiceBinding records, return nil (unsupported)
		return nil
	}

	switch v := lr.(type) {
	case libdns.Address:
		rec.Address = v.IP.String()
	case libdns.TXT:
		rec.Value = v.Text
	case libdns.CNAME:
		rec.Cname = v.Target
	case libdns.MX:
		// already handled
	default:
		// Unsupported record type (including libdns.RR)
		return nil
	}
	return &rec
}

