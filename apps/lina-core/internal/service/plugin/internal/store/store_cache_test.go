// This file verifies store-local release manifest and YAML snapshot caches keep
// parse costs bounded while returning detached projections to callers.

package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
)

// TestLoadReleaseManifestUsesReleaseCache verifies repeated release manifest
// reads reuse the store cache and return detached manifests.
func TestLoadReleaseManifestUsesReleaseCache(t *testing.T) {
	catalogSvc := &storeCacheCatalogStub{storageDir: t.TempDir()}
	writeStoreCacheArtifact(t, catalogSvc.storageDir, "store-cache-plugin.wasm")
	service := New(catalogSvc, nil).(*serviceImpl)
	release := &ReleaseRecord{
		Id:             101,
		PluginId:       "store-cache-plugin",
		ReleaseVersion: "0.1.0",
		PackagePath:    "store-cache-plugin.wasm",
		Checksum:       "checksum-one",
	}

	first, err := service.LoadReleaseManifest(context.Background(), release)
	if err != nil {
		t.Fatalf("first release manifest load failed: %v", err)
	}
	first.Name = "mutated"
	second, err := service.LoadReleaseManifest(context.Background(), release)
	if err != nil {
		t.Fatalf("second release manifest load failed: %v", err)
	}
	if catalogSvc.loadCount != 1 {
		t.Fatalf("expected release manifest to load once, got %d", catalogSvc.loadCount)
	}
	if second.Name == "mutated" {
		t.Fatal("expected cached release manifest to be detached from caller mutation")
	}
}

// TestLoadReleaseManifestCacheRejectsMissingArchive verifies active release
// manifest cache entries are not reused after the archive artifact disappears.
func TestLoadReleaseManifestCacheRejectsMissingArchive(t *testing.T) {
	catalogSvc := &storeCacheCatalogStub{storageDir: t.TempDir()}
	artifactPath := writeStoreCacheArtifact(t, catalogSvc.storageDir, "store-cache-plugin.wasm")
	service := New(catalogSvc, nil).(*serviceImpl)
	release := &ReleaseRecord{
		Id:             102,
		PluginId:       "store-cache-plugin",
		ReleaseVersion: "0.1.0",
		PackagePath:    "store-cache-plugin.wasm",
		Checksum:       "checksum-one",
	}

	if _, err := service.LoadReleaseManifest(context.Background(), release); err != nil {
		t.Fatalf("initial release manifest load failed: %v", err)
	}
	if err := os.Remove(artifactPath); err != nil {
		t.Fatalf("remove release artifact: %v", err)
	}
	if _, err := service.LoadReleaseManifest(context.Background(), release); err == nil {
		t.Fatal("expected missing release artifact to bypass cache and return error")
	}
}

// TestParseManifestSnapshotUsesContentCache verifies the YAML snapshot parser
// returns detached cached snapshots keyed by content.
func TestParseManifestSnapshotUsesContentCache(t *testing.T) {
	service := New(&storeCacheCatalogStub{storageDir: t.TempDir()}, nil).(*serviceImpl)
	content := `
id: store-cache-snapshot
name: Store Cache Snapshot
version: 0.1.0
type: dynamic
manifestDeclared: true
dependencies:
  plugins:
    - id: store-cache-dependency
      version: ">=0.1.0"
requestedHostServices:
  - service: runtime
    methods:
      - log.write
`
	first, err := service.ParseManifestSnapshot(content)
	if err != nil {
		t.Fatalf("first snapshot parse failed: %v", err)
	}
	first.Name = "mutated"
	first.Dependencies.Plugins[0].ID = "mutated"

	second, err := service.ParseManifestSnapshot(content)
	if err != nil {
		t.Fatalf("second snapshot parse failed: %v", err)
	}
	if second.Name == "mutated" {
		t.Fatal("expected cached snapshot to be detached from caller mutation")
	}
	if second.Dependencies.Plugins[0].ID == "mutated" {
		t.Fatal("expected cached dependency snapshot to be detached from caller mutation")
	}
	if len(service.manifestSnapshotCache) != 1 {
		t.Fatalf("expected one cached snapshot, got %d", len(service.manifestSnapshotCache))
	}
}

// writeStoreCacheArtifact creates a real archive path for file-identity guards.
func writeStoreCacheArtifact(t *testing.T, storageDir string, name string) string {
	t.Helper()
	artifactPath := filepath.Join(storageDir, name)
	if err := os.WriteFile(artifactPath, []byte("store cache artifact"), 0o600); err != nil {
		t.Fatalf("write store cache artifact: %v", err)
	}
	return artifactPath
}

// storeCacheCatalogStub provides the catalog methods required by store tests.
type storeCacheCatalogStub struct {
	storageDir string
	loadCount  int
}

// BuildRegistryChecksum returns a stable checksum for test manifests.
func (s *storeCacheCatalogStub) BuildRegistryChecksum(manifest *catalog.Manifest) string {
	if manifest == nil {
		return ""
	}
	return manifest.ID + ":" + manifest.Version
}

// BuildPackagePath returns the runtime artifact path for test manifests.
func (s *storeCacheCatalogStub) BuildPackagePath(manifest *catalog.Manifest) string {
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return ""
	}
	return manifest.RuntimeArtifact.Path
}

// RuntimeStorageDir returns the isolated release storage path.
func (s *storeCacheCatalogStub) RuntimeStorageDir(context.Context) (string, error) {
	return s.storageDir, nil
}

// LoadManifestFromArtifactPath returns a fresh manifest and records load count.
func (s *storeCacheCatalogStub) LoadManifestFromArtifactPath(artifactPath string) (*catalog.Manifest, error) {
	if _, err := os.Stat(artifactPath); err != nil {
		return nil, err
	}
	s.loadCount++
	return &catalog.Manifest{
		ID:      "store-cache-plugin",
		Name:    "Store Cache Plugin",
		Version: "0.1.0",
		Type:    plugintypes.TypeDynamic.String(),
		RuntimeArtifact: &catalog.ArtifactSpec{
			Path: artifactPath,
		},
	}, nil
}

// ListInstallSQLPaths returns no install SQL for store cache tests.
func (s *storeCacheCatalogStub) ListInstallSQLPaths(*catalog.Manifest) []string {
	return []string{}
}

// ListUninstallSQLPaths returns no uninstall SQL for store cache tests.
func (s *storeCacheCatalogStub) ListUninstallSQLPaths(*catalog.Manifest) []string {
	return []string{}
}

// ListMockSQLPaths returns no mock SQL for store cache tests.
func (s *storeCacheCatalogStub) ListMockSQLPaths(*catalog.Manifest) []string {
	return []string{}
}

// HasMockSQLData reports no mock SQL for store cache tests.
func (s *storeCacheCatalogStub) HasMockSQLData(*catalog.Manifest) bool {
	return false
}

// DiscoverSQLPaths returns no SQL files for store cache tests.
func (s *storeCacheCatalogStub) DiscoverSQLPaths(string, bool) []string {
	return []string{}
}

// DiscoverMockSQLPaths returns no mock SQL files for store cache tests.
func (s *storeCacheCatalogStub) DiscoverMockSQLPaths(string) []string {
	return []string{}
}

// ListFrontendPagePaths returns no frontend pages for store cache tests.
func (s *storeCacheCatalogStub) ListFrontendPagePaths(*catalog.Manifest) []string {
	return []string{}
}

// ListFrontendSlotPaths returns no frontend slots for store cache tests.
func (s *storeCacheCatalogStub) ListFrontendSlotPaths(*catalog.Manifest) []string {
	return []string{}
}

// DiscoverPagePaths returns no frontend pages for store cache tests.
func (s *storeCacheCatalogStub) DiscoverPagePaths(string) []string {
	return []string{}
}

// DiscoverSlotPaths returns no frontend slots for store cache tests.
func (s *storeCacheCatalogStub) DiscoverSlotPaths(string) []string {
	return []string{}
}
