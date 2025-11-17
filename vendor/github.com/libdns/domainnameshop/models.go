package domainnameshop

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

// dsZone JSON data structure.
type dsZone struct {
	Name           string   `json:"domain"`
	ID             int      `json:"id"`
	ExpiryDate     string   `json:"expiry_date"`
	Nameservers    []string `json:"nameservers"`
	RegisteredDate string   `json:"registered_date"`
	Registrant     string   `json:"registrant"`
	Renew          bool     `json:"renew"`
	Services       Service  `json:"services"`
	Status         string
}

type Service struct {
	DNS       bool   `json:"dns"`
	Email     bool   `json:"email"`
	Registrar bool   `json:"registrar"`
	Webhotel  string `json:"webhotel"`
}

// dsDNSRecord JSON data structure.
// https://api.domeneshop.no/docs/#tag/dns_record_models
type dsDNSRecord struct {
	ID       int    `json:"id,omitempty"`
	Host     string `json:"host,omitempty"`
	Data     string `json:"data,omitempty"`
	Type     string `json:"type,omitempty"`
	TTL      int    `json:"ttl,omitempty"` // In seconds must be multiple of 60
	Priority string `json:"priority,omitempty"`
	Weight   string `json:"weight,omitempty"`
	Port     string `json:"port,omitempty"`
}

func (r dsDNSRecord) libdnsRecord() (libdns.Record, error) {
	switch r.Type {
	case "MX":
		priority, err := strconv.ParseUint(r.Priority, 10, 16)
		if err != nil {
			return libdns.MX{}, fmt.Errorf("invalid priority %s: %v", r.Priority, err)
		}
		rr := libdns.MX{
			Name:       r.Host,
			TTL:        time.Duration(r.TTL) * time.Second,
			Preference: uint16(priority),
			Target:     r.Data,
		}
		return rr, nil

	case "SRV":
		priority, err := strconv.ParseUint(r.Priority, 10, 16)
		if err != nil {
			return libdns.SRV{}, fmt.Errorf("invalid priority %s: %v", r.Priority, err)
		}
		weight, err := strconv.ParseUint(r.Weight, 10, 16)
		if err != nil {
			return libdns.SRV{}, fmt.Errorf("invalid weight %s: %v", r.Weight, err)
		}
		port, err := strconv.ParseUint(r.Port, 10, 16)
		if err != nil {
			return libdns.SRV{}, fmt.Errorf("invalid port %s: %v", r.Port, err)
		}

		parts := strings.SplitN(r.Host, ".", 2)
		if len(parts) < 2 {
			return libdns.SRV{}, fmt.Errorf("name %v does not contain enough fields; expected format: '_service._proto'", r.Host)
		}

		rr := libdns.SRV{
			Service:   strings.TrimPrefix(parts[0], "_"),
			Transport: strings.TrimPrefix(parts[1], "_"),
			Name:      r.Host,
			TTL:       time.Duration(r.TTL) * time.Second,
			Priority:  uint16(priority),
			Weight:    uint16(weight),
			Port:      uint16(port),
			Target:    r.Data,
		}

		return rr, nil

	default:
		rr := libdns.RR{
			Name: r.Host,
			TTL:  time.Duration(r.TTL) * time.Second,
			Type: r.Type,
			Data: r.Data,
		}
		return rr.Parse()
	}

}

func libdnsRecordTodsDNSRecord(r libdns.Record) (dsDNSRecord, error) {
	rr := r.RR()

	dsRecord := dsDNSRecord{
		Host: rr.Name,
		TTL:  int(rr.TTL.Seconds()),
		Type: rr.Type,
		Data: rr.Data,
	}

	switch rec := r.(type) {
	case libdns.MX:
		dsRecord.Priority = strconv.Itoa(int(rec.Preference))

	case libdns.SRV:
		dsRecord.Priority = strconv.Itoa(int(rec.Priority))
		dsRecord.Port = strconv.Itoa(int(rec.Port))
		dsRecord.Weight = strconv.Itoa(int(rec.Weight))
		dsRecord.Data = rec.Target
	}

	return dsRecord, nil

}
