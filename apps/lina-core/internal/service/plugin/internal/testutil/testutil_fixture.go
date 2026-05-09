// This file creates source and repository-backed plugin fixture directories.

package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/pkg/pluginbridge"
)

// CreateTestPluginDir creates a source plugin directory with the default file layout.
func CreateTestPluginDir(t *testing.T, pluginID string) string {
	t.Helper()

	pluginDir := filepath.Join(testSourcePluginRootDir, pluginID)
	catalog.SetPluginRootDirOverride(testSourcePluginRootDir)
	if err := os.MkdirAll(filepath.Join(pluginDir, "backend"), 0o755); err != nil {
		t.Fatalf("failed to create backend dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(pluginDir, "frontend", "pages"), 0o755); err != nil {
		t.Fatalf("failed to create frontend pages dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(pluginDir, "manifest", "sql", "uninstall"), 0o755); err != nil {
		t.Fatalf("failed to create sql dir: %v", err)
	}

	t.Cleanup(func() {
		catalog.SetPluginRootDirOverride("")
		if cleanupErr := os.RemoveAll(pluginDir); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove test plugin dir %s: %v", pluginDir, cleanupErr)
		}
	})

	WriteTestFile(t, filepath.Join(pluginDir, "go.mod"), "module "+strings.ReplaceAll(pluginID, "-", "_")+"\n\ngo 1.25.0\n")
	WriteTestFile(t, filepath.Join(pluginDir, "backend", "plugin.go"), "package backend\n")
	WriteTestFile(t, filepath.Join(pluginDir, "frontend", "pages", "main-entry.vue"), "<template><div /></template>\n")
	WriteTestFile(t, filepath.Join(pluginDir, "manifest", "sql", "001-"+pluginID+".sql"), "SELECT 1;\n")
	WriteTestFile(t, filepath.Join(pluginDir, "manifest", "sql", "uninstall", "001-"+pluginID+".sql"), "SELECT 1;\n")
	WriteTestFile(t, filepath.Join(pluginDir, "plugin.yaml"), "id: "+pluginID+"\nname: test\nversion: 0.1.0\ntype: source\n")

	return pluginDir
}

// CreateTestRuntimePluginDir creates a runtime plugin source directory with a default frontend bundle.
func CreateTestRuntimePluginDir(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	installSQLAssets []*catalog.ArtifactSQLAsset,
	uninstallSQLAssets []*catalog.ArtifactSQLAsset,
) string {
	return CreateTestRuntimePluginDirWithFrontendAssets(
		t,
		pluginID,
		pluginName,
		version,
		DefaultTestRuntimeFrontendAssets(),
		installSQLAssets,
		uninstallSQLAssets,
	)
}

// CreateTestRuntimePluginDirWithFrontendAssets creates a runtime plugin source directory with one embedded artifact.
func CreateTestRuntimePluginDirWithFrontendAssets(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	frontendAssets []*catalog.ArtifactFrontendAsset,
	installSQLAssets []*catalog.ArtifactSQLAsset,
	uninstallSQLAssets []*catalog.ArtifactSQLAsset,
) string {
	t.Helper()

	repoRoot, err := FindRepoRoot(".")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}

	pluginDir := filepath.Join(repoRoot, "apps", "lina-plugins", pluginID)
	if err = os.MkdirAll(filepath.Join(pluginDir, "runtime"), 0o755); err != nil {
		t.Fatalf("failed to create runtime dir: %v", err)
	}

	t.Cleanup(func() {
		if cleanupErr := os.RemoveAll(pluginDir); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove runtime test plugin dir %s: %v", pluginDir, cleanupErr)
		}
	})

	WriteTestFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: "+pluginID+"\nname: "+pluginName+"\nversion: "+version+"\ntype: dynamic\n",
	)
	WriteRuntimeWasmArtifact(
		t,
		filepath.Join(pluginDir, runtime.BuildArtifactRelativePath(pluginID)),
		&catalog.ArtifactManifest{
			ID:      pluginID,
			Name:    pluginName,
			Version: version,
			Type:    catalog.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:        pluginbridge.RuntimeKindWasm,
			ABIVersion:         pluginbridge.SupportedABIVersion,
			FrontendAssetCount: len(frontendAssets),
			SQLAssetCount:      len(installSQLAssets) + len(uninstallSQLAssets),
		},
		frontendAssets,
		installSQLAssets,
		uninstallSQLAssets,
		nil,
		nil,
		nil,
	)
	return pluginDir
}

// WriteTestFile writes one UTF-8 fixture file to disk for the current test.
func WriteTestFile(t *testing.T, filePath string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("failed to create test file dir %s: %v", filePath, err)
	}
	tempFile, err := os.CreateTemp(filepath.Dir(filePath), filepath.Base(filePath)+".tmp-*")
	if err != nil {
		t.Fatalf("failed to create temp test file %s: %v", filePath, err)
	}
	tempPath := tempFile.Name()
	defer func() {
		if cleanupErr := os.Remove(tempPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove temp test file %s: %v", tempPath, cleanupErr)
		}
	}()

	if _, err = tempFile.Write([]byte(content)); err != nil {
		if closeErr := tempFile.Close(); closeErr != nil {
			t.Fatalf("failed to close temp test file %s after write error: %v", filePath, closeErr)
		}
		t.Fatalf("failed to write temp test file %s: %v", filePath, err)
	}
	if err = tempFile.Chmod(0o644); err != nil {
		if closeErr := tempFile.Close(); closeErr != nil {
			t.Fatalf("failed to close temp test file %s after chmod error: %v", filePath, closeErr)
		}
		t.Fatalf("failed to chmod temp test file %s: %v", filePath, err)
	}
	if err = tempFile.Close(); err != nil {
		t.Fatalf("failed to close temp test file %s: %v", filePath, err)
	}
	if err = os.Rename(tempPath, filePath); err != nil {
		t.Fatalf("failed to move test file into place %s: %v", filePath, err)
	}
}
