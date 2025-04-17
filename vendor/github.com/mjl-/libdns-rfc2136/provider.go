package rfc2136

import (
	"context"
	"fmt"
	"time"

	"github.com/libdns/libdns"
	"github.com/miekg/dns"
)

type Provider struct {
	KeyName string `json:"key_name,omitempty"`
	KeyAlg  string `json:"key_alg,omitempty"`
	Key     string `json:"key,omitempty"`
	Server  string `json:"server,omitempty"`
}

func (p *Provider) keyNameFQDN() string {
	return dns.Fqdn(p.KeyName)
}

func (p *Provider) client() *dns.Client {
	return &dns.Client{
		TsigSecret: map[string]string{p.keyNameFQDN(): p.Key},
		Net:        "tcp",
	}
}

func (p *Provider) setTsig(msg *dns.Msg) {
	msg.SetTsig(p.keyNameFQDN(), dns.Fqdn(p.KeyAlg), 300, time.Now().Unix())
}

func (p *Provider) exchange(ctx context.Context, msg *dns.Msg) error {
	m, _, err := p.client().ExchangeContext(ctx, msg, p.Server)
	if err != nil {
		return err
	}
	if m.Rcode == dns.RcodeSuccess {
		return nil
	}

	err = fmt.Errorf("dns response error code %q (%d)", dns.RcodeToString[m.Rcode], m.Rcode)
	if opt := m.IsEdns0(); opt != nil {
		for _, o := range opt.Option {
			if ede, ok := o.(*dns.EDNS0_EDE); ok {
				err = fmt.Errorf("%s: %s (%d): %s", err, dns.ExtendedErrorCodeToString[ede.InfoCode], ede.InfoCode, ede.ExtraText)
			}
		}
	}
	return err
}

func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	zone = dns.Fqdn(zone)

	conn, err := p.client().DialContext(ctx, p.Server)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	tn := dns.Transfer{
		Conn: conn,
	}
	tn.TsigSecret = map[string]string{p.keyNameFQDN(): p.Key}

	msg := dns.Msg{}
	msg.SetAxfr(zone)
	p.setTsig(&msg)

	res, err := tn.In(&msg, p.Server)
	if err != nil {
		return nil, fmt.Errorf("start zone transfer: %w", err)
	}

	records := make([]libdns.Record, 0)
	for e := range res {
		if e.Error != nil {
			return nil, fmt.Errorf("zone transfer: %w", e.Error)
		}

		for _, rr := range e.RR {
			records = append(records, recordFromRR(rr, zone))
		}
	}

	return records, nil
}

func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	zone = dns.Fqdn(zone)

	msg := dns.Msg{}
	msg.SetUpdate(zone)

	rrs := make([]dns.RR, 0, len(records))
	for _, rec := range records {
		rr, err := recordToRR(rec, zone)
		if err != nil {
			return nil, fmt.Errorf("invalid record %s: %w", rec.Name, err)
		}
		rrs = append(rrs, rr)
	}

	msg.RemoveRRset(rrs)
	msg.Insert(rrs)

	p.setTsig(&msg)
	if err := p.exchange(ctx, &msg); err != nil {
		return nil, err
	}
	return records, nil
}

func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	zone = dns.Fqdn(zone)

	msg := dns.Msg{}
	msg.SetUpdate(zone)
	for _, rec := range records {
		rr, err := recordToRR(rec, zone)
		if err != nil {
			return nil, fmt.Errorf("invalid record %s: %w", rec.Name, err)
		}
		msg.Insert([]dns.RR{rr})
	}
	p.setTsig(&msg)
	if err := p.exchange(ctx, &msg); err != nil {
		return nil, err
	}
	return records, nil
}

func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	zone = dns.Fqdn(zone)

	msg := dns.Msg{}
	msg.SetUpdate(zone)
	for _, rec := range records {
		rr, err := recordToRR(rec, zone)
		if err != nil {
			return nil, fmt.Errorf("invalid record %s: %w", rec.Name, err)
		}
		msg.Remove([]dns.RR{rr})
	}
	p.setTsig(&msg)
	if err := p.exchange(ctx, &msg); err != nil {
		return nil, err
	}
	return records, nil
}

var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
