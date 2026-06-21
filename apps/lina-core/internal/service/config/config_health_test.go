// This file verifies health-probe configuration loading.

package config

import (
	"context"
	"testing"
	"time"
)

// TestGetHealthUsesDefaultWhenUnset verifies health settings retain safe
// defaults when config.yaml omits them.
func TestGetHealthUsesDefaultWhenUnset(t *testing.T) {
	setTestConfigContent(t, ``)

	cfg := New().GetHealth(context.Background())

	if cfg.Timeout != 5*time.Second {
		t.Fatalf("expected default health timeout to be 5s, got %s", cfg.Timeout)
	}
}

// TestGetHealthUsesStaticConfig verifies health settings are loaded from the
// static config file.
func TestGetHealthUsesStaticConfig(t *testing.T) {
	setTestConfigContent(t, `
health:
  timeout: 5s
`)

	cfg := New().GetHealth(context.Background())

	if cfg.Timeout != 5*time.Second {
		t.Fatalf("expected health timeout to be 5s, got %s", cfg.Timeout)
	}
}
