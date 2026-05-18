// This file verifies repository-root anchoring for runtime filesystem paths.

package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestResolveRuntimePathFromWorkingDirAnchorsAtRepositoryRoot verifies relative
// runtime paths resolve from the LinaPro repository root instead of the caller
// process working directory.
func TestResolveRuntimePathFromWorkingDirAnchorsAtRepositoryRoot(t *testing.T) {
	repoRoot := newTestRepoRoot(t)
	backendWorkingDir := filepath.Join(repoRoot, "apps", "lina-core")

	resolvedPath := resolveRuntimePathFromWorkingDir("temp/output", backendWorkingDir)
	expectedPath := filepath.Join(repoRoot, "temp", "output")
	if resolvedPath != expectedPath {
		t.Fatalf("expected path %q, got %q", expectedPath, resolvedPath)
	}
}

// TestConfigRuntimePathGettersAnchorRelativePathsAtRepositoryRoot verifies upload
// and dynamic plugin storage paths share the same root anchoring behavior.
func TestConfigRuntimePathGettersAnchorRelativePathsAtRepositoryRoot(t *testing.T) {
	setTestConfigContent(t, `
upload:
  path: temp/upload
plugin:
  dynamic:
    storagePath: temp/output
`)
	repoRoot := newTestRepoRoot(t)
	backendWorkingDir := filepath.Join(repoRoot, "apps", "lina-core")
	expectedRepoRoot := realTestPath(t, repoRoot)
	withWorkingDir(t, backendWorkingDir)
	SetPluginDynamicStoragePathOverride("")
	t.Cleanup(func() {
		SetPluginDynamicStoragePathOverride("")
	})

	svc := New()
	if path := svc.GetUploadPath(context.Background()); path != filepath.Join(expectedRepoRoot, "temp", "upload") {
		t.Fatalf("expected upload path under repo temp, got %q", path)
	}
	if path := svc.GetPluginDynamicStoragePath(context.Background()); path != filepath.Join(expectedRepoRoot, "temp", "output") {
		t.Fatalf("expected plugin storage path under repo temp, got %q", path)
	}
}

// newTestRepoRoot creates a minimal LinaPro-like repository layout for path
// resolution tests without depending on the real checkout location.
func newTestRepoRoot(t *testing.T) string {
	t.Helper()

	repoRoot := t.TempDir()
	for _, dir := range []string{
		filepath.Join(repoRoot, "apps", "lina-core"),
		filepath.Join(repoRoot, "apps", "lina-vben"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create test repo dir %s: %v", dir, err)
		}
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "go.work"), []byte("go 1.26\n"), 0o644); err != nil {
		t.Fatalf("write test go.work: %v", err)
	}
	return repoRoot
}

// realTestPath resolves symlinks in an existing test path before string
// comparisons on platforms where os.Getwd canonicalizes temporary directories.
func realTestPath(t *testing.T, targetPath string) string {
	t.Helper()

	realPath, err := filepath.EvalSymlinks(targetPath)
	if err != nil {
		t.Fatalf("resolve real test path %s: %v", targetPath, err)
	}
	return realPath
}

// withWorkingDir changes the process working directory for one test and
// restores it during cleanup.
func withWorkingDir(t *testing.T, workingDir string) {
	t.Helper()

	originalWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get original working directory: %v", err)
	}
	if err = os.Chdir(workingDir); err != nil {
		t.Fatalf("change working directory to %s: %v", workingDir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalWorkingDir); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})
}
