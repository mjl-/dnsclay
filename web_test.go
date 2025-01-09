package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/libdns/libdns"
	"github.com/miekg/dns"
)

func TestWebAuthExport(t *testing.T) {
	mux := makeAdminMux()

	testAuth := func(path string) {
		// Missing credentials.
		req := httptest.NewRequest("GET", path, nil)
		rec := httptest.ResponseRecorder{}
		mux.ServeHTTP(&rec, req)
		tcompare(t, rec.Code, http.StatusUnauthorized)

		// Bad password.
		req.SetBasicAuth("admin", "bad")
		rec = httptest.ResponseRecorder{}
		mux.ServeHTTP(&rec, req)
		tcompare(t, rec.Code, http.StatusUnauthorized)

		// Allowed.
		req.SetBasicAuth("admin", adminpassword)
		rec = httptest.ResponseRecorder{}
		mux.ServeHTTP(&rec, req)
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
}

func TestRecordSetUpdate(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		rsn := RecordSetChange{"testhost", 600, Type(dns.TypeA), []string{"10.0.0.1", "10.0.0.3"}}

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
		te.sherpaError("user:error", func() { te.api.ProviderConfigTest(ctxbg, z.Name, "badprovider", te.z0.pc.ProviderConfigJSON) })
		te.sherpaError("user:error", func() { te.api.ProviderConfigTest(ctxbg, z.Name, "fake", "bad json") })
		te.sherpaError("user:error", func() { te.api.ProviderConfigTest(ctxbg, z.Name, "fake", `{"ID": "unknownzone"}`) })

		te.api.ProviderConfigTest(ctxbg, z.Name, "fake", te.z0.pc.ProviderConfigJSON)
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
		tcompare(t, len(sets), 2) // SOA and A.

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
