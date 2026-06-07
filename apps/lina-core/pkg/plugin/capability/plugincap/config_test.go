// This file verifies the plugin-scoped read-only configuration service.

package plugincap

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

// TestScopedConfigReadsDevelopmentPluginConfig verifies development config is
// loaded from the plugin-owned manifest/config/config.yaml file.
func TestScopedConfigReadsDevelopmentPluginConfig(t *testing.T) {
	repoRoot := t.TempDir()
	writePluginConfig(t, repoRoot, "plugin-a", "storage:\n  endpoint: dev\n")

	svc := NewConfigFactory("", repoRoot).ForPlugin("plugin-a")
	value, err := svc.String(context.Background(), "storage.endpoint", "")
	if err != nil {
		t.Fatalf("read development config: %v", err)
	}
	if value != "dev" {
		t.Fatalf("expected development value dev, got %q", value)
	}
}

// TestScopedConfigProductionOverridesDevelopment verifies external production
// config is preferred over source-tree development config.
func TestScopedConfigProductionOverridesDevelopment(t *testing.T) {
	repoRoot := t.TempDir()
	productionRoot := t.TempDir()
	writePluginConfig(t, repoRoot, "plugin-a", "storage:\n  endpoint: dev\n")
	writeProductionPluginConfig(t, productionRoot, "plugin-a", "storage:\n  endpoint: prod\n")

	svc := NewConfigFactory(productionRoot, repoRoot).ForPlugin("plugin-a")
	value, err := svc.String(context.Background(), "storage.endpoint", "")
	if err != nil {
		t.Fatalf("read production config: %v", err)
	}
	if value != "prod" {
		t.Fatalf("expected production value prod, got %q", value)
	}
}

// TestScopedConfigReadsArtifactDefaultAfterFiles verifies artifact config is
// only used after production and development config files are absent.
func TestScopedConfigReadsArtifactDefaultAfterFiles(t *testing.T) {
	factory := NewConfigFactory(t.TempDir(), t.TempDir()).
		WithArtifactConfig("plugin-a", []byte("storage:\n  endpoint: artifact\n"))

	svc := factory.ForPlugin("plugin-a")
	value, err := svc.String(context.Background(), "storage.endpoint", "")
	if err != nil {
		t.Fatalf("read artifact config: %v", err)
	}
	if value != "artifact" {
		t.Fatalf("expected artifact value, got %q", value)
	}
}

// TestScopedConfigIgnoresTemplateConfig verifies config.example.yaml is never
// loaded as runtime defaults.
func TestScopedConfigIgnoresTemplateConfig(t *testing.T) {
	repoRoot := t.TempDir()
	templatePath := filepath.Join(repoRoot, "apps", "lina-plugins", "plugin-a", "manifest", "config", TemplateConfigFileName)
	writeFile(t, templatePath, "storage:\n  endpoint: template\n")

	svc := NewConfigFactory("", repoRoot).ForPlugin("plugin-a")
	value, err := svc.String(context.Background(), "storage.endpoint", "fallback")
	if err != nil {
		t.Fatalf("read missing runtime config: %v", err)
	}
	if value != "fallback" {
		t.Fatalf("expected template to be ignored, got %q", value)
	}
}

// TestScopedConfigDoesNotReadHostConfig verifies plugin keys do not fall back
// to the host global GoFrame configuration tree.
func TestScopedConfigDoesNotReadHostConfig(t *testing.T) {
	svc := NewConfigFactory(t.TempDir(), t.TempDir()).ForPlugin("plugin-a")
	value, err := svc.String(context.Background(), "database.default.link", "not-found")
	if err != nil {
		t.Fatalf("read missing plugin config: %v", err)
	}
	if value != "not-found" {
		t.Fatalf("expected host config to stay isolated, got %q", value)
	}
}

// TestScopedConfigRejectsRootLookup verifies callers cannot request a full
// plugin config snapshot through a blank or root key.
func TestScopedConfigRejectsRootLookup(t *testing.T) {
	svc := NewConfigFactory(t.TempDir(), t.TempDir()).ForPlugin("plugin-a")
	for _, key := range []string{"", " ", "."} {
		if _, err := svc.Get(context.Background(), key); err == nil {
			t.Fatalf("expected root lookup %q to fail", key)
		}
	}
}

// TestScopedConfigTypedHelpers verifies typed helper behavior remains scoped.
func TestScopedConfigTypedHelpers(t *testing.T) {
	repoRoot := t.TempDir()
	writePluginConfig(t, repoRoot, "plugin-a", `
custom:
  name: demo
  enabled: false
  count: 0
duration:
  interval: 45s
  blank: ""
`)

	svc := NewConfigFactory("", repoRoot).ForPlugin("plugin-a")
	ctx := context.Background()

	target := &scanTarget{}
	if err := svc.Scan(ctx, "custom", target); err != nil {
		t.Fatalf("scan config section: %v", err)
	}
	if target.Name != "demo" || target.Enabled || target.Count != 0 {
		t.Fatalf("unexpected scan target: %#v", target)
	}

	interval, err := svc.Duration(ctx, "duration.interval", time.Minute)
	if err != nil {
		t.Fatalf("read duration: %v", err)
	}
	if interval != 45*time.Second {
		t.Fatalf("expected 45s duration, got %s", interval)
	}

	blank, err := svc.Duration(ctx, "duration.blank", time.Minute)
	if err != nil {
		t.Fatalf("read blank duration: %v", err)
	}
	if blank != time.Minute {
		t.Fatalf("expected blank duration default, got %s", blank)
	}
}

// TestScopedConfigDurationReturnsErrorForInvalidValue verifies invalid
// duration strings still report the key.
func TestScopedConfigDurationReturnsErrorForInvalidValue(t *testing.T) {
	repoRoot := t.TempDir()
	writePluginConfig(t, repoRoot, "plugin-a", "duration:\n  interval: invalid\n")

	_, err := NewConfigFactory("", repoRoot).ForPlugin("plugin-a").Duration(context.Background(), "duration.interval", time.Minute)
	if err == nil {
		t.Fatal("expected invalid duration error")
	}
	if !strings.Contains(err.Error(), "duration.interval") {
		t.Fatalf("expected error to mention key, got %v", err)
	}
}

// writePluginConfig writes a development plugin config file.
func writePluginConfig(t *testing.T, repoRoot string, pluginID string, content string) {
	t.Helper()
	writeFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", pluginID, "manifest", "config", RuntimeConfigFileName),
		content,
	)
}

// writeProductionPluginConfig writes a production plugin config file.
func writeProductionPluginConfig(t *testing.T, productionRoot string, pluginID string, content string) {
	t.Helper()
	writeFile(t, filepath.Join(productionRoot, "plugins", pluginID, RuntimeConfigFileName), content)
}

// writeFile writes one fixture file for a test.
func writeFile(t *testing.T, filePath string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("create fixture dir: %v", err)
	}
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture file: %v", err)
	}
}
