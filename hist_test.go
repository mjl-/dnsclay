package main

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/miekg/dns"

	"github.com/mjl-/bstore"
)

func (p TTLPeriod) String() string {
	const format = "15:04:05"
	return fmt.Sprintf("%v %v %v", p.Start.Format(format), p.End.Format(format), p.MaxNegativeTTL)
}

func TestGatherNegativeTTL(t *testing.T) {
	now := time.Now().Round(0)
	start := now.Add(-10 * time.Minute)

	// record-start, record-end, ttl-end, negttl
	rrset := func(s, e, negttl int) RecordSet {
		p := (e - s) / 2
		del := start.Add(time.Duration(s+p) * time.Second)
		ttl := e - (s + p)

		rr, err := dns.NewRR(fmt.Sprintf(". %d SOA ns0.localhost. mjl.localhost. 2024123100 3600 300 1209600 %d", ttl, negttl))
		tcheck(t, err, "newrr")
		hex, value, err := recordData(rr)
		tcheck(t, err, "soa data")

		return RecordSet{
			Records: []Record{
				{
					First:   start.Add(time.Duration(s) * time.Second),
					Deleted: &del,
					TTL:     TTL(rr.Header().Ttl),
					Type:    Type(dns.TypeSOA),
					DataHex: hex,
					Value:   value,
				},
			},
		}
	}
	period := func(s, e, negttl int) TTLPeriod {
		return TTLPeriod{
			start.Add(time.Duration(s) * time.Second),
			start.Add(time.Duration(e) * time.Second),
			time.Duration(negttl) * time.Second,
		}
	}

	test := func(rrsets []RecordSet, expPeriods []TTLPeriod) {
		t.Helper()
		l, err := gatherMaxNegativeTTLs(now, rrsets)
		tcheck(t, err, "gather max negative ttls")

		tcompare(t, l, expPeriods)
	}

	test([]RecordSet{
		rrset(0, 180, 600),
		rrset(60, 120, 900),
	}, []TTLPeriod{
		period(0, 60, 600),
		period(60, 120, 900),
		period(120, 180, 600),
	})

	test([]RecordSet{
		rrset(0, 60, 300),
		rrset(60, 120, 300),
		rrset(120, 180, 300),
	}, []TTLPeriod{
		period(0, 180, 300),
	})

	test([]RecordSet{
		rrset(0, 120, 600),    // initial
		rrset(10, 20, 120),    // smaller than previous, dropped
		rrset(60, 120, 300),   // smaller than previous, dropped
		rrset(110, 140, 60),   // first part smaller, second part new.
		rrset(130, 150, 120),  // last part replaced.
		rrset(150, 160, 120),  // last part extended with same negttl
		rrset(160, 170, 60),   // last part extended with lower negttl
		rrset(170, 180, 180),  // last part extended with higher negttl
		rrset(180, 280, 600),  //
		rrset(190, 270, 1200), //
		rrset(200, 260, 1800), //
		rrset(300, 310, 60),   // with gap before
		rrset(320, 330, 60),   // with gap
	}, []TTLPeriod{
		period(0, 120, 600),
		period(120, 130, 60),
		period(130, 160, 120),
		period(160, 170, 60),
		period(170, 180, 180),
		period(180, 190, 600),
		period(190, 200, 1200),
		period(200, 260, 1800),
		period(260, 270, 1200),
		period(270, 280, 600),
		period(300, 310, 60),
		period(320, 330, 60),
	})
}

type input struct {
	start, end int
	name       string
	ttl        int
	typ        uint16
	value      string
}
type output struct {
	start, end int
	negative   bool
	name       string
	ttl        int
	typ        string
	values     []string
}

func testHist(t *testing.T, inputs []input, outputs []output) {
	t.Helper()

	os.Remove("testdata/dnsclay.db")

	dbopts := bstore.Options{
		Timeout: 5 * time.Second,
	}
	var err error
	database, err = bstore.Open(ctxbg, "testdata/dnsclay.db", &dbopts, databaseTypes...)
	xcheckf(err, "open database")
	defer func() {
		err := database.Close()
		logCheck(slog.Default(), err, "close db")
		database = nil
	}()

	shutdownCtx, shutdownCancel = context.WithCancel(ctxbg)
	defer shutdownCancel()

	zone := "example."

	tx, err := database.Begin(ctxbg, true)
	tcheck(t, err, "begin tx")
	defer func() {
		if tx != nil {
			tx.Rollback()
		}
	}()

	pc := ProviderConfig{"test", "fake", "{}"}
	err = tx.Insert(&pc)
	tcheck(t, err, "insert providerconfig")
	z := Zone{Name: zone, ProviderConfigName: pc.Name}
	err = tx.Insert(&z)
	tcheck(t, err, "insert zone")

	now := time.Now().Round(0)

	serialTimes := map[int]bool{0: true}
	for _, inp := range inputs {
		serialTimes[inp.start] = true
		if inp.end >= 0 {
			serialTimes[inp.end] = true
		}
	}
	time2soa := map[int]Record{}
	times := slices.Sorted(maps.Keys(serialTimes))
	for i, v := range times {
		var deleted *time.Time
		var serialDel uint32
		if i < len(times)-1 {
			tm := now.Add(time.Duration(times[i+1]) * time.Second)
			deleted = &tm
			serialDel = uint32(2 + i + 1)
		}
		first := now.Add(time.Duration(v) * time.Second)
		serialFirst := 2 + i

		value := fmt.Sprintf("ns0.example. mail.example. %d 3600 300 1209600 300", serialFirst)
		rr, err := dns.NewRR(fmt.Sprintf("%s 300 SOA %s", zone, value))
		tcheck(t, err, "parse rr")
		hex, value, err := recordData(rr)
		tcheck(t, err, "parse rr")

		r := Record{0, zone, Serial(serialFirst), Serial(serialDel), first, deleted, zone, Type(dns.TypeSOA), Class(dns.ClassINET), 300, hex, value, ""}
		err = tx.Insert(&r)
		tcheck(t, err, "insert first soa")
		time2soa[v] = r
	}

	for _, inp := range inputs {
		ssoa := time2soa[inp.start]
		esoa := time2soa[inp.end]
		datahex := ""
		var deleted *time.Time
		if esoa.ID > 0 {
			deleted = &esoa.First
		}
		r := Record{0, zone, ssoa.SerialFirst, esoa.SerialFirst, ssoa.First, deleted, inp.name + "." + zone, Type(inp.typ), Class(dns.ClassINET), TTL(inp.ttl), datahex, inp.value, ""}
		err := tx.Insert(&r)
		tcheck(t, err, "insert record")
	}

	err = tx.Commit()
	tx = nil
	tcheck(t, err, "tx commit")

	api := API{}
	records := api.ZoneRecords(ctxbg, zone)
	m, err := propagationStates(now, records, "", -1, true)
	tcheck(t, err, "propagation states")
	tcompare(t, len(m) > 1, true)

	xtime := func(seconds int) *time.Time {
		if seconds < 0 {
			return nil
		}
		tm := now.Add(time.Duration(seconds) * time.Second)
		return &tm
	}

	m, err = propagationStates(now, records, "host1", int(dns.TypeA), false)
	tcheck(t, err, "propagation states")
	tcompare(t, len(m), 1)
	versions := m[RecordSetKey{"host1.example.", Type(dns.TypeA)}]
	hist := versions[len(versions)-1].States
	tcompare(t, len(hist), len(outputs))
	for i, h := range hist {
		o := outputs[i]
		tcompare(t, h.Start, *xtime(o.start))
		tcompare(t, h.End, xtime(o.end))
		tcompare(t, h.Negative, o.negative)
		tcompare(t, len(h.Records), len(o.values))
		if !h.Negative {
			tcompare(t, h.Records[0].TTL, TTL(o.ttl))
		}
	}
}

// todo: updating soa records with new negative ttl in between existence of a set.
// todo: test with various overlapping negatives

func TestHistoryState0(t *testing.T) {
	inputs := []input{
		{300, 900, "host0", 300, dns.TypeA, "9.9.9.9"}, // Irrelevant.
		{600, 1200, "host1", 300, dns.TypeA, "1.1.1.1"},
		{600, 1200, "host1", 300, dns.TypeA, "2.2.2.2"},
		{1500, 1800, "host1", 600, dns.TypeA, "3.3.3.3"},
		{1800, 1900, "host1", 300, dns.TypeA, "3.3.3.3"},
		{1900, -1, "host1", 600, dns.TypeA, "4.4.4.4"},
	}

	outputs := []output{
		{start: 300, end: 900, negative: true},
		{600, 1200 + 300, false, "host1", 300, "A", []string{"1.1.1.1", "2.2.2.2"}},
		{start: 1200, end: 1800, negative: true},
		{1500, 1800 + 600, false, "host1", 600, "A", []string{"3.3.3.3"}},
		{1800, 1900 + 300, false, "host1", 300, "A", []string{"3.3.3.3"}}, // todo: we could merge this into previous...
		{1900, -1, false, "host1", 600, "A", []string{"4.4.4.4"}},
	}

	testHist(t, inputs, outputs)
}

func TestHistoryState1(t *testing.T) {
	inputs := []input{
		{300, 900, "host0", 300, dns.TypeA, "9.9.9.9"}, // Irrelevant.
		{600, 1200, "host1", 300, dns.TypeA, "1.1.1.1"},
		{900, 1500, "*", 600, dns.TypeA, "3.3.3.3"}, //
	}

	outputs := []output{
		{start: 300, end: 900, negative: true},
		{600, 1200 + 300, false, "host1", 300, "A", []string{"1.1.1.1"}},
		{1200, 1500 + 600, false, "*", 600, "A", []string{"3.3.3.3"}},
	}

	testHist(t, inputs, outputs)
}

func TestHistoryState2(t *testing.T) {
	inputs := []input{
		{300, 900, "host0", 300, dns.TypeA, "9.9.9.9"}, // Irrelevant.
		{600, 1200, "*", 300, dns.TypeA, "1.1.1.1"},
		{900, 1500, "host1", 600, dns.TypeA, "3.3.3.3"},
	}

	outputs := []output{
		{start: 600, end: 900, negative: true},
		{600, 900 + 300, false, "*", 300, "A", []string{"1.1.1.1"}},
		{900, 1500 + 600, false, "host1", 600, "A", []string{"3.3.3.3"}},
	}

	testHist(t, inputs, outputs)
}

func TestHistoryState3(t *testing.T) {
	inputs := []input{
		{300, 900, "host0", 300, dns.TypeA, "9.9.9.9"}, // Irrelevant.
		{600, 750, "*", 300, dns.TypeA, "1.1.1.1"},
		{700, 800, "host1", 600, dns.TypeA, "2.2.2.2"},
		{750, 1200, "*", 300, dns.TypeA, "1.1.1.2"},
		{900, 1000, "host1", 600, dns.TypeA, "3.3.3.3"},
	}

	outputs := []output{
		{start: 700 - 300, end: 600 + 300, negative: true},
		{600, 700 + 300, false, "*", 300, "A", []string{"1.1.1.1"}},
		{700, 800 + 600, false, "host1", 600, "A", []string{"2.2.2.2"}},
		{800, 900 + 300, false, "*", 300, "A", []string{"1.1.1.2"}},
		{900, 1000 + 600, false, "host1", 600, "A", []string{"3.3.3.3"}},
		{1000, 1200 + 300, false, "*", 300, "A", []string{"1.1.1.2"}},
	}

	testHist(t, inputs, outputs)
}

func TestHistoryState4(t *testing.T) {
	// Wildcard ends before record is created again 2nd time.
	inputs := []input{
		{300, 900, "host0", 300, dns.TypeA, "9.9.9.9"}, // Irrelevant.
		{600, 1000, "host1", 300, dns.TypeA, "1.1.1.1"},
		{900, 1200, "*", 300, dns.TypeA, "2.2.2.2"},
		{1100, 1300, "host1", 300, dns.TypeA, "3.3.3.3"},
	}

	outputs := []output{
		{start: 600 - 300, end: 600 + 300, negative: true},
		{600, 1000 + 300, false, "host1", 300, "A", []string{"1.1.1.1"}},
		{1000, 1100 + 300, false, "host1", 300, "A", []string{"2.2.2.2"}},
		{1100, 1300 + 300, false, "host1", 300, "A", []string{"3.3.3.3"}},
	}

	testHist(t, inputs, outputs)
}
