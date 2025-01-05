package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestRecordAdd(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		te.sherpaError("user:notFound", func() { te.api.RecordAdd(ctxbg, "bogus", RecordNew{"testhost", 600, Type(dns.TypeA), "10.0.0.3"}) })
		te.sherpaError("user:error", func() { te.api.RecordAdd(ctxbg, z.Name, RecordNew{"bad name", 600, Type(dns.TypeA), "10.0.0.3"}) })
		te.sherpaError("user:error", func() { te.api.RecordAdd(ctxbg, z.Name, RecordNew{"testhost", 600, 0, ""}) }) // Bad type

		tc := te.zoneChanged(func() {
			rn := RecordNew{"testhost", 600, Type(dns.TypeA), "10.0.0.3"}
			te.api.RecordAdd(ctxbg, z.Name, rn)
		})
		// Two existing A records with same name removed, and 3 new ones added.
		tc.checkRecordDelta(typecounts{"A": 2}, typecounts{"A": 3})
	})
}

func TestRecordDelete(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		tc := te.zoneChanged(func() {
			for _, r := range te.z0.records {
				if r.Type == Type(dns.TypeA) {
					te.sherpaError("user:notFound", func() { te.api.RecordDelete(ctxbg, "bogus", r.ID) })
					te.sherpaError("user:notFound", func() { te.api.RecordDelete(ctxbg, z.Name, r.ID+999) })
					te.sherpaError("user:notFound", func() { te.api.RecordDelete(ctxbg, te.z1.z.Name, r.ID) }) // Zone mismatch.

					te.api.RecordDelete(ctxbg, z.Name, r.ID)
					te.sherpaError("user:error", func() { te.api.RecordDelete(ctxbg, z.Name, r.ID) })
					return
				}
			}
			t.Fatalf("did not find existing A record")
		})
		tc.checkRecordDelta(typecounts{"A": 2}, typecounts{"A": 1})
	})
}

func TestRecordUpdate(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		tc := te.zoneChanged(func() {
			for _, r := range te.z0.records {
				if r.Type == Type(dns.TypeA) {
					rn := RecordNew{"testhost", 600, Type(dns.TypeA), "10.0.0.3"}

					te.sherpaError("user:notFound", func() { te.api.RecordUpdate(ctxbg, "bogus", r.ID, rn) })
					te.sherpaError("user:notFound", func() { te.api.RecordUpdate(ctxbg, z.Name, r.ID+999, rn) })
					te.sherpaError("user:notFound", func() { te.api.RecordUpdate(ctxbg, te.z1.z.Name, r.ID, rn) })                               // Zone mismatch.
					te.sherpaError("user:error", func() { te.api.RecordUpdate(ctxbg, z.Name, r.ID, RecordNew{"testhost", 600, 0, "10.0.0.3"}) }) // Bad type.

					te.api.RecordUpdate(ctxbg, z.Name, r.ID, rn)
					return
				}
			}
			t.Fatalf("did not find existing A record")
		})
		tc.checkRecordDelta(typecounts{"A": 2}, typecounts{"A": 2})
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
			rn := RecordNew{"testhost", 600, Type(dns.TypeA), "10.0.0.2"}
			te.api.RecordAdd(ctxbg, z.Name, rn)
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
