package v1

import (
	"testing"
	"time"
)

func TestRefreshWindowShanghai(t *testing.T) {
	for _, test := range []struct {
		utc, day string
		allowed  bool
	}{
		{"2026-11-23T06:59:59Z", "2026-11-23", false},
		{"2026-11-23T07:00:00Z", "2026-11-23", true},
		{"2026-11-23T15:59:59Z", "2026-11-23", true},
		{"2026-11-23T16:00:00Z", "2026-11-24", false},
	} {
		now, err := time.Parse(time.RFC3339, test.utc)
		if err != nil {
			t.Fatal(err)
		}
		date, ok := refreshWindow(now)
		if date != test.day || ok != test.allowed {
			t.Fatalf("%s: %s %v", test.utc, date, ok)
		}
	}
}
