// Package dnsupdate implements a DNS record management client compatible
// with the libdns interfaces for the DNS UPDATE protocol defined in RFC 2136.
package dnsupdate

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/libdns/libdns"
	"github.com/miekg/dns"
)

// Provider facilitates DNS record manipulation with the DNS UPDATE protocol.
type Provider struct {
	// DNS server address
	Addr string `json:"addr,omitempty"`
}

func (p *Provider) roundTrip(ctx context.Context, query *dns.Msg) (*dns.Msg, error) {
	client := dns.Client{Net: "tcp"}
	reply, _, err := client.ExchangeContext(ctx, query, p.Addr)
	if err != nil {
		return nil, err
	} else if reply.Rcode != dns.RcodeSuccess {
		return nil, fmt.Errorf("DNS error: %v", dns.RcodeToString[reply.Rcode])
	}
	return reply, nil
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	var query dns.Msg
	query.SetAxfr(zone)

	reply, err := p.roundTrip(ctx, &query)
	if err != nil {
		return nil, err
	}

	return unmarshalRecords(zone, reply.Answer), nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	rrs, err := marshalRecords(zone, records)
	if err != nil {
		return nil, err
	}

	var query dns.Msg
	query.SetUpdate(zone)
	query.Insert(rrs)

	if _, err := p.roundTrip(ctx, &query); err != nil {
		return nil, err
	}

	return unmarshalRecords(zone, rrs), nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	insertRRs, err := marshalRecords(zone, records)
	if err != nil {
		return nil, err
	}

	// Build a set of RRs that are new
	m := make(map[string]struct{})
	for _, rr := range insertRRs {
		m[rr.String()] = struct{}{}
	}

	// Remove old RRs that aren't in the set of new RRs
	var removeRRs []dns.RR
	for _, record := range records {
		if record.ID == "" {
			continue
		}
		if _, ok := m[record.ID]; ok {
			continue
		}

		rr, err := parseRecordID(record.ID)
		if err != nil {
			return nil, err
		}

		removeRRs = append(removeRRs, rr)
	}

	var query dns.Msg
	query.SetUpdate(zone)
	query.Insert(insertRRs)
	query.Remove(removeRRs)

	if _, err := p.roundTrip(ctx, &query); err != nil {
		return nil, err
	}

	return unmarshalRecords(zone, insertRRs), nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	rrs := make([]dns.RR, len(records))
	for i, record := range records {
		// If a record ID was supplied, use that. Otherwise, generate a RR
		// from the record data fields.
		var (
			rr  dns.RR
			err error
		)
		if record.ID != "" {
			rr, err = parseRecordID(record.ID)
		} else {
			rr, err = marshalRecord(zone, &record)
		}
		if err != nil {
			return nil, err
		}
		rrs[i] = rr
	}

	rrs, err := marshalRecords(zone, records)
	if err != nil {
		return nil, err
	}

	var query dns.Msg
	query.SetUpdate(zone)
	query.Remove(rrs)

	if _, err := p.roundTrip(ctx, &query); err != nil {
		return nil, err
	}

	return unmarshalRecords(zone, rrs), err
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)

func marshalRecords(zone string, records []libdns.Record) ([]dns.RR, error) {
	rrs := make([]dns.RR, 0, len(records))
	for _, record := range records {
		rr, err := marshalRecord(zone, &record)
		if err != nil {
			return nil, err
		}
		rrs = append(rrs, rr)
	}
	return rrs, nil
}

func marshalRecord(zone string, record *libdns.Record) (dns.RR, error) {
	fqdn := libdns.AbsoluteName(record.Name, zone)
	ttl := uint32(record.TTL / time.Second)
	raw := fmt.Sprintf("%v %v IN %v %v", fqdn, ttl, record.Type, record.Value)
	return dns.NewRR(raw)
}

func unmarshalRecords(zone string, rrs []dns.RR) []libdns.Record {
	records := make([]libdns.Record, 0, len(rrs))
	for _, rr := range rrs {
		hdr := rr.Header()
		records = append(records, libdns.Record{
			ID:    formatRecordID(rr, zone),
			Type:  dns.Type(hdr.Rrtype).String(),
			Name:  hdr.Name,
			Value: strings.TrimPrefix(rr.String(), hdr.String()),
			TTL:   time.Duration(hdr.Ttl) * time.Second,
		})
	}
	return records
}

func formatRecordID(rr dns.RR, zone string) string {
	// We use the zone file representation of the record (with FQDN) as ID
	rr = dns.Copy(rr)
	rr.Header().Name = libdns.AbsoluteName(rr.Header().Name, zone)
	return rr.String()
}

func parseRecordID(id string) (dns.RR, error) {
	rr, err := dns.NewRR(id)
	if err != nil {
		return nil, fmt.Errorf("invalid record ID (%v)", err)
	}
	return rr, nil
}
