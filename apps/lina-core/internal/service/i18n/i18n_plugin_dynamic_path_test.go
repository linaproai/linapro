// This file verifies relative dynamic-plugin storage paths resolve from the
// repository root so runtime i18n can read enabled plugin artifacts.

package i18n

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	configsvc "lina-core/internal/service/config"
)

// TestResolveDynamicPluginPackagePathAnchorsRelativeStoragePathAtRepoRoot verifies
// that runtime i18n resolves temp/output against the repository root instead of
// the apps/lina-core working directory.
func TestResolveDynamicPluginPackagePathAnchorsRelativeStoragePathAtRepoRoot(t *testing.T) {
	t.Helper()

	originalWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	repoRoot := t.TempDir()
	if err = os.WriteFile(filepath.Join(repoRoot, "go.work"), []byte("go 1.26\n"), 0o644); err != nil {
		t.Fatalf("write go.work: %v", err)
	}

	storageRoot := filepath.Join(repoRoot, "temp", "output")
	packagePath := filepath.Join("releases", "plugin-demo-dynamic", "v0.1.0", "plugin-demo-dynamic.wasm")
	expectedPath := filepath.Join(storageRoot, packagePath)
	if err = os.MkdirAll(filepath.Dir(expectedPath), 0o755); err != nil {
		t.Fatalf("create runtime storage dir: %v", err)
	}
	if err = os.WriteFile(expectedPath, []byte("wasm"), 0o644); err != nil {
		t.Fatalf("write runtime artifact: %v", err)
	}

	workingDir := filepath.Join(repoRoot, "apps", "lina-core")
	if err = os.MkdirAll(workingDir, 0o755); err != nil {
		t.Fatalf("create working directory: %v", err)
	}
	if err = os.Chdir(workingDir); err != nil {
		t.Fatalf("chdir to fake apps/lina-core: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWorkingDir)
		configsvc.SetPluginDynamicStoragePathOverride("")
	})

	configsvc.SetPluginDynamicStoragePathOverride("temp/output")

	svc := New().(*serviceImpl)

	resolvedPath, err := svc.resolveDynamicPluginPackagePath(context.Background(), filepath.ToSlash(packagePath))
	if err != nil {
		t.Fatalf("resolve dynamic plugin package path: %v", err)
	}
	expectedRealPath, err := filepath.EvalSymlinks(expectedPath)
	if err != nil {
		t.Fatalf("eval expected path symlink: %v", err)
	}
	resolvedRealPath, err := filepath.EvalSymlinks(resolvedPath)
	if err != nil {
		t.Fatalf("eval resolved path symlink: %v", err)
	}
	if resolvedRealPath != expectedRealPath {
		t.Fatalf("expected resolved path %q, got %q", expectedRealPath, resolvedRealPath)
	}
}
