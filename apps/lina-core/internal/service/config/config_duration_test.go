// This file verifies shared duration-based configuration parsing helpers.

package config

import (
	"strings"
	"testing"
	"time"
)

// TestMustParsePositiveDuration verifies duration helper parsing and positive
// value enforcement.
func TestMustParsePositiveDuration(t *testing.T) {
	if duration := mustParsePositiveDuration("test.duration", "2m"); duration != 2*time.Minute {
		t.Fatalf("expected duration 2m, got %s", duration)
	}

	func() {
		defer assertConfigPanicContains(t, "greater than 0")
		mustParsePositiveDuration("test.duration", "0s")
	}()
}

// TestMustValidateSecondAlignedDuration verifies second alignment validation
// rejects sub-second or fractional-second durations.
func TestMustValidateSecondAlignedDuration(t *testing.T) {
	if duration := mustValidateSecondAlignedDuration("test.duration", 2*time.Second); duration != 2*time.Second {
		t.Fatalf("expected duration 2s, got %s", duration)
	}

	func() {
		defer assertConfigPanicContains(t, "whole seconds")
		mustValidateSecondAlignedDuration("test.duration", 1500*time.Millisecond)
	}()
}

// assertConfigPanicContains verifies the current deferred panic contains the expected text.
func assertConfigPanicContains(t *testing.T, expected string) {
	t.Helper()

	recovered := recover()
	if recovered == nil {
		t.Fatalf("expected panic containing %q, but no panic occurred", expected)
	}
	if !strings.Contains(recovered.(error).Error(), expected) {
		t.Fatalf("expected panic containing %q, got %v", expected, recovered)
	}
}
