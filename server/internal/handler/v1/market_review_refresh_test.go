package v1

import (
	"testing"
	"time"
)

func TestReviewDateShanghai(t *testing.T) {
	for _, test := range []struct {
		utc, day string
	}{
		{"2026-11-23T06:59:59Z", "2026-11-23"},
		{"2026-11-23T07:00:00Z", "2026-11-23"},
		{"2026-11-23T15:59:59Z", "2026-11-23"},
		{"2026-11-23T16:00:00Z", "2026-11-24"},
	} {
		now, err := time.Parse(time.RFC3339, test.utc)
		if err != nil {
			t.Fatal(err)
		}
		date := reviewLocalTime(now).Format(time.DateOnly)
		if date != test.day {
			t.Fatalf("%s: %s", test.utc, date)
		}
	}
}

func TestReviewSchedule(t *testing.T) {
	for _, test := range []struct{ utc, phase string }{
		{"2026-11-22T23:59:59Z", ""},
		{"2026-11-23T00:00:00Z", "warm"},
		{"2026-11-23T07:04:59Z", ""},
		{"2026-11-23T07:05:00Z", "close"},
		{"2026-11-23T11:00:00Z", ""},
		{"2026-11-23T14:00:00Z", ""},
		{"2026-11-22T07:30:00Z", ""},
	} {
		now, err := time.Parse(time.RFC3339, test.utc)
		if err != nil {
			t.Fatal(err)
		}
		if got := reviewSchedule(reviewLocalTime(now)); got != test.phase {
			t.Fatalf("%s phase %q, want %q", test.utc, got, test.phase)
		}
	}
}
