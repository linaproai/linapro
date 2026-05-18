// This file verifies host runtime operation configuration defaults and overrides.

package config

import (
	"context"
	"testing"
	"time"
)

// TestRuntimeOperationConfigsUseDefaultsWhenUnset verifies newly added
// operation settings retain safe defaults when config.yaml omits them.
func TestRuntimeOperationConfigsUseDefaultsWhenUnset(t *testing.T) {
	setTestConfigContent(t, ``)

	svc := New()
	healthCfg := svc.GetHealth(context.Background())
	shutdownCfg := svc.GetShutdown(context.Background())
	schedulerCfg := svc.GetScheduler(context.Background())

	if healthCfg.Timeout != 5*time.Second {
		t.Fatalf("expected default health timeout to be 5s, got %s", healthCfg.Timeout)
	}
	if shutdownCfg.Timeout != 30*time.Second {
		t.Fatalf("expected default shutdown timeout to be 30s, got %s", shutdownCfg.Timeout)
	}
	if schedulerCfg.DefaultTimezone != "UTC" {
		t.Fatalf("expected default scheduler timezone to be UTC, got %q", schedulerCfg.DefaultTimezone)
	}
	if timezone := svc.GetSchedulerDefaultTimezone(context.Background()); timezone != "UTC" {
		t.Fatalf("expected default scheduler timezone getter to return UTC, got %q", timezone)
	}
}

// TestRuntimeOperationConfigsUseStaticConfig verifies health, shutdown, and
// scheduler settings are loaded from the static config file.
func TestRuntimeOperationConfigsUseStaticConfig(t *testing.T) {
	setTestConfigContent(t, `
health:
  timeout: 5s
shutdown:
  timeout: 45s
scheduler:
  defaultTimezone: "Europe/Berlin"
`)

	svc := New()
	healthCfg := svc.GetHealth(context.Background())
	shutdownCfg := svc.GetShutdown(context.Background())
	schedulerCfg := svc.GetScheduler(context.Background())

	if healthCfg.Timeout != 5*time.Second {
		t.Fatalf("expected health timeout to be 5s, got %s", healthCfg.Timeout)
	}
	if shutdownCfg.Timeout != 45*time.Second {
		t.Fatalf("expected shutdown timeout to be 45s, got %s", shutdownCfg.Timeout)
	}
	if schedulerCfg.DefaultTimezone != "Europe/Berlin" {
		t.Fatalf("expected scheduler timezone to be Europe/Berlin, got %q", schedulerCfg.DefaultTimezone)
	}
	if timezone := svc.GetSchedulerDefaultTimezone(context.Background()); timezone != "Europe/Berlin" {
		t.Fatalf("expected scheduler timezone getter to return Europe/Berlin, got %q", timezone)
	}
}
