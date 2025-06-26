package main

import (
	"testing"
	"time"
)

func TestHumanDate(t *testing.T) {
	tm := time.Date(2025, 6, 26, 10, 15, 0, 0, time.UTC)
	hd := HumanDate(tm)

	if hd != "26 Jun 2025 at 10:15" {
		t.Errorf("want %q; got %q", "26 Jun 2025 at 10:15", hd)
	}
}
