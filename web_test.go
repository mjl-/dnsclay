package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/libdns/libdns"
	"github.com/miekg/dns"
)

var adminmux = makeAdminMux()

func TestWebAuthExport(t *testing.T) {

	testAuth := func(path string) {
		// Missing credentials.
		req := httptest.NewRequest("GET", path, nil)
		rec := httptest.ResponseRecorder{}
		adminmux.ServeHTTP(&rec, req)
		tcompare(t, rec.Code, http.StatusUnauthorized)

		// Bad password.
		req.SetBasicAuth("admin", "bad")
		rec = httptest.ResponseRecorder{}
		adminmux.ServeHTTP(&rec, req)
		tcompare(t, rec.Code, http.StatusUnauthorized)

		// Allowed.
		req.SetBasicAuth("admin", adminpassword)
		rec = httptest.ResponseRecorder{}
		adminmux.ServeHTTP(&rec, req)
		tcompare(t, rec.Code, http.StatusOK)
	}

	testDNS(t, func(te testEnv, z Zone) {
		testAuth("/dnsclay.db")
		testAuth("/")
		testAuth("/api/")
		testAuth("/license")
	})
}

func TestZoneRefresh(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		te.zoneUnchanged(func() {
			te.sherpaError("user:notFound", func() { te.api.ZoneRefresh(ctxbg, "bogus") })

			te.api.ZoneRefresh(ctxbg, z.Name)
		})
	})
}

func TestRecordSetAdd(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		te.sherpaError("user:notFound", func() {
			te.api.RecordSetAdd(ctxbg, "bogus", RecordSetChange{"testhost2", 600, Type(dns.TypeA), []string{"10.0.0.3"}})
		})
		te.sherpaError("user:error", func() {
			te.api.RecordSetAdd(ctxbg, z.Name, RecordSetChange{"bad name", 600, Type(dns.TypeA), []string{"10.0.0.3"}})
		})
		te.sherpaError("user:error", func() { te.api.RecordSetAdd(ctxbg, z.Name, RecordSetChange{"testhost2", 600, 0, []string{""}}) })      // Bad type
		te.sherpaError("user:error", func() { te.api.RecordSetAdd(ctxbg, z.Name, RecordSetChange{"testhost2", 600, Type(dns.TypeA), nil}) }) // No values.

		tc := te.zoneChanged(func() {
			rsn := RecordSetChange{"testhost2", 600, Type(dns.TypeA), []string{"10.0.0.2", "10.0.0.3"}}
			te.api.RecordSetAdd(ctxbg, z.Name, rsn)
		})
		tc.checkRecordDelta(typecounts{}, typecounts{"A": 2})

		te.sherpaError("user:error", func() {
			te.api.RecordSetAdd(ctxbg, z.Name, RecordSetChange{"testhost2", 600, Type(dns.TypeA), []string{"10.0.0.3"}})
		}) // Already exists.

		// Add long TXT record, testing for escaping.
		tc = te.zoneChanged(func() {
			longvalue := make([]byte, 500)
			for i := range longvalue {
				longvalue[i] = 'a'
			}
			rsn := RecordSetChange{"testtxt2", 600, Type(dns.TypeTXT), []string{string(longvalue)}}
			te.api.RecordSetAdd(ctxbg, z.Name, rsn)
		})
		tc.checkRecordDelta(typecounts{}, typecounts{"TXT": 1})

		// Add SRV record that stock libdns doesn't think is valid.
		tc = te.zoneChanged(func() {
			rsn := RecordSetChange{"testsrv2", 600, Type(dns.TypeSRV), []string{"0 0 143 other.example.com."}}
			te.api.RecordSetAdd(ctxbg, z.Name, rsn)
		})
		tc.checkRecordDelta(typecounts{}, typecounts{"SRV": 1})

		// Add SVCB record that stock libdns doesn't think is valid.
		tc = te.zoneChanged(func() {
			rsn := RecordSetChange{"testsvcb2", 600, Type(dns.TypeSVCB), []string{"0 test.example."}}
			te.api.RecordSetAdd(ctxbg, z.Name, rsn)
		})
		tc.checkRecordDelta(typecounts{}, typecounts{"SVCB": 1})

		// Add MX record.
		tc = te.zoneChanged(func() {
			rsn := RecordSetChange{"mx2", 600, Type(dns.TypeMX), []string{"10 test.example."}}
			te.api.RecordSetAdd(ctxbg, z.Name, rsn)
		})
		tc.checkRecordDelta(typecounts{}, typecounts{"MX": 1})
	})
}

func TestRecordSetDelete(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		tc := te.zoneChanged(func() {
			var prevRecords []int64
			for _, r := range te.z0.records {
				if r.AbsName == "testhost."+z.Name && r.Type == Type(dns.TypeA) {
					prevRecords = append(prevRecords, r.ID)
				}
			}
			tcompare(t, len(prevRecords), 2)

			te.sherpaError("user:notFound", func() { te.api.RecordSetDelete(ctxbg, "bogus", "testhost", Type(dns.TypeA), prevRecords) })
			te.sherpaError("user:error", func() { te.api.RecordSetDelete(ctxbg, z.Name, "testhost", Type(dns.TypeTLSA), prevRecords) })
			te.sherpaError("user:error", func() { te.api.RecordSetDelete(ctxbg, z.Name, "testhost", Type(0), prevRecords) })

			te.api.RecordSetDelete(ctxbg, z.Name, "testhost", Type(dns.TypeA), prevRecords)
			te.sherpaError("user:error", func() { te.api.RecordSetDelete(ctxbg, z.Name, "testhost", Type(dns.TypeA), prevRecords) })
		})
		tc.checkRecordDelta(typecounts{"A": 2}, typecounts{})
	})

	testDNS(t, func(te testEnv, z Zone) {
		tc := te.zoneChanged(func() {
			var prevRecords []int64
			for _, r := range te.z0.records {
				if r.Type == Type(dns.TypeTXT) {
					prevRecords = append(prevRecords, r.ID)
				}
			}
			tcompare(t, len(prevRecords), 1)

			te.api.RecordSetDelete(ctxbg, z.Name, "testtxt", Type(dns.TypeTXT), prevRecords)
			te.sherpaError("user:error", func() { te.api.RecordSetDelete(ctxbg, z.Name, "testtxt", Type(dns.TypeTXT), prevRecords) })
		})
		tc.checkRecordDelta(typecounts{"TXT": 1}, typecounts{})
	})
}

func TestRecordSetUpdate(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		rsn := RecordSetChange{"testhost", 600, Type(dns.TypeA), []string{"10.0.0.1", "10.0.0.3"}}

		records := map[uint16][]int64{}
		for _, r := range te.z0.records {
			records[uint16(r.Type)] = append(records[uint16(r.Type)], r.ID)
		}

		tc := te.zoneChanged(func() {
			var prevRecords []int64
			var valueRecords []int64
			for _, r := range te.z0.records {
				if r.AbsName == "testhost."+z.Name && r.Type == Type(dns.TypeA) {
					prevRecords = append(prevRecords, r.ID)
					if r.Value == "10.0.0.1" {
						valueRecords = []int64{r.ID}
					}
				}
			}
			tcompare(t, len(prevRecords), 2)
			tcompare(t, len(valueRecords), 1)
			valueRecords = append(valueRecords, 0)

			te.sherpaError("user:notFound", func() { te.api.RecordSetUpdate(ctxbg, "bogus", "testhost", rsn, prevRecords, valueRecords) })
			te.sherpaError("user:error", func() { te.api.RecordSetUpdate(ctxbg, z.Name, "bogus", rsn, prevRecords, valueRecords) })
			te.sherpaError("user:error", func() { te.api.RecordSetUpdate(ctxbg, z.Name, "testhost", rsn, prevRecords, []int64{999}) })     // Bad valueRecordID.
			te.sherpaError("user:error", func() { te.api.RecordSetUpdate(ctxbg, z.Name, "testhost", rsn, prevRecords[:1], valueRecords) }) // Incomplete prevRecords.
			te.sherpaError("user:error", func() { te.api.RecordSetUpdate(ctxbg, z.Name, "testhost", rsn, []int64{999}, valueRecords) })    // Unknown prevRecords.
			te.sherpaError("user:error", func() {
				te.api.RecordSetUpdate(ctxbg, z.Name, "testhost", RecordSetChange{"testhost", 600, Type(dns.TypeAAAA), []string{""}}, prevRecords, valueRecords)
			}) // Type mismatch.

			te.api.RecordSetUpdate(ctxbg, z.Name, "testhost", rsn, prevRecords, valueRecords)
		})
		tc.checkRecordDelta(typecounts{"A": 2}, typecounts{"A": 2})

		// Try making "changes" to a record set, but keeping everything the same.
		te.zoneUnchanged(func() {
			var prevRecords []int64
			for _, r := range te.z0.records {
				if r.AbsName == "testhost."+z.Name && r.Type == Type(dns.TypeA) {
					prevRecords = append(prevRecords, r.ID)
				}
			}
			tcompare(t, len(prevRecords), 2)
			te.sherpaError("user:error", func() { te.api.RecordSetUpdate(ctxbg, z.Name, "testhost", rsn, prevRecords, prevRecords) }) // No changes.
		})

		// Update TXT record, testing for escaping.
		tc = te.zoneChanged(func() {
			rsn := RecordSetChange{"testtxt", 300, Type(dns.TypeTXT), []string{`modified`}}
			te.api.RecordSetUpdate(ctxbg, z.Name, "testtxt", rsn, records[dns.TypeTXT], records[dns.TypeTXT])
		})
		tc.checkRecordDelta(typecounts{"TXT": 1}, typecounts{"TXT": 1})

		// Add SVCB record that stock libdns doesn't think is valid.
		tc = te.zoneChanged(func() {
			rsn := RecordSetChange{"", 600, Type(dns.TypeSVCB), []string{"0 modified.example."}}
			te.api.RecordSetUpdate(ctxbg, z.Name, "", rsn, records[dns.TypeSVCB], records[dns.TypeSVCB])
		})
		tc.checkRecordDelta(typecounts{"SVCB": 1}, typecounts{"SVCB": 1})

		// Add MX record.
		tc = te.zoneChanged(func() {
			rsn := RecordSetChange{"", 600, Type(dns.TypeMX), []string{"10 modified.example."}}
			te.api.RecordSetUpdate(ctxbg, z.Name, "", rsn, records[dns.TypeMX], records[dns.TypeMX])
		})
		tc.checkRecordDelta(typecounts{"MX": 1}, typecounts{"MX": 1})
	})
}

func TestZoneImportRecords(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		te.sherpaError("user:notFound", func() { te.api.ZoneImportRecords(ctxbg, "bogus", "") })
		te.sherpaError("user:error", func() { te.api.ZoneImportRecords(ctxbg, z.Name, "bad zone file") })

		tc := te.zoneChanged(func() {
			zonefile := `
testhost2 300 IN A 10.0.0.4
testhost3 600 IN MX 10 testhost2
`
			te.api.ZoneImportRecords(ctxbg, z.Name, zonefile)
		})
		tc.checkRecordDelta(typecounts{}, typecounts{"A": 1, "MX": 1})
	})
}

func TestPurgeHistory(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		te.sherpaError("user:notFound", func() { te.api.ZonePurgeHistory(ctxbg, "bogus") })

		tc := te.zoneChanged(func() {
			_, err := te.z0.p.AppendRecords(ctxbg, z.Name, []libdns.Record{ldr("", "testhost", 300, "A", "10.0.0.3")})
			tcheck(t, err, "append record")
			te.api.ZoneRefresh(ctxbg, z.Name)
		})
		tc.checkRecordDelta(typecounts{"A": 2}, typecounts{"A": 3})

		te.api.ZonePurgeHistory(ctxbg, z.Name)

		// History is purged, so we don't see the deletion.
		records := te.api.ZoneRecords(ctxbg, z.Name)
		for _, r := range records {
			if r.Deleted != nil {
				t.Fatalf("still have a deleted record after purging history")
			}
		}
	})
}

func TestZoneDelete(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		te.api.ZoneDelete(ctxbg, z.Name)
		te.sherpaError("user:notFound", func() { te.api.ZoneDelete(ctxbg, z.Name) })
		te.sherpaError("user:notFound", func() { te.api.Zone(ctxbg, z.Name) })
	})
}

func TestZones(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		l := te.api.Zones(ctxbg)
		tcompare(t, l, []Zone{z, te.z1.z})
	})
}

func TestProviderConfigTest(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		te.sherpaError("user:error", func() { te.api.ProviderConfigTest(ctxbg, z.Name, 3600, "badprovider", te.z0.pc.ProviderConfigJSON) })
		te.sherpaError("user:error", func() { te.api.ProviderConfigTest(ctxbg, z.Name, 3600, "fake", "bad json") })
		te.sherpaError("user:error", func() { te.api.ProviderConfigTest(ctxbg, z.Name, 3600, "fake", `{"ID": "unknownzone"}`) })

		te.api.ProviderConfigTest(ctxbg, z.Name, 3600, "fake", te.z0.pc.ProviderConfigJSON)
	})
}

func TestZoneCredentialAdd(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		// Unknown zone.
		te.sherpaError("user:notFound", func() {
			te.api.ZoneCredentialAdd(ctxbg, "bogus", Credential{0, time.Time{}, "tsig1", "tsig", "bWFkZSB5b3UgbG9vayEK", ""})
		})

		te.api.ZoneCredentialAdd(ctxbg, z.Name, Credential{0, time.Time{}, "tsig0", "tsig", "bWFkZSB5b3UgbG9vayEK", ""})
		nc1 := te.api.ZoneCredentialAdd(ctxbg, z.Name, Credential{0, time.Time{}, "tsig1", "tsig", "", ""})
		tcompare(t, len(nc1.TSIGSecret) > 0, true)
		te.api.ZoneCredentialAdd(ctxbg, z.Name, Credential{0, time.Time{}, "pubkey0", "tlspubkey", "", "MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDE"})

		// Duplicate name.
		te.sherpaError("user:error", func() {
			te.api.ZoneCredentialAdd(ctxbg, z.Name, Credential{0, time.Time{}, "tsig1", "tsig", "bWFkZSB5b3UgbG9vayEK", ""})
		})
		// Bad type.
		te.sherpaError("user:error", func() {
			te.api.ZoneCredentialAdd(ctxbg, z.Name, Credential{0, time.Time{}, "tsig2", "badtype", "", ""})
		})
		// Bad tls pub key length.
		te.sherpaError("user:error", func() {
			te.api.ZoneCredentialAdd(ctxbg, z.Name, Credential{0, time.Time{}, "tlspubkey2", "tlspubkey", "", "bWFkZSB5b3UgbG9vayEK"})
		})
	})
}

func TestZoneCredentialDelete(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		nc0 := te.api.ZoneCredentialAdd(ctxbg, z.Name, Credential{0, time.Time{}, "tsig1", "tsig", "bWFkZSB5b3UgbG9vayEK", ""})
		nc1 := te.api.ZoneCredentialAdd(ctxbg, z.Name, Credential{0, time.Time{}, "pubkey1", "tlspubkey", "", "MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDE"})

		te.api.ZoneCredentialDelete(ctxbg, nc0.ID)
		te.api.ZoneCredentialDelete(ctxbg, nc1.ID)

		te.sherpaError("user:notFound", func() { te.api.ZoneCredentialDelete(ctxbg, nc0.ID) })
		te.sherpaError("user:notFound", func() { te.api.ZoneCredentialDelete(ctxbg, nc1.ID) })
	})
}

func TestZoneNotify(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		_, _, notifies, _, _ := te.api.Zone(ctxbg, z.Name)
		tcompare(t, len(notifies), 1)
		on := notifies[0]

		te.api.ZoneNotifyDelete(ctxbg, on.ID)
		te.sherpaError("user:notFound", func() { te.api.ZoneNotifyDelete(ctxbg, on.ID) })
		bogusn := on
		bogusn.Zone = "bogus"
		te.sherpaError("user:error", func() { te.api.ZoneNotifyAdd(ctxbg, bogusn) })
		on = te.api.ZoneNotifyAdd(ctxbg, on)
		te.z0.n.drain()
		te.api.ZoneNotify(ctxbg, on.ID)
		te.z0.n.wait()
	})
}

func TestZoneUpdate(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		nz := z
		nz.ProviderConfigName = te.z1.pc.Name
		te.api.ZoneUpdate(ctxbg, nz)

		nz.ProviderConfigName = "doesnotexist"
		te.sherpaError("user:error", func() { te.api.ZoneUpdate(ctxbg, nz) })
	})
}

func TestProviderConfigs(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		te.api.ProviderConfigs(ctxbg)
	})
}

func TestProviderConfigUpdate(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		te.api.ProviderConfigUpdate(ctxbg, te.z0.pc)
	})
}

func TestZoneRecordSets(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		sets := te.api.ZoneRecordSets(ctxbg, z.Name)
		tcompare(t, len(sets), 7)

		te.sherpaError("user:notFound", func() { te.api.ZoneRecordSets(ctxbg, "bogus") })
	})
}
func TestZoneRecordSetHistory(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		hist := te.api.ZoneRecordSetHistory(ctxbg, z.Name, "testhost", Type(dns.TypeA))
		tcompare(t, len(hist), 2) // Negative and current.

		te.sherpaError("user:notFound", func() { te.api.ZoneRecordSetHistory(ctxbg, "bogus", "testhost", Type(dns.TypeA)) })
		te.sherpaError("user:error", func() { te.api.ZoneRecordSetHistory(ctxbg, te.z1.z.Name, "testhost", Type(dns.TypeA)) })
		te.sherpaError("user:error", func() { te.api.ZoneRecordSetHistory(ctxbg, z.Name, "bogus", Type(dns.TypeA)) })
		te.sherpaError("user:error", func() { te.api.ZoneRecordSetHistory(ctxbg, z.Name, "testhost", Type(dns.TypeAAAA)) })
	})
}
