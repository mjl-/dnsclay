package glesys

import (
	"fmt"
	"strings"
	"time"

	"github.com/libdns/glesys/internal/impl"
	"github.com/libdns/libdns"
)

// cleanZ removes trailing dots and spaces from a zone name.
func cleanZ(z string) string {
	return strings.TrimRight(z, ". ")
}

// toLibDNS converts a GleSYS DNSDomainRecord to a libdns Record.
func toLibDNS(dr *impl.DNSDomainRecord) (libdns.Record, error) {
	r, err := libdns.RR{
		Type: dr.Type,
		Name: dr.Host,
		Data: dr.Data,
		TTL:  time.Duration(dr.TTL) * time.Second,
	}.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse glesys record: %w", err)
	}
	return r, nil
}

func mustToLibDNS(dr *impl.DNSDomainRecord) libdns.Record {
	r, err := toLibDNS(dr)
	if err != nil {
		panic(fmt.Sprintf("failed to convert glesys record: %v", err))
	}
	return r
}

type matchParams struct {
	Name bool
	Type bool
	Data bool
	TTL  bool
}

func (mp matchParams) all() bool {
	return mp.Name && mp.Type && mp.Data && mp.TTL
}

func (mp matchParams) any() bool {
	return mp.Name || mp.Type || mp.Data || mp.TTL
}

func (mp matchParams) none() bool {
	return !mp.Name && !mp.Type && !mp.Data && !mp.TTL
}

// checkParamsMatching checks if the given libdns.RR matches the DNSDomainRecord.
// It returns a matchParams struct with the results of the comparison.
// The comparison is done by checking if the fields of the libdns.RR
// are equal to the corresponding fields of the DNSDomainRecord.
// If the fields are 'zero' (empty string or zero value), they are considered
// to checkParamsMatching. This is useful for checking if a record is already present in
// the DNS provider.
func checkParamsMatching(rr libdns.RR, dr *impl.DNSDomainRecord) matchParams {
	return matchParams{
		Name: rr.Name == "" || rr.Name == dr.Host,
		Type: rr.Type == "" || rr.Type == dr.Type,
		Data: rr.Data == "" || rr.Data == dr.Data,
		TTL:  rr.TTL == 0 || rr.TTL == time.Duration(dr.TTL)*time.Second,
	}
}
