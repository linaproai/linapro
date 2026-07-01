// This file verifies public API timestamp projection semantics for optional
// and absolute runtime instants.

package apitime

import (
	"testing"
	"time"
)

func TestMilliKeepsAbsoluteInstant(t *testing.T) {
	shanghai, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	value := time.Date(2026, time.June, 29, 17, 30, 8, 0, shanghai)

	actual := Milli(&value)
	if actual == nil {
		t.Fatal("expected timestamp projection")
	}
	if *actual != value.UnixMilli() {
		t.Fatalf("expected %d, got %d", value.UnixMilli(), *actual)
	}
}

func TestMilliKeepsNilAndZeroAsNil(t *testing.T) {
	if actual := Milli(nil); actual != nil {
		t.Fatalf("expected nil for absent timestamp, got %d", *actual)
	}
	zero := time.Time{}
	if actual := Milli(&zero); actual != nil {
		t.Fatalf("expected nil for zero timestamp, got %d", *actual)
	}
}

func TestMilliFromTimeKeepsAbsoluteInstant(t *testing.T) {
	shanghai, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	value := time.Date(2026, time.June, 29, 17, 30, 8, 0, shanghai)

	actual := MilliFromTime(value)
	if actual == nil {
		t.Fatal("expected timestamp projection")
	}
	if *actual != value.UnixMilli() {
		t.Fatalf("expected %d, got %d", value.UnixMilli(), *actual)
	}
}
