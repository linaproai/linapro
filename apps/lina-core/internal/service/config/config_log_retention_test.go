// This file verifies global log-retention runtime parameter handling.

package config

import (
	"context"
	"testing"
)

// TestGetLogRetentionDaysUsesRuntimeOverride verifies the shared log retention
// getter reads the protected runtime parameter snapshot.
func TestGetLogRetentionDaysUsesRuntimeOverride(t *testing.T) {
	withCachedRuntimeParamValue(t, RuntimeParamKeyLogRetentionDays, "120")

	days, err := New().GetLogRetentionDays(context.Background())
	if err != nil {
		t.Fatalf("get log retention days: %v", err)
	}
	if days != 120 {
		t.Fatalf("expected log retention days override 120, got %d", days)
	}
}

// TestGetLogRetentionDaysUsesDefaultWhenRuntimeMissing verifies the typed
// getter falls back to host default metadata when sys_config is absent.
func TestGetLogRetentionDaysUsesDefaultWhenRuntimeMissing(t *testing.T) {
	withCachedRuntimeParamSnapshot(t, &runtimeParamSnapshot{})

	days, err := New().GetLogRetentionDays(context.Background())
	if err != nil {
		t.Fatalf("get default log retention days: %v", err)
	}
	if days != 90 {
		t.Fatalf("expected default log retention days 90, got %d", days)
	}
}

// TestGetLogRetentionDaysReturnsInvalidRuntimeValue verifies malformed runtime
// values are visible to callers instead of silently falling back.
func TestGetLogRetentionDaysReturnsInvalidRuntimeValue(t *testing.T) {
	withCachedRuntimeParamValue(t, RuntimeParamKeyLogRetentionDays, "0")

	if _, err := New().GetLogRetentionDays(context.Background()); err == nil {
		t.Fatal("expected invalid log retention days override to return an error")
	}
}

// TestGetRawReturnsRuntimeProtectedParameter verifies source-plugin host config
// reads can observe protected runtime parameter values.
func TestGetRawReturnsRuntimeProtectedParameter(t *testing.T) {
	withCachedRuntimeParamValue(t, RuntimeParamKeyLogRetentionDays, "180")

	value, err := New().(*serviceImpl).GetRaw(context.Background(), RuntimeParamKeyLogRetentionDays)
	if err != nil {
		t.Fatalf("get raw protected runtime parameter: %v", err)
	}
	if value == nil || value.IsNil() {
		t.Fatal("expected protected runtime parameter value")
	}
	if got := value.Int(); got != 180 {
		t.Fatalf("expected raw protected runtime parameter 180, got %d", got)
	}
}

// TestGetRawReturnsLogRetentionDefaultWhenRuntimeMissing verifies plugin-facing
// host config raw reads use the generic default fallback for log retention.
func TestGetRawReturnsLogRetentionDefaultWhenRuntimeMissing(t *testing.T) {
	withCachedRuntimeParamSnapshot(t, &runtimeParamSnapshot{})

	value, err := New().(*serviceImpl).GetRaw(context.Background(), RuntimeParamKeyLogRetentionDays)
	if err != nil {
		t.Fatalf("get raw log retention default: %v", err)
	}
	if value == nil || value.String() != "90" {
		t.Fatalf("expected raw log retention default 90, got %#v", value)
	}
}
