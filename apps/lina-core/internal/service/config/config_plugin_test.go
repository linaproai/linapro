// This file verifies plugin configuration defaults, legacy fallback handling,
// and test-time dynamic storage-path overrides.

package config

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
)

// TestGetPluginUsesDefaultStoragePathAndClonesCachedConfig verifies callers
// receive detached plugin config copies even when the package uses defaults.
func TestGetPluginUsesDefaultStoragePathAndClonesCachedConfig(t *testing.T) {
	setTestConfigContent(t, "test: 1\n")
	SetPluginDynamicStoragePathOverride("")
	t.Cleanup(func() {
		SetPluginDynamicStoragePathOverride("")
	})

	svc := New()
	cfg := svc.GetPlugin(context.Background())
	if cfg.Dynamic.StoragePath != "temp/output" {
		t.Fatalf("expected default dynamic storage path, got %q", cfg.Dynamic.StoragePath)
	}

	cfg.Dynamic.StoragePath = "mutated/by/test"
	refreshed := svc.GetPlugin(context.Background())
	if refreshed.Dynamic.StoragePath != "temp/output" {
		t.Fatalf("expected cached plugin config to stay immutable, got %q", refreshed.Dynamic.StoragePath)
	}
}

// TestGetPluginFallsBackToLegacyRuntimeStoragePath verifies the legacy runtime
// field still provides the effective storage path when the new field is empty.
func TestGetPluginFallsBackToLegacyRuntimeStoragePath(t *testing.T) {
	setTestConfigContent(t, `
plugin:
  dynamic:
    storagePath: "   "
  runtime:
    storagePath: legacy/runtime/plugins
`)
	SetPluginDynamicStoragePathOverride("")
	t.Cleanup(func() {
		SetPluginDynamicStoragePathOverride("")
	})

	svc := New()
	cfg := svc.GetPlugin(context.Background())
	if cfg.Dynamic.StoragePath != "legacy/runtime/plugins" {
		t.Fatalf("expected legacy runtime storage path fallback, got %q", cfg.Dynamic.StoragePath)
	}
	if path := svc.GetPluginDynamicStoragePath(context.Background()); path != filepath.Clean("legacy/runtime/plugins") {
		t.Fatalf("expected cleaned legacy runtime storage path, got %q", path)
	}
}

// TestGetPluginDynamicStoragePathUsesDefaultAndOverride verifies the default
// path is exposed when config is absent and test overrides take precedence.
func TestGetPluginDynamicStoragePathUsesDefaultAndOverride(t *testing.T) {
	setTestConfigContent(t, "test: 1\n")
	SetPluginDynamicStoragePathOverride("")
	t.Cleanup(func() {
		SetPluginDynamicStoragePathOverride("")
	})

	svc := New()
	if path := svc.GetPluginDynamicStoragePath(context.Background()); path != filepath.Clean("temp/output") {
		t.Fatalf("expected default plugin storage path temp/output, got %q", path)
	}

	SetPluginDynamicStoragePathOverride(" ./temp/output/../plugin-bundles ")
	if path := svc.GetPluginDynamicStoragePath(context.Background()); path != filepath.Clean("./temp/output/../plugin-bundles") {
		t.Fatalf("expected override storage path to win, got %q", path)
	}
}

// TestGetPluginDynamicStoragePathOverrideIgnoresBlankValues verifies blank
// overrides are treated as absent so callers can fall back safely.
func TestGetPluginDynamicStoragePathOverrideIgnoresBlankValues(t *testing.T) {
	SetPluginDynamicStoragePathOverride("   ")
	t.Cleanup(func() {
		SetPluginDynamicStoragePathOverride("")
	})

	if path := getPluginDynamicStoragePathOverride(); path != "" {
		t.Fatalf("expected blank override to be treated as absent, got %q", path)
	}
}

// TestGetPluginAutoEnableNormalizesListAndAppliesOverrides verifies startup
// auto-enable IDs are trimmed, de-duplicated, cloned, and overrideable in tests.
func TestGetPluginAutoEnableNormalizesListAndAppliesOverrides(t *testing.T) {
	setTestConfigContent(t, `
plugin:
  autoEnable:
    - " plugin-demo-source "
    - "plugin-demo-source"
    - "plugin-demo-dynamic"
`)
	SetPluginAutoEnableOverride(nil)
	t.Cleanup(func() {
		SetPluginAutoEnableOverride(nil)
	})

	svc := New()
	cfg := svc.GetPlugin(context.Background())
	if len(cfg.AutoEnable) != 2 {
		t.Fatalf("expected two normalized plugin IDs, got %#v", cfg.AutoEnable)
	}
	if cfg.AutoEnable[0] != "plugin-demo-source" || cfg.AutoEnable[1] != "plugin-demo-dynamic" {
		t.Fatalf("expected normalized auto-enable IDs, got %#v", cfg.AutoEnable)
	}

	cfg.AutoEnable[0] = "mutated"
	refreshed := svc.GetPlugin(context.Background())
	if refreshed.AutoEnable[0] != "plugin-demo-source" {
		t.Fatalf("expected cached auto-enable IDs to stay immutable, got %#v", refreshed.AutoEnable)
	}

	SetPluginAutoEnableOverride([]string{" override-plugin ", "override-plugin", "second-plugin"})
	overridden := svc.GetPlugin(context.Background())
	if len(overridden.AutoEnable) != 2 {
		t.Fatalf("expected override to replace auto-enable IDs, got %#v", overridden.AutoEnable)
	}
	if overridden.AutoEnable[0] != "override-plugin" || overridden.AutoEnable[1] != "second-plugin" {
		t.Fatalf("expected normalized override IDs, got %#v", overridden.AutoEnable)
	}

	autoEnable := svc.GetPluginAutoEnable(context.Background())
	if len(autoEnable) != 2 {
		t.Fatalf("expected cloned auto-enable list, got %#v", autoEnable)
	}
	autoEnable[0] = "mutated-again"
	reloaded := svc.GetPluginAutoEnable(context.Background())
	if reloaded[0] != "override-plugin" {
		t.Fatalf("expected GetPluginAutoEnable to clone result, got %#v", reloaded)
	}
}

// TestGetPluginRejectsBlankAutoEnableEntry verifies plugin.autoEnable rejects
// blank plugin IDs during host startup.
func TestGetPluginRejectsBlankAutoEnableEntry(t *testing.T) {
	setTestConfigContent(t, `
plugin:
  autoEnable:
    - "plugin-demo-source"
    - "   "
`)

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected blank plugin.autoEnable entry to panic")
		}
		if message := fmt.Sprint(recovered); message == "" {
			t.Fatal("expected blank plugin.autoEnable panic message")
		}
	}()

	_ = New().GetPlugin(context.Background())
}

// TestGetPluginRejectsInvalidAutoEnableType verifies plugin.autoEnable must be
// configured as a string array instead of a scalar value.
func TestGetPluginRejectsInvalidAutoEnableType(t *testing.T) {
	setTestConfigContent(t, `
plugin:
  autoEnable: "plugin-demo-source"
`)

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected invalid plugin.autoEnable type to panic")
		}
	}()

	_ = New().GetPlugin(context.Background())
}
