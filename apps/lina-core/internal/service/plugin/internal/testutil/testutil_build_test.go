// This file verifies dynamic plugin build workspace helpers.

package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEnsureBuildWasmPluginWorkspaceGeneratesAggregateModule verifies test
// helpers use the generated bridge module even if a stale root module exists.
func TestEnsureBuildWasmPluginWorkspaceGeneratesAggregateModule(t *testing.T) {
	repoRoot := t.TempDir()
	writeBuildWorkspaceTestFile(t, filepath.Join(repoRoot, "go.work"), `go 1.25.0

use (
	./apps/lina-core
)
`)
	pluginRoot := filepath.Join(repoRoot, "apps", "lina-plugins")
	writeBuildWorkspaceTestFile(t, filepath.Join(pluginRoot, "go.mod"), "module lina-plugins\n")
	writeBuildWorkspaceTestFile(t, filepath.Join(pluginRoot, "linapro-demo-dynamic", "go.mod"), "module linapro-demo-dynamic\n")
	writeBuildWorkspaceTestFile(t, filepath.Join(pluginRoot, "linapro-demo-dynamic", "plugin.yaml"), "id: linapro-demo-dynamic\n")

	if err := ensureBuildWasmPluginWorkspace(repoRoot, filepath.Join(pluginRoot, "linapro-demo-dynamic")); err != nil {
		t.Fatalf("ensureBuildWasmPluginWorkspace returned error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(repoRoot, "temp", "go.work.plugins"))
	if err != nil {
		t.Fatalf("read generated plugin workspace: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "./official-plugins") {
		t.Fatalf("expected generated workspace to include official plugin aggregate, got:\n%s", text)
	}
	if strings.Contains(text, "../apps/lina-plugins\n") {
		t.Fatalf("expected generated workspace to skip official plugin root module, got:\n%s", text)
	}
	if !strings.Contains(text, "../apps/lina-plugins/linapro-demo-dynamic\n") {
		t.Fatalf("expected generated workspace to include dynamic plugin module, got:\n%s", text)
	}
	if _, err = os.Stat(filepath.Join(repoRoot, "temp", "official-plugins", "go.mod")); err != nil {
		t.Fatalf("expected generated aggregate go.mod: %v", err)
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
