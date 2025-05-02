package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/libdns/libdns"
	"github.com/miekg/dns"
)

// todo: figure out which other dns record types/constructs we need to handle. at least dname records?

// RecordSet holds the records (values) for a name and type, and optionally
// historic state of the records over time.
type RecordSet struct {
	// Always at least one record. All with the same name, type and ttl. Sorted by
	// value.
	Records []Record

	// Filled with either the full historic propagation state including the
	// current/latest value (when returned by ZoneRecordSetHistory), or only those
	// propagation states that are not the current value but may still be relevant (may
	// still be in DNS caches).
	States []PropagationState
}

// PropagationState indicates value(s) of a record set in a period, or that a
// negative lookup result may be cached somewhere.
type PropagationState struct {
	Start time.Time
	End   *time.Time // If nil, then still active.

	// If true, this state represents a period during which a negative lookup result
	// may be cached. Records will be nil.
	Negative bool

	// Records active during the period Start-End.
	Records []Record
}

// TTLPeriod indicates the maximum TTL for negative lookups for a period.
type TTLPeriod struct {
	Start, End     time.Time
	MaxNegativeTTL time.Duration
}

// gatherMaxNegativeTTLs returns effective maximum negative TTLs based on
// (historic) SOA records. soaSets must be sorted by increasing "First" creation. A
// zone always has a SOA record, so periods should be contiguous, and callers
// should only use the TTLPeriod.Start (ignoring End) to ensure that effect.
//
// The returned periods are non-overlapping, sorted by Start.
func gatherMaxNegativeTTLs(now time.Time, soaSets []RecordSet) (l []TTLPeriod, rerr error) {
	if len(soaSets) == 0 {
		return nil, fmt.Errorf("got zero soa record sets, need at least 1")
	}

	// We will be building up "l" without overlap, sorted by Start, which is also how
	// we return it.
	// o is the offset into l that we won't have to change anymore. Each SOA record
	// could require changing/merging multiple period negative ttls.
	o := 0

SOA:
	for i, set := range soaSets {
		if len(set.Records) != 1 {
			return nil, fmt.Errorf("got %d soa records in a set, 1 required", len(set.Records))
		}
		if i > 0 && !soaSets[i-1].Records[0].First.Before(set.Records[0].First) {
			return nil, fmt.Errorf("soa record sets not sorted")
		}

		r := set.Records[0]
		rr, err := r.RR()
		if err != nil {
			return nil, fmt.Errorf("parsing soa record: %v", err)
		}
		soa, ok := rr.(*dns.SOA)
		if !ok {
			return nil, fmt.Errorf("not a soa record")
		}
		ttl := time.Duration(r.Header().Ttl) * time.Second
		negTTL := time.Duration(soa.Minttl) * time.Second

		// Effective start & end of SOA record given TTL.
		start := r.First
		var end time.Time
		if r.Deleted != nil {
			end = r.Deleted.Add(ttl)
		} else {
			end = now.Add(ttl)
		}

		if len(l) == 0 {
			l = append(l, TTLPeriod{start, end, negTTL})
			continue
		}

		if o > 0 && start.Before(l[o-1].End) {
			panic(fmt.Sprintf("internal error, o %d, p %v, start %v, l%v", o, l[o-1], start, l))
		}

		// We'll go through the periods, adjust periods. We'll align start with the
		// periods, and split/replace/skip the current period.
		i := o
		for i < len(l) && start.Before(end) {
			p := l[i]

			if start.After(p.End) {
				break
			}

			if negTTL <= p.MaxNegativeTTL {
				i++
				start = p.End
				continue
			}

			if start.After(p.Start) {
				if i != o {
					panic("internal error")
				}
				// Split p. Next iteration will replace the negTTL and handle end.
				l[i].End = start
				l = append(l, TTLPeriod{})
				copy(l[i+2:], l[i+1:len(l)-1])
				l[i+1] = TTLPeriod{start, p.End, p.MaxNegativeTTL}
				i += 1
				o += 1
				continue
			}

			// start is now equal to p.Start.

			if end.Before(p.End) {
				// We'll split p, keeping p's end and inserting a new element before it.
				l = append(l, TTLPeriod{})
				copy(l[i+1:], l[i:len(l)-1])
				l[i] = TTLPeriod{start, end, negTTL}
				l[i+1].Start = end
				continue SOA
			}

			// end >= p.End

			// Overwrite p.
			l[i].MaxNegativeTTL = negTTL
			i++
			o++
			start = p.End
			// if end > p.End, there is likely a next entry to be handled on the next
			// iteration.
		}
		if !start.Before(end) {
			continue
		}

		last := l[len(l)-1]
		if start.Equal(last.End) && negTTL == last.MaxNegativeTTL {
			l[i-1].End = end
		} else if !start.Before(last.End) {
			// This may leave a gap, but should be okay if callers just use TTLPeriod.Start,
			// ignoring End.
			l = append(l, TTLPeriod{start, end, negTTL})
			o = len(l) - 1
		} else {
		}
	}
	return l, nil
}

type RecordSetKey struct {
	AbsName string
	Type    Type
}

// propagationStates groups records (that should include deleted records) by record
// set and adds historic & current propagation states to the last version of each
// record set. Negative lookup results and the "negative ttl" (minttl) of the SOA
// record is taken into account, and so are wildcard records.
//
// If typ is >= 0, the returned map only has a single record set with for that
// record type and name. Otherwise the returned map includes all record sets.
//
// If oldActiveOnly is set, only historic propagation states are included (not the
// current value), and only those still in effect (not those that can't be in any
// cache anymore).
func propagationStates(now time.Time, l []Record, relName string, typ int, oldActiveOnly bool) (map[RecordSetKey][]RecordSet, error) {
	m := map[RecordSetKey][]RecordSet{}

	if len(l) == 0 {
		return m, nil
	}
	absName := libdns.AbsoluteName(relName, l[0].Zone)

	// Sort by first. We'll go through the records below to merge them into record sets.
	sort.Slice(l, func(i, j int) bool {
		return l[i].First.Before(l[j].First)
	})

	// Split records into sets, over time. Each name+type can have multiple record sets.
	for _, r := range l {
		// If we are not returning all record sets, we only need:
		// - Exact name and type match.
		// - Matching wildcard name and type.
		// - Zone SOA records, for negative TTL.
		switch {
		default:
			continue
		case typ < 0,
			typ == int(r.Type) && r.AbsName == absName,
			typ == int(r.Type) && r.AbsName == "*."+strings.SplitN(absName, ".", 2)[1],
			r.Type == Type(dns.TypeSOA) && r.AbsName == r.Zone:
		}

		k := RecordSetKey{r.AbsName, r.Type}
		sets := m[k]
		if len(sets) > 0 && sets[len(sets)-1].Records[0].First.Equal(r.First) {
			sets[len(sets)-1].Records = append(sets[len(sets)-1].Records, r)
		} else {
			m[k] = append(m[k], RecordSet{Records: []Record{r}, States: []PropagationState{}})
		}
	}

	// Get SOA record sets.
	soaSets := m[RecordSetKey{l[0].Zone, Type(dns.TypeSOA)}]
	if len(soaSets) == 0 {
		return nil, fmt.Errorf("no soa record sets found")
	}
	// todo: more sanity checks on SOA records. must not start after, or end before regular records?

	// Determine effective max negative TTL for zone for full period, based on SOA records.
	negTTLs, err := gatherMaxNegativeTTLs(now, soaSets)
	if err != nil {
		return nil, fmt.Errorf("gathering effective max negative ttl periods: %v", err)
	}

	// For each name,type record sets, add propagation state to the latest rrset of a name+type.
	for k, versions := range m {
		if typ >= 0 && (typ != int(k.Type) || k.AbsName != absName) {
			continue
		}

		wildcardAbsName := "*." + strings.SplitN(k.AbsName, ".", 2)[1]
		var wildcards []RecordSet
		if wildcardAbsName != k.AbsName {
			wildcards = m[RecordSetKey{wildcardAbsName, k.Type}]
		}

		addSetsPropagationStates(now, versions, negTTLs, wildcards, oldActiveOnly)
	}

	if typ >= 0 {
		// Only keep the versions explicitly requested.
		k := RecordSetKey{absName, Type(typ)}
		set, ok := m[k]
		if !ok {
			return nil, fmt.Errorf("%w: record set does not exist", errUser)
		}
		m = map[RecordSetKey][]RecordSet{k: set}
	}

	// Sort the records within each set version by value, for consistent display and
	// simpler testing.
	for _, versions := range m {
		for _, set := range versions {
			sort.Slice(set.Records, func(i, j int) bool {
				return set.Records[i].Value < set.Records[j].Value
			})
		}
	}

	return m, nil
}

// addSetsPropagationStates adds propagation state to the last RecordSet in "sets".
// negTTLs is used for the effective max negative ttl (from the soa records).
// wilcard is any matching wildcard record set, applicable during periods where a
// set does not exist.
func addSetsPropagationStates(now time.Time, versions []RecordSet, negTTLs []TTLPeriod, wildcards []RecordSet, oldActiveOnly bool) {
	// We go through each version of the set, building up states.
	states := []PropagationState{}
	for i, set := range versions {
		r0 := set.Records[0]
		end := effectiveEnd(r0)
		if r0.Deleted == nil {
			end = nil
		}

		// If the previous version ended before this version appeared, we have a gap. We
		// fill it with wildcard records and/or negative TTLs.
		var start time.Time
		if i > 0 {
			start = *versions[i-1].Records[0].Deleted
		}
		states = fillGapWildcardNegative(states, now, versions[0].Records[0].First, start, r0.First, negTTLs, wildcards, oldActiveOnly)

		if !oldActiveOnly || end == nil || end.After(now) {
			states = append(states, PropagationState{r0.First, end, false, set.Records})
		}
	}

	// Add effective wildcards from after deletion.
	deleted := versions[len(versions)-1].Records[0].Deleted
	if deleted != nil {
		for len(wildcards) > 0 {
			if wildcards[0].Records[0].Deleted == nil || wildcards[0].Records[0].Deleted.After(*deleted) {
				break
			}
			wildcards = wildcards[1:]
		}
		for _, set := range wildcards {
			start := set.Records[0].First
			if start.Before(*deleted) {
				start = *deleted
			}
			end := set.Records[0].Deleted
			if end != nil {
				tm := end.Add(time.Duration(set.Records[0].TTL) * time.Second)
				end = &tm
			}
			states = append(states, PropagationState{start, end, false, set.Records})
		}
	}

	// Drop the latest state if it is still active and we are only asked to return
	// non-current still cachable states.
	if oldActiveOnly && len(states) > 0 && !states[len(states)-1].Negative && states[len(states)-1].Records[0].Deleted == nil {
		states = states[:len(states)-1]
	}

	versions[len(versions)-1].States = states
}

// fillGapWildcardNegative adds to state based on wildcards or max negative ttls,
// for the period gapstart to gapend. The added states may cover periods beyond
// gapend due to caching.
//
// If gapstart is the zero time, it is for the period before the first record. Only
// negative lookup result states are added for the period before the first exact or
// wildcard record.  first is the first time the regular (not wildcard) record
// was created.
func fillGapWildcardNegative(states []PropagationState, now, first, gapstart, gapend time.Time, negTTLs []TTLPeriod, wildcards []RecordSet, oldActiveOnly bool) []PropagationState {
	// We fill up the period gapstart-gapend with entries from wildcard or negTTLs. We
	// iterate through them.

	for gapstart.Before(gapend) {
		// Skip ttls and wildcards that are no longer relevant.
		for len(negTTLs) > 0 && gapstart.After(negTTLs[0].End) {
			negTTLs = negTTLs[1:]
		}
		for len(wildcards) > 0 && wildcards[0].Records[0].Deleted != nil && gapstart.After(*effectiveEnd(wildcards[0].Records[0])) {
			wildcards = wildcards[1:]
		}

		// Add state for wildcard record.
		if len(wildcards) > 0 && !gapstart.Before(wildcards[0].Records[0].First) {
			r0 := wildcards[0].Records[0]
			effEnd := effectiveEnd(r0)
			if effEnd == nil || effEnd.After(gapend) {
				tm := gapend.Add(time.Duration(r0.TTL) * time.Second)
				effEnd = &tm
			}
			if !oldActiveOnly || effEnd.After(now) {
				states = append(states, PropagationState{gapstart, effEnd, false, wildcards[0].Records})
			}
			if r0.First.Before(first) {
				first = r0.First
			}
			gapstart = effEnd.Add(-time.Duration(r0.TTL) * time.Second)
			wildcards = wildcards[1:]
			continue
		}

		// This can only happen in SOA records aren't complete. OK to leave it as is.
		if len(negTTLs) == 0 {
			if len(wildcards) > 0 {
				gapstart = wildcards[0].Records[0].First
				continue
			}
			break
		}

		p := negTTLs[0]
		pend := p.End
		if pend.After(gapend) {
			pend = gapend
		}
		if len(wildcards) > 0 && pend.After(wildcards[0].Records[0].First) {
			pend = wildcards[0].Records[0].First
		}
		pend = pend.Add(p.MaxNegativeTTL)
		if !oldActiveOnly || pend.After(now) && (!gapstart.IsZero() || len(states) > 0 || !p.Start.Before(gapstart)) {
			if len(states) == 0 && gapstart.Before(first) {
				gapstart = first.Add(-p.MaxNegativeTTL)
			}
			states = append(states, PropagationState{gapstart, &pend, true, []Record{}})
		}
		gapstart = pend.Add(-p.MaxNegativeTTL)
	}

	return states
}

func effectiveEnd(r Record) *time.Time {
	if r.Deleted == nil {
		return nil
	}
	tm := r.Deleted.Add(time.Duration(r.TTL) * time.Second)
	return &tm
}
