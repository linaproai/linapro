// This file verifies plugin-scoped manifest resource reads.

package manifestcap

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// testMetadata captures manifest metadata fixture values.
type testMetadata struct {
	// Name is the fixture metadata name.
	Name string `yaml:"name"`
	// Enabled is the fixture metadata switch.
	Enabled bool `yaml:"enabled"`
}

// TestManifestReadsDevelopmentResources verifies a plugin can read its own
// raw manifest resources from the development source tree.
func TestManifestReadsDevelopmentResources(t *testing.T) {
	repoRoot := t.TempDir()
	fixtures := map[string]string{
		"metadata.yaml":              "name: alpha\nenabled: true\n",
		"config/config.example.yaml": "feature:\n  enabled: false\n",
		"sql/001-schema.sql":         "CREATE TABLE plugin_demo(id bigint);\n",
		"i18n/zh-CN/plugin.json":     `{"plugin.demo":"demo"}`,
	}
	for path, content := range fixtures {
		writeManifestFile(t, repoRoot, "plugin-a", path, content)
	}

	svc := NewFactory(repoRoot).ForPlugin("plugin-a")
	for path, expected := range fixtures {
		content, err := svc.Get(context.Background(), path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if string(content) != expected {
			t.Fatalf("expected %s content %q, got %q", path, expected, string(content))
		}
	}
}

// TestManifestScansYAML verifies YAML documents scan into caller-owned structs.
func TestManifestScansYAML(t *testing.T) {
	repoRoot := t.TempDir()
	writeManifestFile(t, repoRoot, "plugin-a", "metadata.yaml", "name: alpha\nenabled: true\n")

	target := &testMetadata{}
	err := NewFactory(repoRoot).ForPlugin("plugin-a").Scan(context.Background(), "metadata.yaml", "", target)
	if err != nil {
		t.Fatalf("scan metadata: %v", err)
	}
	if target.Name != "alpha" || !target.Enabled {
		t.Fatalf("unexpected target: %#v", target)
	}
}

// TestManifestRejectsUnsafePaths verifies path governance rejects unsafe inputs.
func TestManifestRejectsUnsafePaths(t *testing.T) {
	svc := NewFactory(t.TempDir()).ForPlugin("plugin-a")
	for _, path := range []string{
		"",
		".",
		"../plugin-b/manifest/metadata.yaml",
		"/etc/passwd",
		"C:\\secret.yaml",
		"http://example.com/config.yaml",
		"manifest/metadata.yaml",
	} {
		if _, err := svc.Get(context.Background(), path); err == nil {
			t.Fatalf("expected path %q to be rejected", path)
		}
	}
}

// TestManifestDoesNotCrossPluginScope verifies another plugin's manifest file
// cannot be reached through relative path traversal.
func TestManifestDoesNotCrossPluginScope(t *testing.T) {
	repoRoot := t.TempDir()
	writeManifestFile(t, repoRoot, "plugin-b", "metadata.yaml", "name: beta\n")

	svc := NewFactory(repoRoot).ForPlugin("plugin-a")
	if _, err := svc.Get(context.Background(), "../plugin-b/manifest/metadata.yaml"); err == nil {
		t.Fatal("expected cross-plugin traversal to fail")
	}
	exists, err := svc.Exists(context.Background(), "metadata.yaml")
	if err != nil {
		t.Fatalf("check missing own metadata: %v", err)
	}
	if exists {
		t.Fatal("expected plugin-a metadata to be absent")
	}
}

// TestManifestMissingResourceReturnsNil verifies missing resources are reported
// without errors or synthetic placeholder metadata files.
func TestManifestMissingResourceReturnsNil(t *testing.T) {
	svc := NewFactory(t.TempDir()).ForPlugin("plugin-a")

	content, err := svc.Get(context.Background(), "metadata.yaml")
	if err != nil {
		t.Fatalf("read missing metadata: %v", err)
	}
	if len(content) != 0 {
		t.Fatalf("expected missing metadata to return empty content, got %q", string(content))
	}
}

// writeManifestFile writes one plugin manifest fixture file.
func writeManifestFile(t *testing.T, repoRoot string, pluginID string, resourcePath string, content string) {
	t.Helper()
	filePath := filepath.Join(repoRoot, "apps", "lina-plugins", pluginID, "manifest", filepath.FromSlash(resourcePath))
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("create manifest fixture dir: %v", err)
	}
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write manifest fixture: %v", err)
	}
}
