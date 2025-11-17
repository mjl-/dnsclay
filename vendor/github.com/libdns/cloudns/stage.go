package cloudns

import (
	"iter"
	"slices"

	"github.com/libdns/libdns"
)

const (
	nop = iota
	addRecord
	modifyRecord
	deleteRecord
)

type operation int

type operationEntry struct {
	op     operation
	record ApiDnsRecord
}

func compareIDlessRecord(a ApiDnsRecord, b ApiDnsRecord) bool {
	return a.Type == b.Type &&
		a.Host == b.Host &&
		a.Record == b.Record &&
		a.Ttl == b.Ttl &&
		a.CAAFlag == b.CAAFlag &&
		a.CAAType == b.CAAType &&
		a.Priority == b.Priority &&
		a.Port == b.Port &&
		a.Weight == b.Weight
}

// createUpdateOperations processes an existing rrset and a new rrset and comes
// up with a set of operations to sync them. This could be a lot better,
// since we'll generate a bunch of update operations if there's a new
// entry in the middle of the list or if the lists are not sorted.
func createUpdateOperations(existingRRSet []ApiDnsRecord, desiredRRSet []libdns.RR, deleted map[ApiDnsRecord]bool) []operationEntry {
	existingIter, existingStop := iter.Pull(slices.Values(existingRRSet))
	defer existingStop()
	desiredIter, desiredStop := iter.Pull(slices.Values(desiredRRSet))
	defer desiredStop()
	ret := make([]operationEntry, 0, max(len(existingRRSet)+len(desiredRRSet)))

	for {
		existingRR, existingOk := existingIter()
		desiredRR, desiredOk := desiredIter()
		if existingOk && desiredOk {
			modifiedRR := fromLibdnsRecord(desiredRR, existingRR.Id)
			if !compareIDlessRecord(existingRR, modifiedRR) {
				ret = append(ret, operationEntry{
					op:     modifyRecord,
					record: modifiedRR,
				})
			}
		}

		if existingOk && !desiredOk {
			deleted[existingRR] = true
		}

		if !existingOk && desiredOk {
			ret = append(ret, operationEntry{
				op:     addRecord,
				record: fromLibdnsRecord(desiredRR, ""),
			})
		}

		if !existingOk && !desiredOk {
			break
		}
	}

	return ret
}

func makeOperationList(desired map[nameAndType][]libdns.RR, existing map[nameAndType][]ApiDnsRecord) []operationEntry {
	ret := make([]operationEntry, 0, len(desired))
	deleted := make(map[ApiDnsRecord]bool)

	for nt, desiredRRSet := range desired {
		existingRRSet := existing[nt]

		if len(existingRRSet) == 0 {
			// create
			for _, desiredRR := range desiredRRSet {
				ret = append(ret, operationEntry{
					op:     addRecord,
					record: fromLibdnsRecord(desiredRR, ""),
				})
			}
		} else {
			// update
			ret = append(
				ret,
				createUpdateOperations(
					existingRRSet,
					desiredRRSet,
					deleted,
				)...,
			)
		}
	}

	// Now prepend all of our deletions so that we don't get any
	// errors for duplicate records
	ops := make([]operationEntry, 0, len(deleted)+len(ret))
	for deletion := range deleted {
		ops = append(ops, operationEntry{
			op:     deleteRecord,
			record: deletion,
		})
	}
	return append(ops, ret...)
}
