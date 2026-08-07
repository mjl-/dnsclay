package googleclouddns

import (
	"fmt"
	"strings"
	"time"

	"github.com/libdns/libdns"
	"google.golang.org/api/dns/v1"
)

// libdnsRecords is a convience type to make the code more readable as we move between
// sets of libdns.Records and the Cloud DNS record set.
type libdnsRecords []libdns.Record

type dnsMetadata struct {
	name       string
	recordType string
}

// groupRecordsByType groups libdns.Record entries by name and type to ensure multiple
// values are sent at the same time to Google Cloud DNS.
func (l libdnsRecords) groupRecordsByType() map[dnsMetadata]libdnsRecords {
	gdrs := make(map[dnsMetadata]libdnsRecords)
	for _, record := range l {
		dnsRecord := dnsMetadata{
			name:       record.Name,
			recordType: record.Type,
		}
		if gdr, ok := gdrs[dnsRecord]; !ok {
			gdrs[dnsRecord] = libdnsRecords{record}
		} else {
			gdrs[dnsRecord] = append(gdr, record)
		}
	}
	return gdrs
}

// hasRecord returns if this set of records contains the specified record. Only the name,
// type, and value are compared; the TTL is ignored.
func (l libdnsRecords) hasRecord(record libdns.Record) bool {
	for _, existingRecord := range l {
		if record.Name == existingRecord.Name && record.Type == existingRecord.Type && record.Value == existingRecord.Value {
			return true
		}
	}
	return false
}

// doesNotHaveRecords returns true if this set of records does not contain the specified
// record. Only the name, type, and value are compared; the TTL is ignored.
func (l libdnsRecords) doesNotHaveRecord(record libdns.Record) bool {
	return !l.hasRecord(record)
}

// prepValuesForCloudDNS returns a slice of strings containing the values from this set of
// records. Note that if the value contains spaces, this will add quotes to ensure
// it is properly populated in Cloud DNS.
func (l libdnsRecords) prepValuesForCloudDNS() []string {
	values := make([]string, 0)
	for _, record := range l {
		value := record.Value
		if strings.Contains(value, " ") {
			//ensure we quote a value with spaces but do not double quote
			value = fmt.Sprintf(`"%s"`, strings.Trim(value, `"`))
		}
		values = append(values, value)

	}
	return values
}

// convertToLibDNS takes Cloud DNS record set and converts it into a set of libdns
// records. Note that this will remove the quotes around a value.
func convertToLibDNS(googleRecord *dns.ResourceRecordSet, zone string) libdnsRecords {
	records := make([]libdns.Record, 0)
	for _, value := range googleRecord.Rrdatas {
		// there can be multiple values per record  so
		// let's treat each one as a separate libdns Record
		record := libdns.Record{
			Type:  googleRecord.Type,
			Name:  libdns.RelativeName(googleRecord.Name, zone),
			Value: strings.Trim(value, `"`),
			TTL:   time.Duration(googleRecord.Ttl) * time.Second,
		}
		records = append(records, record)
	}
	return records
}
