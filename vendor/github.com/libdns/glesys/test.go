package glesys

import (
	"os"
	"testing"
)

func skipUnauth(t *testing.T) {
	if len(os.Getenv("GLESYS_PROJECT")) < 6 || len(os.Getenv("GLESYS_KEY")) < 10 || len(os.Getenv("GLESYS_ZONE")) < 4 {
		t.Skip("Skipping testing because missing credentials")
	}
}
