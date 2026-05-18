// This file verifies the generic read-only plugin configuration service.

package config

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
)

// scanTarget captures nested test configuration values.
type scanTarget struct {
	// Name is a sample string value.
	Name string `json:"name"`
	// Enabled is a sample boolean value.
	Enabled bool `json:"enabled"`
	// Count is a sample integer value.
	Count int `json:"count"`
}

// TestGetReturnsRawConfigForAnyKey verifies callers can read arbitrary config keys.
func TestGetReturnsRawConfigForAnyKey(t *testing.T) {
	setTestConfigAdapter(t, `
custom:
  name: demo
`)

	value, err := New().Get(context.Background(), "custom.name")
	if err != nil {
		t.Fatalf("get config value: %v", err)
	}
	if value == nil {
		t.Fatal("expected config value")
	}
	if got := value.String(); got != "demo" {
		t.Fatalf("expected custom.name to be demo, got %q", got)
	}
}

// TestExistsReportsConfiguredFalseAndZero verifies existence does not treat false or zero as missing.
func TestExistsReportsConfiguredFalseAndZero(t *testing.T) {
	setTestConfigAdapter(t, `
feature:
  enabled: false
  retries: 0
`)

	svc := New()
	ctx := context.Background()

	exists, err := svc.Exists(ctx, "feature.enabled")
	if err != nil {
		t.Fatalf("check bool key exists: %v", err)
	}
	if !exists {
		t.Fatal("expected feature.enabled to exist")
	}

	exists, err = svc.Exists(ctx, "feature.retries")
	if err != nil {
		t.Fatalf("check int key exists: %v", err)
	}
	if !exists {
		t.Fatal("expected feature.retries to exist")
	}

	exists, err = svc.Exists(ctx, "feature.missing")
	if err != nil {
		t.Fatalf("check missing key exists: %v", err)
	}
	if exists {
		t.Fatal("expected missing key to not exist")
	}
}

// TestScanBindsConfigSection verifies section scanning into caller-owned structs.
func TestScanBindsConfigSection(t *testing.T) {
	setTestConfigAdapter(t, `
custom:
  name: demo
  enabled: false
  count: 0
`)

	target := &scanTarget{}
	if err := New().Scan(context.Background(), "custom", target); err != nil {
		t.Fatalf("scan config section: %v", err)
	}

	if target.Name != "demo" {
		t.Fatalf("expected name demo, got %q", target.Name)
	}
	if target.Enabled {
		t.Fatal("expected enabled to be false")
	}
	if target.Count != 0 {
		t.Fatalf("expected count 0, got %d", target.Count)
	}
}

// TestDefaultsForMissingAndBlankKeys verifies default handling for absent and blank values.
func TestDefaultsForMissingAndBlankKeys(t *testing.T) {
	setTestConfigAdapter(t, `
strings:
  blank: ""
feature:
  enabled: false
  retries: 0
duration:
  blank: ""
`)

	svc := New()
	ctx := context.Background()

	text, err := svc.String(ctx, "strings.blank", "fallback")
	if err != nil {
		t.Fatalf("read blank string: %v", err)
	}
	if text != "fallback" {
		t.Fatalf("expected blank string default, got %q", text)
	}

	enabled, err := svc.Bool(ctx, "feature.enabled", true)
	if err != nil {
		t.Fatalf("read bool: %v", err)
	}
	if enabled {
		t.Fatal("expected configured false bool to override default true")
	}

	retries, err := svc.Int(ctx, "feature.retries", 3)
	if err != nil {
		t.Fatalf("read int: %v", err)
	}
	if retries != 0 {
		t.Fatalf("expected configured zero int to override default, got %d", retries)
	}

	interval, err := svc.Duration(ctx, "duration.blank", time.Minute)
	if err != nil {
		t.Fatalf("read blank duration: %v", err)
	}
	if interval != time.Minute {
		t.Fatalf("expected blank duration default, got %s", interval)
	}
}

// TestDurationParsesDurationString verifies duration strings are parsed through the generic service.
func TestDurationParsesDurationString(t *testing.T) {
	setTestConfigAdapter(t, `
duration:
  interval: 45s
`)

	interval, err := New().Duration(context.Background(), "duration.interval", time.Minute)
	if err != nil {
		t.Fatalf("read duration: %v", err)
	}
	if interval != 45*time.Second {
		t.Fatalf("expected 45s duration, got %s", interval)
	}
}

// TestDurationReturnsErrorForInvalidValue verifies invalid duration strings return errors.
func TestDurationReturnsErrorForInvalidValue(t *testing.T) {
	setTestConfigAdapter(t, `
duration:
  interval: invalid
`)

	_, err := New().Duration(context.Background(), "duration.interval", time.Minute)
	if err == nil {
		t.Fatal("expected invalid duration error")
	}
	if !strings.Contains(err.Error(), "duration.interval") {
		t.Fatalf("expected error to mention key, got %v", err)
	}
}

// TestScanRejectsNilTarget verifies scan calls cannot silently ignore nil targets.
func TestScanRejectsNilTarget(t *testing.T) {
	setTestConfigAdapter(t, `
custom:
  name: demo
`)

	err := New().Scan(context.Background(), "custom", nil)
	if err == nil {
		t.Fatal("expected nil target error")
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
