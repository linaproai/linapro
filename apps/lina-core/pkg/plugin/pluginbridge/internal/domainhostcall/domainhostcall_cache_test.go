// This file verifies cache host-service client helper behavior that remains
// internal to the domainhostcall implementation.

package domainhostcall

import (
	"testing"
	"time"
)

// TestCacheDurationSecondsPreservesTTLSign verifies cache TTLs are encoded in
// whole seconds without hiding negative durations from host-side validation.
func TestCacheDurationSecondsPreservesTTLSign(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     int64
	}{
		{name: "zero", duration: 0, want: 0},
		{name: "round positive up", duration: 1500 * time.Millisecond, want: 2},
		{name: "round negative away from zero", duration: -1500 * time.Millisecond, want: -2},
		{name: "minimum negative duration", duration: time.Duration(-1 << 63), want: -9223372037},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := durationSeconds(tt.duration); got != tt.want {
				t.Fatalf("durationSeconds(%s) = %d, want %d", tt.duration, got, tt.want)
			}
		})
	}
}
