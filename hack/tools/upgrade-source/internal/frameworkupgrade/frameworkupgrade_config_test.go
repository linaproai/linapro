// This file verifies the development-only upgrade metadata stays aligned with
// the runtime metadata exposed by the host.

package frameworkupgrade

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"gopkg.in/yaml.v3"
)

// runtimeMetadataConfig stores only the runtime metadata fields mirrored by hack config.
type runtimeMetadataConfig struct {
	Framework struct {
		Version       string `yaml:"version"`
		RepositoryURL string `yaml:"repositoryUrl"`
	} `yaml:"framework"`
}

// TestHackConfigMatchesRuntimeMetadata verifies the mirrored upgrade metadata
// stays aligned with manifest/config/metadata.yaml in the current source tree.
func TestHackConfigMatchesRuntimeMetadata(t *testing.T) {
	t.Parallel()

	repoRoot := mustResolveRepoRoot(t)
	upgradeConfig, err := readCurrentUpgradeMetadata(repoRoot)
	if err != nil {
		t.Fatalf("read current upgrade metadata: %v", err)
	}

	runtimeConfigPath := filepath.Join(repoRoot, "apps", "lina-core", "manifest", "config", "metadata.yaml")
	content, err := os.ReadFile(runtimeConfigPath)
	if err != nil {
		t.Fatalf("read runtime metadata: %v", err)
	}

	cfg := &runtimeMetadataConfig{}
	if err = yaml.Unmarshal(content, cfg); err != nil {
		t.Fatalf("unmarshal runtime metadata: %v", err)
	}
	if upgradeConfig.Version != cfg.Framework.Version {
		t.Fatalf("expected hack config version %q to equal runtime metadata version %q", upgradeConfig.Version, cfg.Framework.Version)
	}
	if upgradeConfig.RepositoryURL != cfg.Framework.RepositoryURL {
		t.Fatalf("expected hack config repository url %q to equal runtime metadata repository url %q", upgradeConfig.RepositoryURL, cfg.Framework.RepositoryURL)
	}
}

// mustResolveRepoRoot resolves the repository root from the current test file location.
func mustResolveRepoRoot(t *testing.T) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current file path failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "..", ".."))
}
