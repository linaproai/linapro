// This file verifies dynamic plugin build workspace helpers.

package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEnsureBuildWasmPluginWorkspaceReusesOfficialPluginRootModule verifies
// test helpers keep the official plugin root module as the lina-plugins bridge.
func TestEnsureBuildWasmPluginWorkspaceReusesOfficialPluginRootModule(t *testing.T) {
	repoRoot := t.TempDir()
	writeBuildWorkspaceTestFile(t, filepath.Join(repoRoot, "go.work"), `go 1.25.0

use (
	./apps/lina-core
)
`)
	pluginRoot := filepath.Join(repoRoot, "apps", "lina-plugins")
	writeBuildWorkspaceTestFile(t, filepath.Join(pluginRoot, "go.mod"), "module lina-plugins\n")
	writeBuildWorkspaceTestFile(t, filepath.Join(pluginRoot, "plugin-demo-dynamic", "go.mod"), "module plugin-demo-dynamic\n")
	writeBuildWorkspaceTestFile(t, filepath.Join(pluginRoot, "plugin-demo-dynamic", "plugin.yaml"), "id: plugin-demo-dynamic\n")

	if err := ensureBuildWasmPluginWorkspace(repoRoot, filepath.Join(pluginRoot, "plugin-demo-dynamic")); err != nil {
		t.Fatalf("ensureBuildWasmPluginWorkspace returned error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(repoRoot, "temp", "go.work.plugins"))
	if err != nil {
		t.Fatalf("read generated plugin workspace: %v", err)
	}
	text := string(content)
	if strings.Contains(text, "./official-plugins") {
		t.Fatalf("expected official plugin root module to replace fallback aggregate, got:\n%s", text)
	}
	if !strings.Contains(text, "../apps/lina-plugins\n") {
		t.Fatalf("expected generated workspace to include official plugin root module, got:\n%s", text)
	}
	if !strings.Contains(text, "../apps/lina-plugins/plugin-demo-dynamic\n") {
		t.Fatalf("expected generated workspace to include dynamic plugin module, got:\n%s", text)
	}
}

// writeBuildWorkspaceTestFile writes one fixture file for workspace helper tests.
func writeBuildWorkspaceTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
