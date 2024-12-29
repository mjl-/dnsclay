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
func getRecords(ctx context.Context, log *slog.Logger, provider Provider, zone string) ([]libdns.Record, error) {
	soa, err := getSOA(ctx, log, zone)
	if err != nil {
		return nil, fmt.Errorf("get latest soa: %w", err)
	}

	records, err := provider.GetRecords(ctx, zone)
	if err != nil {
		return nil, fmt.Errorf("get records: %w", err)
	}

	for _, r := range records {
		if r.Name == "" && strings.EqualFold(r.Type, "SOA") {
			return records, nil
		}
	}
	h := soa.Header()
	lrsoa := libdns.Record{
		Type:  "SOA",
		Name:  h.Name,
		Value: strings.TrimPrefix(soa.String(), h.String()),
		TTL:   time.Duration(h.Ttl) * time.Second,
	}
	records = append(records, lrsoa)
	return records, nil
}
