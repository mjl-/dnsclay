// Utility functions needed by the provider methods

package netcup

import (
	"strings"
	"time"

	"github.com/libdns/libdns"
)

// Strips the trailing dot from a FQDN
func unFQDN(fqdn string) string {
	return strings.TrimSuffix(fqdn, ".")
}

// Convert single netcup record to libdns record.
func toLibdnsRecord(netcupRecord dnsRecord, ttl int64) (libdns.Record, error) {
	switch netcupRecord.RecType {
	case "MX":
		// MX is the only one where the preference/priority is set dedicated on netcup api side
		return libdns.MX{
			Name:       netcupRecord.HostName,
			Target:     netcupRecord.Destination,
			Preference: uint16(netcupRecord.Priority),
			TTL:        time.Duration(ttl * int64(time.Second)),
		}, nil
	default:
		{
			rr := libdns.RR{
				Name: netcupRecord.HostName,
				Type: netcupRecord.RecType,
				Data: netcupRecord.Destination,
				TTL:  time.Duration(ttl * int64(time.Second)),
			}

			// Make sure we return a native entry type instead of RR type
			return rr.Parse()
		}
	}
}

// Converts netcup records to libdns records. Since the netcup records don't have individual TTLs, the given TTL is used for all libdns records.
func toLibdnsRecords(netcupRecords []dnsRecord, ttl int64) ([]libdns.Record, error) {
	var libdnsRecords []libdns.Record
	for _, record := range netcupRecords {
		libdnsRecord, err := toLibdnsRecord(record, ttl)
		// NOTE: Aborting here on a single invalid record is the best option.
		// Additionally, netcup already validates DNS entries upon saving, so this is very unlikely.
		if err != nil {
			return []libdns.Record{}, err
		}
		libdnsRecords = append(libdnsRecords, libdnsRecord)
	}
	return libdnsRecords, nil
}

// Converts libdns records to netcup records.
func toNetcupRecords(libnsRecords []libdns.Record) []dnsRecord {
	var netcupRecords []dnsRecord
	for _, record := range libnsRecords {
		// Make sure we have an RR record at hand
		rr := record.RR()

		// Parse the priority out of the RR record, when required
		// NOTE: This is not the cleanest solution, but it works reliably
		priority := 0
		if rr.Type == "MX" {
			libdnsRecord, _ := rr.Parse()
			mxRecord, ok := libdnsRecord.(libdns.MX)
			if ok {
				priority = int(mxRecord.Preference)
			}
		}

		// NOTE: We loose the ID during conversion
		netcupRecord := dnsRecord{
			ID:          "",
			HostName:    rr.Name,
			RecType:     rr.Type,
			Destination: rr.Data,
			Priority:    priority,
		}
		netcupRecords = append(netcupRecords, netcupRecord)
	}
	return netcupRecords
}

// difference returns the records that are in a but not in b
func difference(a, b []dnsRecord) []dnsRecord {
	bIDmap := make(map[dnsRecord]struct{}, len(b))
	for _, elm := range b {
		bIDmap[elm] = struct{}{}
	}

	var diff []dnsRecord
	for _, elm := range a {
		if _, found := bIDmap[elm]; !found {
			diff = append(diff, elm)
		}
	}

	return diff
}

// Searches for a record with the given ID in the given records.
func findRecordByID(id string, records []dnsRecord) *dnsRecord {
	for _, record := range records {
		if record.ID == id {
			return &record
		}
	}

	return nil
}

// Searches for a record with the given host name and record type in the given records.
// Only the first one found is returned.
func findRecordByNameAndType(hostName string, recType string, records []dnsRecord) *dnsRecord {
	for _, record := range records {
		if record.HostName == hostName && record.RecType == recType {
			return &record
		}
	}

	return nil
}

// Searches for a record with the given host name, record type and priority in the given records.
// Only the first one found is returned.
func findRecordByNameAndTypeAndPriority(hostName string, recType string, priority int, records []dnsRecord) *dnsRecord {
	for _, record := range records {
		if record.HostName == hostName && record.RecType == recType && record.Priority == priority {
			return &record
		}
	}

	return nil
}

// Searches for a record in the given records.
// The first criterion is the ID. If that's not set, then the name and type (and optionally the priority, if it's an MX record) are used.
// Only the first one found is returned.
func findRecord(record dnsRecord, records []dnsRecord) *dnsRecord {
	var foundRecord *dnsRecord
	if record.ID != "" {
		foundRecord = findRecordByID(record.ID, records)
	} else if record.RecType != "MX" {
		foundRecord = findRecordByNameAndType(record.HostName, record.RecType, records)
	} else {
		foundRecord = findRecordByNameAndTypeAndPriority(record.HostName, record.RecType, record.Priority, records)
	}

	return foundRecord
}

// Returns all records from appendRecords, that are not in existingRecords.
func getRecordsToAppend(appendRecords []dnsRecord, existingRecords []dnsRecord) []dnsRecord {
	var recordsToAppend []dnsRecord
	for _, record := range appendRecords {
		foundRecord := findRecord(record, existingRecords)
		if foundRecord == nil || !foundRecord.equals(record) {
			recordsToAppend = append(recordsToAppend, record)
		}
	}
	return recordsToAppend
}

// Returns all records from setRecords, that either are not in existingRecords or have a differentValue there.
func getRecordsToSet(setRecords []dnsRecord, existingRecords []dnsRecord) []dnsRecord {
	var recordsToUpdate []dnsRecord
	var recordsToAppend []dnsRecord
	for _, record := range setRecords {
		foundRecord := findRecord(record, existingRecords)
		if foundRecord != nil && !foundRecord.equals(record) {
			record.ID = foundRecord.ID
			recordsToUpdate = append(recordsToUpdate, record)
		} else if foundRecord == nil {
			recordsToAppend = append(recordsToAppend, record)
		}
	}
	return append(recordsToUpdate, recordsToAppend...)
}

// Returns all records from deleteRecords, that are in existingRecords.
func getRecordsToDelete(deleteRecords []dnsRecord, existingRecords []dnsRecord) []dnsRecord {
	var recordsToDelete []dnsRecord
	for _, record := range deleteRecords {
		foundRecord := findRecord(record, existingRecords)
		if foundRecord != nil {
			record.ID = foundRecord.ID
			record.Destination = foundRecord.Destination
			record.DeleteRecord = true
			recordsToDelete = append(recordsToDelete, record)
		}
	}
	return recordsToDelete
}

// Do a full comparison of two records.
func rrEqualsRR(a, b libdns.RR) bool {
	return a.Name == b.Name && a.Type == b.Type && a.TTL.String() == b.TTL.String() && a.Data == b.Data
}

// Compare two records by their ID (Name and Type).
func rrMatchID(a, b libdns.RR) bool {
	return a.Name == b.Name && a.Type == b.Type
}
