package hosttech

import (
	"fmt"
	"github.com/libdns/libdns"
	"strconv"
	"time"
)

// HosttechRecord must be implemented by each different type of record representation from the Hosttech.ch API, to allow a transformation from and to libdns.record.
type HosttechRecord interface {
	toLibdnsRecord(zone string) libdns.Record
	fromLibdnsRecord(record libdns.Record) HosttechRecord
}

// Base holds all the values that are present in each record
type Base struct {
	Id      int    `json:"id,omitempty"`
	Type    string `json:"type,omitempty"`
	TTL     int    `json:"ttl,omitempty"`
	Comment string `json:"comment,omitempty"`
}

// AAAARecord is an implementation of the AAAA record type
type AAAARecord struct {
	Base
	Name string `json:"name,omitempty"`
	IPV6 string `json:"ipv6,omitempty"`
}

func (a AAAARecord) toLibdnsRecord(zone string) libdns.Record {
	return libdns.Record{
		ID:    strconv.Itoa(a.Id),
		Type:  a.Type,
		Name:  libdns.RelativeName(a.Name, zone),
		Value: a.IPV6,
		TTL:   time.Duration(a.TTL * 1000000000),
	}
}

func (a AAAARecord) fromLibdnsRecord(record libdns.Record) HosttechRecord {
	a.Name = record.Name
	a.Type = record.Type
	a.IPV6 = record.Value
	a.TTL = durationToIntSeconds(record.TTL)
	a.Comment = generateComment()

	return a
}

// ARecord is an implementation of the A record type
type ARecord struct {
	Base
	Name string `json:"name,omitempty"`
	IPV4 string `json:"ipv4,omitempty"`
}

func (a ARecord) toLibdnsRecord(zone string) libdns.Record {
	return libdns.Record{
		ID:    strconv.Itoa(a.Id),
		Type:  a.Type,
		Name:  libdns.RelativeName(a.Name, zone),
		Value: a.IPV4,
		TTL:   time.Duration(a.TTL * 1000000000),
	}
}

func (a ARecord) fromLibdnsRecord(record libdns.Record) HosttechRecord {
	a.Name = record.Name
	a.Type = record.Type
	a.IPV4 = record.Value
	a.TTL = durationToIntSeconds(record.TTL)
	a.Comment = generateComment()

	return a
}

// CNAMERecord is an implementation of the CNAME record type
type CNAMERecord struct {
	Base
	Name  string `json:"name,omitempty"`
	Cname string `json:"cname,omitempty"`
}

func (c CNAMERecord) toLibdnsRecord(zone string) libdns.Record {
	return libdns.Record{
		ID:    strconv.Itoa(c.Id),
		Type:  c.Type,
		Name:  libdns.RelativeName(c.Name, zone),
		Value: c.Cname,
		TTL:   time.Duration(c.TTL * 1000000000),
	}
}

func (c CNAMERecord) fromLibdnsRecord(record libdns.Record) HosttechRecord {
	c.Name = record.Name
	c.Type = record.Type
	c.Cname = record.Value
	c.TTL = durationToIntSeconds(record.TTL)
	c.Comment = generateComment()

	return c
}

// MXRecord is an implementation of the MX record type
type MXRecord struct {
	Base
	Name      string `json:"name,omitempty"`
	OwnerName string `json:"ownername,omitempty"`
	Pref      uint   `json:"pref,omitempty"`
}

func (m MXRecord) toLibdnsRecord(zone string) libdns.Record {
	return libdns.Record{
		ID:       strconv.Itoa(m.Id),
		Type:     m.Type,
		Name:     libdns.RelativeName(m.OwnerName, zone),
		Value:    m.Name,
		TTL:      time.Duration(m.TTL * 1000000000),
		Priority: m.Pref,
	}
}

func (m MXRecord) fromLibdnsRecord(record libdns.Record) HosttechRecord {
	m.OwnerName = record.Name
	m.Type = record.Type
	m.TTL = durationToIntSeconds(record.TTL)
	m.Name = record.Value
	m.Pref = record.Priority
	m.Comment = generateComment()

	return m
}

// NSRecord is an implementation of the NS record type
type NSRecord struct {
	Base
	OwnerName  string `json:"ownername,omitempty"`
	TargetName string `json:"targetname,omitempty"`
}

func (n NSRecord) toLibdnsRecord(zone string) libdns.Record {
	return libdns.Record{
		ID:    strconv.Itoa(n.Id),
		Type:  n.Type,
		Name:  libdns.RelativeName(n.OwnerName, zone),
		Value: n.TargetName,
		TTL:   time.Duration(n.TTL * 1000000000),
	}
}

func (n NSRecord) fromLibdnsRecord(record libdns.Record) HosttechRecord {
	n.OwnerName = record.Name
	n.Type = record.Type
	n.TargetName = record.Value
	n.TTL = durationToIntSeconds(record.TTL)
	n.Comment = generateComment()

	return n
}

// TXTRecord is an implementation of the TXT record type
type TXTRecord struct {
	Base
	Name string `json:"name,omitempty"`
	Text string `json:"text,omitempty"`
}

func (t TXTRecord) toLibdnsRecord(zone string) libdns.Record {
	return libdns.Record{
		ID:    strconv.Itoa(t.Id),
		Type:  t.Type,
		Name:  libdns.RelativeName(t.Name, zone),
		Value: t.Text,
		TTL:   time.Duration(t.TTL * 1000000000),
	}
}

func (t TXTRecord) fromLibdnsRecord(record libdns.Record) HosttechRecord {
	t.Name = RemoveTrailingDot(record.Name)
	t.Type = record.Type
	t.Text = record.Value
	t.TTL = durationToIntSeconds(record.TTL)
	t.Comment = generateComment()

	return t
}

// TLSARecord is an implementation of the TLSA record type
type TLSARecord struct {
	Base
	Name string `json:"name,omitempty"`
	Text string `json:"text,omitempty"`
}

// HosttechZone is an implementation of the zone without records
type HosttechZone struct {
	Id          uint   `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Email       string `json:"email,omitempty"`
	TTL         uint32 `json:"ttl,omitempty"`
	Nameserver  string `json:"nameserver,omitempty"`
	DNSSEC      bool   `json:"dnssec,omitempty"`
	DNSSECEmail string `json:"dnssec_email,omitempty"`
}

func (z HosttechZone) toLibdnsZone() libdns.Zone {
	return libdns.Zone{
		Name: z.Name,
	}
}

func (t TLSARecord) toLibdnsRecord(zone string) libdns.Record {
	return libdns.Record{
		ID:    strconv.Itoa(t.Id),
		Type:  t.Type,
		Name:  libdns.RelativeName(t.Name, zone),
		Value: t.Text,
		TTL:   time.Duration(t.TTL * 1000000000),
	}
}

func (t TLSARecord) fromLibdnsRecord(record libdns.Record) HosttechRecord {
	t.Name = record.Name
	t.Type = record.Type
	t.Text = record.Value
	t.TTL = durationToIntSeconds(record.TTL)
	t.Comment = generateComment()

	return t
}

func durationToIntSeconds(duration time.Duration) int {
	durationInSeconds := duration.Seconds()
	// The minimum amount is 600 seconds
	if durationInSeconds < 600 {
		return 600
	}
	return int(durationInSeconds)
}

func generateComment() string {
	return fmt.Sprintf("This record was created or updated with libdns at %s UTC", time.Now().UTC().Format(time.Stamp))
}

type ApiError struct {
	s         string
	ErrorCode int
}

func (a ApiError) Error() string {
	return a.s
}
