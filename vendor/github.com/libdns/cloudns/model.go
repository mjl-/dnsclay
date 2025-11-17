package cloudns

import (
	"fmt"
	"net/netip"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

// ApiDnsRecord represents a DNS record retrieved from or sent to the API.
// It includes fields for record identification, configuration, and status.
type ApiDnsRecord struct {
	Id       string `json:"id"                        parameters:"record-id"`
	Type     string `json:"type"                      parameters:"record-type"`
	Host     string `json:"host"`
	Record   string `json:"record,omitempty"`
	Failover string `json:"failover"`
	Ttl      string `json:"ttl"`
	CAAFlag  uint8  `json:"caa_flag,string,omitempty"`
	CAAType  string `json:"caa_type,omitempty"`
	CAAValue string `json:"caa_value,omitempty"`
	Priority uint16 `json:"priority,string,omitempty"`
	Port     uint16 `json:"port,string,omitempty"`
	Weight   uint16 `json:"weight,string,omitempty"`
	Status   int    `json:"status"`
}

func fromLibdnsRecord(rec libdns.Record, id string) ApiDnsRecord {
	ttl := strconv.Itoa(ttlRounder(rec.RR().TTL))
	type_ := rec.RR().Type

	switch impl := rec.(type) {
	case libdns.Address:
		return ApiDnsRecord{
			Id:     id,
			Ttl:    ttl,
			Type:   type_,
			Host:   impl.Name,
			Record: impl.IP.String(),
		}
	case libdns.CAA:
		return ApiDnsRecord{
			Id:       id,
			Ttl:      ttl,
			Type:     type_,
			Host:     impl.Name,
			CAAFlag:  impl.Flags,
			CAAType:  impl.Tag,
			CAAValue: impl.Value,
		}

	case libdns.CNAME:
		return ApiDnsRecord{
			Id:     id,
			Ttl:    ttl,
			Type:   type_,
			Host:   impl.Name,
			Record: impl.Target,
		}

	case libdns.MX:
		return ApiDnsRecord{
			Id:       id,
			Ttl:      ttl,
			Type:     type_,
			Host:     impl.Name,
			Priority: impl.Preference,
			Record:   impl.Target,
		}

	case libdns.NS:
		return ApiDnsRecord{
			Id:     id,
			Ttl:    ttl,
			Type:   type_,
			Host:   impl.Name,
			Record: impl.Target,
		}
	case libdns.SRV:
		return ApiDnsRecord{
			Id:       id,
			Ttl:      ttl,
			Type:     type_,
			Host:     fmt.Sprintf("_%v._%v.%v", impl.Service, impl.Transport, impl.Name),
			Priority: impl.Priority,
			Weight:   impl.Weight,
			Port:     impl.Port,
			Record:   impl.Target,
		}
	default:
		rr := rec.RR()
		return ApiDnsRecord{
			Id:     id,
			Ttl:    ttl,
			Type:   type_,
			Host:   rr.Name,
			Record: rr.Data,
		}
	}
}

// toLibdnsRecord translates an upstream API object into a libdns
// record object.
func (r ApiDnsRecord) toLibdnsRecord() (libdns.Record, error) {
	rawttl, err := strconv.Atoi(r.Ttl)
	if err != nil {
		return libdns.RR{}, fmt.Errorf("Invalid TTL %q", r.Ttl)
	}
	ttl := time.Duration(rawttl) * time.Second

	switch r.Type {
	case "A", "AAAA":
		addr, err := netip.ParseAddr(r.Record)
		if err != nil {
			return libdns.Address{}, fmt.Errorf("Invalid IP %q: %w", r.Record, err)
		}

		return libdns.Address{
			Name: r.Host,
			TTL:  ttl,
			IP:   addr,
		}, nil
	case "CAA":
		return libdns.CAA{
			Name:  r.Host,
			TTL:   ttl,
			Flags: r.CAAFlag,
			Tag:   r.CAAType,
			Value: r.CAAValue,
		}, nil
	case "CNAME":
		return libdns.CNAME{
			Name:   r.Host,
			TTL:    ttl,
			Target: r.Record,
		}, nil
	case "MX":
		return libdns.MX{
			Name:       r.Host,
			TTL:        ttl,
			Preference: r.Priority,
			Target:     r.Record,
		}, nil
	case "NS":
		return libdns.NS{
			Name:   r.Host,
			TTL:    ttl,
			Target: r.Record,
		}, nil
	case "SRV":
		parts := strings.SplitN(r.Host, ".", 3)
		if len(parts) < 3 {
			return libdns.SRV{}, fmt.Errorf("Name %q does not have enough components (expected >3, got %v)", r.Host, len(parts))
		}
		return libdns.SRV{
			Service:   strings.TrimPrefix(parts[0], "_"),
			Transport: strings.TrimPrefix(parts[1], "_"),
			Name:      parts[2],
			TTL:       ttl,
			Priority:  r.Priority,
			Weight:    r.Weight,
			Port:      r.Port,
			Target:    r.Record,
		}, nil
	case "TXT":
		return libdns.TXT{
			Name: r.Host,
			TTL:  ttl,
			Text: r.Record,
		}, nil
	// HTTPS and SVCB do not appear supported by ClouDNS rn
	default:
		return libdns.RR{
			Name: r.Host,
			TTL:  ttl,
			Type: r.Type,
			Data: r.Record,
		}, nil
	}
}

func (r ApiDnsRecord) toParameters() map[string]string {
	ret := make(map[string]string)

	val := reflect.ValueOf(r)
	typ := reflect.TypeOf(r)
	for idx := range val.NumField() {
		field := val.Field(idx)
		if !field.IsZero() {
			name := typ.Field(idx).Tag.Get("parameters")
			if name == "" {
				name = typ.Field(idx).Tag.Get("json")
				name = strings.Split(name, ",")[0]
			}
			ret[name] = fmt.Sprintf("%v", field)
		}
	}

	return ret
}

// ApiResponse represents the structure of a standard response from the API, including status and optional data.
type ApiResponse struct {
	Status            string `json:"status"`
	StatusDescription string `json:"statusDescription"`
	Data              struct {
		Id int `json:"id"`
	} `json:"data,omitempty"`
}
