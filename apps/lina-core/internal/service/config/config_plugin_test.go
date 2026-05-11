// This file verifies plugin configuration defaults, legacy fallback handling,
// and test-time dynamic storage-path overrides.

package config

import (
	"context"
	"fmt"
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
	if !cfg.AllowForceUninstall {
		t.Fatal("expected force uninstall to be enabled by default")
	}
	if cfg.Dynamic.StoragePath != "temp/output" {
		t.Fatalf("expected default dynamic storage path, got %q", cfg.Dynamic.StoragePath)
	}
	if len(cfg.AutoEnable) != 0 {
		t.Fatalf("expected default auto-enable list to stay empty, got %#v", cfg.AutoEnable)
	}
	if autoEnable := svc.GetPluginAutoEnable(context.Background()); autoEnable != nil {
		t.Fatalf("expected default GetPluginAutoEnable result to be nil, got %#v", autoEnable)
	}

	cfg.Dynamic.StoragePath = "mutated/by/test"
	refreshed := svc.GetPlugin(context.Background())
	if refreshed.Dynamic.StoragePath != "temp/output" {
		t.Fatalf("expected cached plugin config to stay immutable, got %q", refreshed.Dynamic.StoragePath)
	}
}

// TestGetPluginAllowsStrictForceUninstallOptOut verifies operators can
// explicitly close the force-uninstall channel for strict compliance
// deployments even though development defaults keep it enabled.
func TestGetPluginAllowsStrictForceUninstallOptOut(t *testing.T) {
	setTestConfigContent(t, `
plugin:
  allowForceUninstall: false
`)
	SetPluginAllowForceUninstallOverride(nil)
	t.Cleanup(func() {
		SetPluginAllowForceUninstallOverride(nil)
	})

	cfg := New().GetPlugin(context.Background())
	if cfg.AllowForceUninstall {
		t.Fatal("expected explicit plugin.allowForceUninstall=false to be honored")
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
	if path := svc.GetPluginDynamicStoragePath(context.Background()); path != resolveRuntimePath("legacy/runtime/plugins") {
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
	if path := svc.GetPluginDynamicStoragePath(context.Background()); path != resolveRuntimePath("temp/output") {
		t.Fatalf("expected default plugin storage path temp/output, got %q", path)
	}

	SetPluginDynamicStoragePathOverride(" ./temp/output/../plugin-bundles ")
	if path := svc.GetPluginDynamicStoragePath(context.Background()); path != resolveRuntimePath("./temp/output/../plugin-bundles") {
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
	// plugin.autoEnable accepts only the structured object form. The previous
	// bare-string form has been removed for schema uniformity, so duplicate
	// detection now exercises duplicate ids in the structured form together
	// with whitespace-trimming and order preservation.
	setTestConfigContent(t, `
plugin:
  autoEnable:
    - id: " demo-control "
    - id: "demo-control"
    - id: " plugin-demo-source "
    - id: "plugin-demo-source"
    - id: "plugin-demo-dynamic"
`)
	SetPluginAutoEnableOverride(nil)
	t.Cleanup(func() {
		SetPluginAutoEnableOverride(nil)
	})

	svc := New()
	cfg := svc.GetPlugin(context.Background())
	if len(cfg.AutoEnable) != 3 {
		t.Fatalf("expected three normalized plugin IDs, got %#v", cfg.AutoEnable)
	}
	if cfg.AutoEnable[0].ID != "demo-control" || cfg.AutoEnable[1].ID != "plugin-demo-source" || cfg.AutoEnable[2].ID != "plugin-demo-dynamic" {
		t.Fatalf("expected normalized auto-enable IDs, got %#v", cfg.AutoEnable)
	}
	for index, entry := range cfg.AutoEnable {
		if entry.WithMockData {
			t.Fatalf("expected bare-string YAML entries to default WithMockData=false at index %d, got %#v", index, entry)
		}
	}

	cfg.AutoEnable[0] = PluginAutoEnableEntry{ID: "mutated"}
	refreshed := svc.GetPlugin(context.Background())
	if refreshed.AutoEnable[0].ID != "demo-control" {
		t.Fatalf("expected cached auto-enable IDs to stay immutable, got %#v", refreshed.AutoEnable)
	}

	SetPluginAutoEnableOverride([]string{" override-plugin ", "override-plugin", "second-plugin"})
	overridden := svc.GetPlugin(context.Background())
	if len(overridden.AutoEnable) != 2 {
		t.Fatalf("expected override to replace auto-enable IDs, got %#v", overridden.AutoEnable)
	}
	if overridden.AutoEnable[0].ID != "override-plugin" || overridden.AutoEnable[1].ID != "second-plugin" {
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

// TestGetPluginAutoEnableEntriesParsesObjectForm verifies the structured-only
// schema: every entry must be a {id, withMockData} object. Missing
// withMockData defaults to false; explicit true opts into mock-data load.
func TestGetPluginAutoEnableEntriesParsesObjectForm(t *testing.T) {
	setTestConfigContent(t, `
plugin:
  autoEnable:
    - id: "demo-control"
    - id: "plugin-demo-source"
      withMockData: true
    - id: "plugin-demo-dynamic"
`)
	SetPluginAutoEnableOverride(nil)
	t.Cleanup(func() {
		SetPluginAutoEnableOverride(nil)
	})

	svc := New()
	entries := svc.GetPluginAutoEnableEntries(context.Background())
	if len(entries) != 3 {
		t.Fatalf("expected three normalized entries, got %#v", entries)
	}
	expected := []PluginAutoEnableEntry{
		{ID: "demo-control", WithMockData: false},
		{ID: "plugin-demo-source", WithMockData: true},
		{ID: "plugin-demo-dynamic", WithMockData: false},
	}
	for index, want := range expected {
		got := entries[index]
		if got.ID != want.ID || got.WithMockData != want.WithMockData {
			t.Fatalf("entry %d: expected %#v, got %#v", index, want, got)
		}
	}

	// IDs-only accessor should drop the WithMockData flag for callers that
	// only care about the ID list (e.g., the management UI's autoEnable badge).
	ids := svc.GetPluginAutoEnable(context.Background())
	if len(ids) != 3 || ids[1] != "plugin-demo-source" {
		t.Fatalf("expected IDs accessor to retain order, got %#v", ids)
	}
}

// TestGetPluginAutoEnableEntriesRejectsInvalidObjectForms verifies the schema
// validator panics with a clear message for each malformed shape so startup
// fails fast.
func TestGetPluginAutoEnableEntriesRejectsInvalidObjectForms(t *testing.T) {
	cases := []struct {
		name    string
		yaml    string
		wantSub string
	}{
		{
			name: "empty id field",
			yaml: `
plugin:
  autoEnable:
    - id: ""
      withMockData: true
`,
			wantSub: "field id cannot be empty",
		},
		{
			name: "missing id field",
			yaml: `
plugin:
  autoEnable:
    - withMockData: true
`,
			wantSub: "field id cannot be empty",
		},
		{
			name: "wrong type for withMockData",
			yaml: `
plugin:
  autoEnable:
    - id: "plugin-x"
      withMockData: "yes"
`,
			wantSub: "field withMockData must be a boolean",
		},
		{
			name: "unsupported field",
			yaml: `
plugin:
  autoEnable:
    - id: "plugin-x"
      enable: true
`,
			wantSub: "unsupported field",
		},
		{
			name: "non-array shape",
			yaml: `
plugin:
  autoEnable: "demo-control"
`,
			wantSub: "must be an array",
		},
		{
			name: "bare string entry rejected",
			yaml: `
plugin:
  autoEnable:
    - "demo-control"
`,
			wantSub: "must be a {id, withMockData} object",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setTestConfigContent(t, tc.yaml)
			SetPluginAutoEnableOverride(nil)
			t.Cleanup(func() {
				SetPluginAutoEnableOverride(nil)
			})

			defer func() {
				rec := recover()
				if rec == nil {
					t.Fatalf("expected panic with substring %q, got nil", tc.wantSub)
				}
				msg := fmt.Sprintf("%v", rec)
				if !contains(msg, tc.wantSub) {
					t.Fatalf("expected panic message to contain %q, got %q", tc.wantSub, msg)
				}
			}()
			svc := New()
			_ = svc.GetPluginAutoEnableEntries(context.Background())
		})
	}
}

// contains reports whether substring s appears anywhere in str.
func contains(str, s string) bool {
	if s == "" {
		return true
	}
	for i := 0; i+len(s) <= len(str); i++ {
		if str[i:i+len(s)] == s {
			return true
		}
	}
	return false
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
