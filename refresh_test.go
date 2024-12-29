package main

import (
	"log/slog"
	"testing"
)

func TestRefreshZoneSync(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		err := refreshZoneSync(slog.Default(), z)
		tcheck(t, err, "refresh: zone sync")
	})
}

func TestRefreshZoneSOACheck(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		err := refreshZoneSOACheck(slog.Default(), z)
		tcheck(t, err, "refresh: zone soa check")
	})
}
