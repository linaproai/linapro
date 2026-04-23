// This file verifies plugin configuration defaults, legacy fallback handling,
// and test-time dynamic storage-path overrides.

package config

import (
	"context"
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
