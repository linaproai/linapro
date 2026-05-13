// This file verifies monitor-server plugin configuration loading.

package config

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"

	configsvc "lina-core/pkg/pluginservice/config"
)

// TestLoadUsesDefaultsWhenUnset verifies monitor config defaults when config is absent.
func TestLoadUsesDefaultsWhenUnset(t *testing.T) {
	setTestConfigAdapter(t, `
server:
  address: ":8080"
`)

	cfg, err := Load(context.Background(), configsvc.New())
	if err != nil {
		t.Fatalf("load monitor config: %v", err)
	}
	if cfg.Interval != time.Minute {
		t.Fatalf("expected default interval 1m, got %s", cfg.Interval)
	}
	if cfg.RetentionMultiplier != 5 {
		t.Fatalf("expected default retention multiplier 5, got %d", cfg.RetentionMultiplier)
	}
}

// TestLoadUsesConfiguredValues verifies configured monitor values override defaults.
func TestLoadUsesConfiguredValues(t *testing.T) {
	setTestConfigAdapter(t, `
monitor:
  interval: 45s
  retentionMultiplier: 8
`)

	cfg, err := Load(context.Background(), configsvc.New())
	if err != nil {
		t.Fatalf("load monitor config: %v", err)
	}
	if cfg.Interval != 45*time.Second {
		t.Fatalf("expected interval 45s, got %s", cfg.Interval)
	}
	if cfg.RetentionMultiplier != 8 {
		t.Fatalf("expected retention multiplier 8, got %d", cfg.RetentionMultiplier)
	}
}

// TestLoadReturnsErrorForInvalidDuration verifies invalid duration strings fail config loading.
func TestLoadReturnsErrorForInvalidDuration(t *testing.T) {
	setTestConfigAdapter(t, `
monitor:
  interval: invalid
`)

	_, err := Load(context.Background(), configsvc.New())
	if err == nil {
		t.Fatal("expected invalid duration error")
	}
	if !strings.Contains(err.Error(), "monitor.interval") {
		t.Fatalf("expected error to mention monitor.interval, got %v", err)
	}
}

// TestLoadRejectsSubSecondInterval verifies monitor interval lower bound validation.
func TestLoadRejectsSubSecondInterval(t *testing.T) {
	setTestConfigAdapter(t, `
monitor:
  interval: 500ms
`)

	_, err := Load(context.Background(), configsvc.New())
	if err == nil {
		t.Fatal("expected sub-second interval error")
	}
	if !strings.Contains(err.Error(), "at least 1s") {
		t.Fatalf("expected at least 1s error, got %v", err)
	}
}

// TestLoadRejectsFractionalSecondInterval verifies monitor interval alignment validation.
func TestLoadRejectsFractionalSecondInterval(t *testing.T) {
	setTestConfigAdapter(t, `
monitor:
  interval: 1500ms
`)

	_, err := Load(context.Background(), configsvc.New())
	if err == nil {
		t.Fatal("expected fractional-second interval error")
	}
	if !strings.Contains(err.Error(), "whole seconds") {
		t.Fatalf("expected whole seconds error, got %v", err)
	}
}

// setTestConfigAdapter swaps the process config adapter for one test case.
func setTestConfigAdapter(t *testing.T, content string) {
	t.Helper()

	adapter, err := gcfg.NewAdapterContent(content)
	if err != nil {
		t.Fatalf("create content adapter: %v", err)
	}

	originalAdapter := g.Cfg().GetAdapter()
	g.Cfg().SetAdapter(adapter)

	t.Cleanup(func() {
		g.Cfg().SetAdapter(originalAdapter)
	})
}
