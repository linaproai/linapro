// Package testsupport provides test-only helpers shared by LinaCore test suites.
package testsupport

import (
	"os"
	"path/filepath"
)

const officialPluginsRelativeDir = "apps/lina-plugins"

// OfficialPluginsRoot returns the default official-plugin workspace path.
func OfficialPluginsRoot(repoRoot string) string {
	return filepath.Join(repoRoot, officialPluginsRelativeDir)
}

// OfficialPluginsWorkspaceReady reports whether the official plugin workspace
// has been initialized with at least one source plugin manifest.
func OfficialPluginsWorkspaceReady(repoRoot string) bool {
	rootDir := OfficialPluginsRoot(repoRoot)
	info, err := os.Stat(rootDir)
	if err != nil || !info.IsDir() {
		return false
	}

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, err = os.Stat(filepath.Join(rootDir, entry.Name(), "plugin.yaml")); err == nil {
			return true
		}
	}
	return false
}
