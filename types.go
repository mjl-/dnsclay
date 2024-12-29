package main

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/libdns/libdns"
	"github.com/miekg/dns"
)

// These are not typed in the "dns" package API. We keep them typed internally and
// convert when needed to plain uints to/from package dns.
type TTL uint32
type Serial uint32
type Class uint16
type Type uint16

// Zone for which DNS records are managed, for which a delegation with NS records
// exists. Commonly called "domains". Subdomains are not necessarily zones, they
// are just names with dots in a zone.
type Zone struct {
	// Absolute name with trailing dot. In lower-case form.
	Name string

	ProviderConfigName string `bstore:"nonzero,ref ProviderConfig"`

	// Locally known serial. Will be 0 for newly created zones. Can be different from
	// SerialRemote since not all name servers change serials on zone changes.
	SerialLocal Serial

	// Serial as known at remote. Used during refresh to decide whether to sync. Not
	// meaningful when <= 1 (e.g. always for AWS Route53).
	SerialRemote Serial

	// Last time an attempt to sync was made. Used for periodic sync.
	LastSync *time.Time

	// Last time a change in records was detected.
	LastRecordChange *time.Time

	// Time between automatic synchronizations by getting all records.
	SyncInterval time.Duration

	// Time between zone refresh: checks for an updated SOA record (after which a sync
	// is initiated). After a detected record change, checks are done more often. For 1
	// RefreshInterval, during the first 1/10th of time, a check is done 5 times. For
	// the remaining 9/10th of time, a check is also done every 10 times.
	RefreshInterval time.Duration

	NextSync    time.Time
	NextRefresh time.Time
}

type ProviderConfig struct {
	Name string

	// Name of a libdns package.
	ProviderName string

	// JSON encoding of the "Provider" type from the libdns package referenced by
	// ProviderName.
	ProviderConfigJSON string
}

// ZoneNotify is an address to DNS NOTIFY when a change to the zone is discovered.
type ZoneNotify struct {
	ID       int64
	Created  time.Time `bstore:"nonzero,default now"`
	Zone     string    `bstore:"nonzero,ref Zone"`
	Address  string    `bstore:"nonzero"` // E.g. 127.0.0.1:53
	Protocol string    `bstore:"nonzero"` // "tcp" or "udp"
}

// Credential is used for TSIG or mutual TLS authentication during DNS.
type Credential struct {
	ID           int64
	Created      time.Time `bstore:"nonzero,default now"`
	Name         string    `bstore:"nonzero,unique"` // Without trailing dot for TSIG, we add it during DNS. rfc/8945:245
	Type         string    `bstore:"nonzero"`        // "tsig" or "tlspubkey"
	TSIGSecret   string    // Base64-encoded.
	TLSPublicKey string    `bstore:"index"` // Raw-url-base64-encoded SHA-256 hash of TLS certificate subject public key info ("SPKI").
}

// ZoneCredential indicates a credential is allowed to access (get and change
// records) for a zone.
type ZoneCredential struct {
	ID           int64
	Zone         string `bstore:"nonzero,ref Zone"`
	CredentialID int64  `bstore:"nonzero,ref Credential"`
}

// Record is a DNS record that discovered through the API of the provider.
type Record struct {
	ID            int64
	Zone          string    `bstore:"nonzero,ref Zone"` // Name of zone, lower-case.
	SerialFirst   Serial    // Serial where this record first appeared. For SOA records, this is equal to its Serial field.
	SerialDeleted Serial    // Serial when record was removed. For future IXFR.
	First         time.Time `bstore:"default now,nonzero"`
	Deleted       *time.Time
	AbsName       string // Fully qualified, in lower-case.
	Type          Type   // eg A, etc.
	Class         Class
	TTL           TTL
	DataHex       string
	Value         string // Human-readable.
	ProviderID    string // From libdns.
}

func (r Record) libdnsRecord() libdns.Record {
	typ, ok := dns.TypeToString[uint16(r.Type)]
	if !ok {
		typ = fmt.Sprintf("%d", r.Type)
	}
	return libdns.Record{
		ID:    r.ProviderID,
		Type:  typ,
		Name:  libdns.RelativeName(r.AbsName, r.Zone),
		Value: r.Value,
		TTL:   time.Second * time.Duration(r.TTL),
	}
}

func libdnsRecords(l []Record) []libdns.Record {
	x := make([]libdns.Record, 0, len(l))
	for _, r := range l {
		x = append(x, r.libdnsRecord())
	}
	return x
}

type rrsetKey struct {
	AbsName string
	Type    Type
	Class   Class
}

func (r Record) rrsetKey() rrsetKey {
	return rrsetKey{r.AbsName, r.Type, r.Class}
}

// For finding existing records.
type recordKey struct {
	Name    string
	Type    Type
	Class   Class
	TTL     TTL
	DataHex string
}

func (r Record) recordKey() recordKey {
	return recordKey{r.AbsName, r.Type, r.Class, r.TTL, r.DataHex}
}

func (r Record) Header() dns.RR_Header {
	return dns.RR_Header{
		Name:   r.AbsName,
		Rrtype: uint16(r.Type),
		Class:  uint16(r.Class),
		Ttl:    uint32(r.TTL),
	}
}

func (r Record) GenericRR() *dns.RFC3597 {
	return &dns.RFC3597{
		Hdr:   r.Header(),
		Rdata: r.DataHex,
	}
}

func (r Record) RR() (dns.RR, error) {
	buf, err := hex.DecodeString(r.DataHex)
	if err != nil {
		return nil, fmt.Errorf("decode hex: %v", err)
	}

	h := r.Header()
	h.Rdlength = uint16(len(buf))
	rr, _, err := dns.UnpackRRWithHeader(h, buf, 0)
	if err != nil {
		return nil, fmt.Errorf("parse rr from buffer: %v", err)
	}
	return rr, nil
}
