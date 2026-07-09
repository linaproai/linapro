// This file verifies scheduler configuration loading for host-managed jobs.

package config

import (
	"context"
	"testing"
)

// TestGetSchedulerUsesDefaultWhenUnset verifies scheduler settings retain safe
// defaults when config.yaml omits them.
func TestGetSchedulerUsesDefaultWhenUnset(t *testing.T) {
	setTestConfigContent(t, ``)

	svc := New()
	cfg := svc.GetScheduler(context.Background())

	if cfg.DefaultTimezone != "UTC" {
		t.Fatalf("expected default scheduler timezone to be UTC, got %q", cfg.DefaultTimezone)
	}
	if timezone := svc.GetSchedulerDefaultTimezone(context.Background()); timezone != "UTC" {
		t.Fatalf("expected default scheduler timezone getter to return UTC, got %q", timezone)
	}
}

// TestGetSchedulerUsesStaticConfig verifies scheduler settings are loaded from
// the static config file.
func TestGetSchedulerUsesStaticConfig(t *testing.T) {
	setTestConfigContent(t, `
scheduler:
  defaultTimezone: "Europe/Berlin"
`)

	svc := New()
	cfg := svc.GetScheduler(context.Background())

	if cfg.DefaultTimezone != "Europe/Berlin" {
		t.Fatalf("expected scheduler timezone to be Europe/Berlin, got %q", cfg.DefaultTimezone)
	}
	if timezone := svc.GetSchedulerDefaultTimezone(context.Background()); timezone != "Europe/Berlin" {
		t.Fatalf("expected scheduler timezone getter to return Europe/Berlin, got %q", timezone)
	}
}
