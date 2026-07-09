// This file verifies online-session configuration loading and runtime
// overrides.

package config

import (
	"context"
	"testing"
	"time"
)

// TestGetSessionUsesDefaultWhenUnset verifies session config falls back to its
// defaults when static config and runtime overrides are absent.
func TestGetSessionUsesDefaultWhenUnset(t *testing.T) {
	setTestConfigContent(t, `
database:
  default:
    link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
`)
	withRuntimeParamAbsent(t, RuntimeParamKeySessionTimeout)

	cfg, err := New().GetSession(context.Background())
	if err != nil {
		t.Fatalf("get session config: %v", err)
	}

	if cfg.Timeout != 24*time.Hour {
		t.Fatalf("expected default session timeout to be 24h, got %s", cfg.Timeout)
	}
	if cfg.CleanupInterval != 5*time.Minute {
		t.Fatalf("expected default session cleanup interval to be 5m, got %s", cfg.CleanupInterval)
	}
}

// TestGetSessionUsesDurationConfig verifies session duration settings come from
// static config.
func TestGetSessionUsesDurationConfig(t *testing.T) {
	setTestConfigContent(t, `
database:
  default:
    link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
session:
  timeout: 36h
  cleanupInterval: 10m
`)
	withRuntimeParamAbsent(t, RuntimeParamKeySessionTimeout)

	svc := New()
	cfg, err := svc.GetSession(context.Background())
	if err != nil {
		t.Fatalf("get session config: %v", err)
	}

	if cfg.Timeout != 36*time.Hour {
		t.Fatalf("expected session timeout to be 36h, got %s", cfg.Timeout)
	}
	if cfg.CleanupInterval != 10*time.Minute {
		t.Fatalf("expected session cleanup interval to be 10m, got %s", cfg.CleanupInterval)
	}
	timeout, err := svc.GetSessionTimeout(context.Background())
	if err != nil {
		t.Fatalf("get session timeout: %v", err)
	}
	if timeout != 36*time.Hour {
		t.Fatalf("expected GetSessionTimeout to be 36h, got %s", timeout)
	}
}

// TestGetSessionPrefersRuntimeParamTimeout verifies the session timeout can be
// overridden by runtime parameters without disturbing static cleanup interval.
func TestGetSessionPrefersRuntimeParamTimeout(t *testing.T) {
	withRuntimeParamValue(t, RuntimeParamKeySessionTimeout, "2h")

	svc := New()
	cfg, err := svc.GetSession(context.Background())
	if err != nil {
		t.Fatalf("get session config: %v", err)
	}

	if cfg.Timeout != 2*time.Hour {
		t.Fatalf("expected runtime param session timeout to be 2h, got %s", cfg.Timeout)
	}
	if cfg.CleanupInterval <= 0 {
		t.Fatalf("expected cleanup interval to remain positive, got %s", cfg.CleanupInterval)
	}
	timeout, err := svc.GetSessionTimeout(context.Background())
	if err != nil {
		t.Fatalf("get session timeout: %v", err)
	}
	if timeout != 2*time.Hour {
		t.Fatalf("expected runtime getter session timeout to be 2h, got %s", timeout)
	}
}

// TestGetSessionRejectsNonSecondAlignedCleanupInterval verifies invalid
// fractional-second cleanup intervals panic during config load.
func TestGetSessionRejectsNonSecondAlignedCleanupInterval(t *testing.T) {
	setTestConfigContent(t, `
session:
  cleanupInterval: 1500ms
`)

	defer assertConfigPanicContains(t, "whole seconds")

	cfg, err := New().GetSession(context.Background())
	if err != nil {
		t.Fatalf("get session config after invalid static config: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected session config")
	}
}
