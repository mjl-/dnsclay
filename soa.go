package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"

	"github.com/miekg/dns"
)

var mockGetSOA func(zone string) (*dns.SOA, error)

// getSOA gets a SOA record from the authoritative name servers. The record is not
// DNSSEC-verified.
func getSOA(ctx context.Context, log *slog.Logger, zone string) (rsoa *dns.SOA, rerr error) {
	if mockGetSOA != nil {
		return mockGetSOA(zone)
	}

	metricSOAGet.Inc()
	defer func() {
		if rerr != nil {
			metricSOAGetErrors.Inc()
		}
	}()

	// Lookup NS. We use the default resolver, using the locally configured recursive
	// resolver, which may verify DNSSEC signatures.
	nsl, err := net.DefaultResolver.LookupNS(ctx, zone)
	if err == nil && len(nsl) == 0 {
		err = errors.New("no nameservers")
	}
	if err != nil {
		return nil, fmt.Errorf("looking up nameservers: %w", err)
	}

	// We will be asking for the SOA record directly from the authoritative name
	// servers. We don't do DNSSEC verification for simplicity.
	client := dns.Client{Net: "tcp"}

	lastErr := errors.New("cannot happen")
	for _, ns := range nsl {
		ips, err := net.DefaultResolver.LookupIPAddr(ctx, ns.Host)
		if err == nil && len(ips) == 0 {
			err = fmt.Errorf("no ips")
		}
		if err != nil {
			lastErr = fmt.Errorf("looking up ips for nameserver: %w", err)
			log.Error("looking up ips for nameserver", "err", err, "nameserver", ns.Host)
			continue
		}

		for _, ip := range ips {
			addr := net.JoinHostPort(ip.String(), "53")
			var om dns.Msg
			om.SetQuestion(zone, dns.TypeSOA)
			om.RecursionDesired = false
			var soa *dns.SOA
			im, _, err := client.ExchangeContext(ctx, &om, addr)
			if err == nil {
				err = responseError(im)
			}
			if err == nil && len(im.Answer) != 1 {
				err = fmt.Errorf("got %d answer resource records (%v), expected 1 soa", len(im.Answer), im.Answer)
			} else if err == nil {
				var ok bool
				soa, ok = im.Answer[0].(*dns.SOA)
				if !ok {
					err = fmt.Errorf("response not soa record, but %v", im.Answer[0])
				}
			}
			if err != nil {
				lastErr = err
				log.Error("querying soa record from nameserver, continuing with others", "err", err, "nameserver", ns.Host, "ip", ip)
				continue
			}
			return soa, nil
		}
	}
	return nil, lastErr
}
