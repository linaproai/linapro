// This file verifies global log-retention runtime parameter handling.

package config

import (
	"context"
	"testing"

	"lina-core/pkg/bizerr"
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

// TestGetLogRetentionDaysRequiresSeedRow verifies the delivery SQL seed row is
// required instead of hidden behind a synthetic default.
func TestGetLogRetentionDaysRequiresSeedRow(t *testing.T) {
	withCachedRuntimeParamSnapshot(t, &runtimeParamSnapshot{})

	_, err := New().GetLogRetentionDays(context.Background())
	if !bizerr.Is(err, CodeConfigParamRequired) {
		t.Fatalf("expected required config error, got %v", err)
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

// TestGetRawRequiresLogRetentionSeedRow verifies plugin-facing host config
// reads do not synthesize the log-retention value when the seed row is absent.
func TestGetRawRequiresLogRetentionSeedRow(t *testing.T) {
	withCachedRuntimeParamSnapshot(t, &runtimeParamSnapshot{})

	_, err := New().(*serviceImpl).GetRaw(context.Background(), RuntimeParamKeyLogRetentionDays)
	if !bizerr.Is(err, CodeConfigParamRequired) {
		t.Fatalf("expected required config error, got %v", err)
	}
}
