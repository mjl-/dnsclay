package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

// getRecords gets records from the provider and adds a SOA fetched directly from
// the authoritative name server if not present in the records from the provider.
// The separately fetched SOA record is not DNSSEC-verified.
func getRecords(ctx context.Context, log *slog.Logger, provider Provider, zone string, needSOA bool) ([]libdnsRR, error) {
	records, err := provider.GetRecords(ctx, zone)
	if err != nil {
		return nil, fmt.Errorf("get records: %w", err)
	}

	if !needSOA {
		for _, rr := range records {
			if rr.Name == "@" && strings.EqualFold(rr.Type, "SOA") {
				return records, nil
			}
		}
	}

	soa, err := getSOA(ctx, log, zone)
	if err != nil {
		return nil, fmt.Errorf("get latest soa: %w", err)
	}

	h := soa.Header()
	lrsoa := libdnsRR{
		RR: libdns.RR{
			Name: libdns.RelativeName(h.Name, zone),
			Type: "SOA",
			TTL:  time.Duration(h.Ttl) * time.Second,
			Data: strings.TrimPrefix(soa.String(), h.String()),
		},
	}
	records = append(records, lrsoa)
	return records, nil
}
