package leaseweb

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

// @see https://developer.leaseweb.com/api-docs/domains_v2.html#tag/DNS/operation/domains-resourcerecordsets-post
var supportedTTLs = []int{60, 300, 1800, 3600, 14400, 28800, 43200, 86400}

func fromLibdns(zone string, records []libdns.Record) ([]leasewebRecordSet, error) {
	var recordsInfo = []struct {
		libdnsRecord libdns.Record
		consumed     bool
	}{}
	for _, record := range records {
		recordsInfo = append(recordsInfo, struct {
			libdnsRecord libdns.Record
			consumed     bool
		}{
			libdnsRecord: record,
			consumed:     false,
		})
	}

	var errors []string
	var recordSets []leasewebRecordSet
	var domainName = strings.TrimSuffix(zone, ".")

	for currentIdx := 0; currentIdx < len(recordsInfo); currentIdx++ {
		var currentRecordInfo = &recordsInfo[currentIdx]

		if currentRecordInfo.consumed {
			continue
		}
		currentRecordInfo.consumed = true

		// Cleanup record name and ensure it ends with domain.ext[dot] even if dns_challenge_override_domain is set
		// trimming both zone & domainName is probably overzealous, but better be safe then sorry
		// Example:
		//   zone: example.com.
		//   domainName: example.com
		//   dnsRecord.Name 1: _acme-challenge.example.com
		//   dnsRecord.Name 2: _acme-challenge.example.com.
		//   dnsRecord.Name 3: _acme-challenge.
		//   all after cleanup -> _acme-challenge.example.com.
		var recordName = fmt.Sprintf("%s.%s", strings.TrimSuffix(strings.TrimSuffix(currentRecordInfo.libdnsRecord.Name, zone), domainName), zone)
		var recordTTL = int(currentRecordInfo.libdnsRecord.TTL.Seconds())
		if !slices.Contains(supportedTTLs, recordTTL) {
			// Use the first listed TTL if the user did not provide a TTL or provided a unsupported value
			// It would probably be nice to log a warning about this, but that doesn't seem to be supported in libdns
			// Note: Leaseweb uses a default value of 3600. But as this lib will probably mostly be used
			// with caddy-dns/leaseweb to solve ACME DNS-01 challenges, let's use the lowest possible value
			recordTTL = supportedTTLs[0]
		}

		var newRecordSet = leasewebRecordSet{
			Name:    recordName,
			Type:    currentRecordInfo.libdnsRecord.Type,
			TTL:     recordTTL,
			Content: []string{currentRecordInfo.libdnsRecord.Value},
		}

		for otherIdx := 0; otherIdx < len(recordsInfo); otherIdx++ {
			var otherRecordInfo = &recordsInfo[otherIdx]
			if otherIdx == currentIdx {
				continue
			}

			if otherRecordInfo.libdnsRecord.Name == newRecordSet.Name && otherRecordInfo.libdnsRecord.Type == currentRecordInfo.libdnsRecord.Type {
				otherRecordInfo.consumed = true

				var otherTTL = int(otherRecordInfo.libdnsRecord.TTL.Seconds())
				if otherTTL != newRecordSet.TTL {
					errors = append(errors, fmt.Sprintf("Found different TTL values for %s: %d and %d.", newRecordSet.Name, newRecordSet.TTL, otherTTL))
				}

				newRecordSet.Content = append(newRecordSet.Content, otherRecordInfo.libdnsRecord.Value)
			}
		}

		recordSets = append(recordSets, newRecordSet)
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("%v", errors)
	}

	return recordSets, nil
}

func fromLeaseweb(recordSets leasewebRecordSets) []libdns.Record {
	var records []libdns.Record
	for _, resourceRecordSet := range recordSets.ResourceRecordSets {
		for _, content := range resourceRecordSet.Content {
			record := libdns.Record{
				Name:  resourceRecordSet.Name,
				Value: content,
				Type:  resourceRecordSet.Type,
				TTL:   time.Duration(resourceRecordSet.TTL) * time.Second,
			}
			records = append(records, record)
		}
	}
	return records
}
