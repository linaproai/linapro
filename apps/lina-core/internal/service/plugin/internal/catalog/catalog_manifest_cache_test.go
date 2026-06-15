// This file verifies catalog manifest read-model caching for dynamic artifacts
// and source manifest clones.

package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/resourcefs"
	"lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-core/pkg/plugin/pluginhost"
)

// TestRuntimeManifestCacheReusesStatGuardForThirtyArtifacts verifies repeated
// scans only stat artifact files and reuse parsed manifests when size and mtime
// are unchanged.
func TestRuntimeManifestCacheReusesStatGuardForThirtyArtifacts(t *testing.T) {
	service := newCacheTestCatalogService(t)
	artifactPaths := createCacheTestRuntimeArtifacts(t, service.runtimeStorageDir, 30, "0.1.0")

	if _, err := service.ScanManifests(); err != nil {
		t.Fatalf("initial scan failed: %v", err)
	}
	for _, artifactPath := range artifactPaths {
		if got := service.runtimeArtifactParseCount(artifactPath); got != 1 {
			t.Fatalf("expected first scan to parse %s once, got %d", artifactPath, got)
		}
	}

	if _, err := service.ScanManifests(); err != nil {
		t.Fatalf("cached scan failed: %v", err)
	}
	for _, artifactPath := range artifactPaths {
		if got := service.runtimeArtifactParseCount(artifactPath); got != 1 {
			t.Fatalf("expected cached scan to avoid reparsing %s, got %d", artifactPath, got)
		}
	}
}

// TestGetDesiredManifestUsesRuntimeIndex verifies a known dynamic plugin ID can
// be loaded without reparsing unrelated artifacts.
func TestGetDesiredManifestUsesRuntimeIndex(t *testing.T) {
	service := newCacheTestCatalogService(t)
	artifactPaths := createCacheTestRuntimeArtifacts(t, service.runtimeStorageDir, 30, "0.1.0")
	if _, err := service.ScanManifests(); err != nil {
		t.Fatalf("initial scan failed: %v", err)
	}

	targetID := "cache-test-plugin-017"
	manifest, err := service.GetDesiredManifest(targetID)
	if err != nil {
		t.Fatalf("get desired manifest failed: %v", err)
	}
	if manifest == nil || manifest.ID != targetID {
		t.Fatalf("expected target manifest %s, got %#v", targetID, manifest)
	}
	for _, artifactPath := range artifactPaths {
		if got := service.runtimeArtifactParseCount(artifactPath); got != 1 {
			t.Fatalf("expected indexed desired lookup to avoid reparsing %s, got %d", artifactPath, got)
		}
	}
}

// TestRuntimeManifestCacheDetectsFileReplacement verifies file stat changes
// force a fresh parse even when the artifact path stays the same.
func TestRuntimeManifestCacheDetectsFileReplacement(t *testing.T) {
	service := newCacheTestCatalogService(t)
	pluginID := "cache-test-replace"
	artifactPath := writeCacheTestRuntimeArtifact(t, service.runtimeStorageDir, pluginID, "0.1.0")
	manifest, err := service.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("initial load failed: %v", err)
	}
	if manifest.Version != "0.1.0" {
		t.Fatalf("expected initial version 0.1.0, got %s", manifest.Version)
	}

	time.Sleep(time.Millisecond)
	writeCacheTestRuntimeArtifactAtPath(t, artifactPath, pluginID, "0.2.0")
	reloaded, err := service.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("reload after replace failed: %v", err)
	}
	if reloaded.Version != "0.2.0" {
		t.Fatalf("expected replaced version 0.2.0, got %s", reloaded.Version)
	}
	if got := service.runtimeArtifactParseCount(artifactPath); got != 2 {
		t.Fatalf("expected replacement to parse twice, got %d", got)
	}
}

// TestRuntimeManifestCacheDetectsAtomicReplacementWithSameStat verifies
// same-path package replacement cannot reuse stale manifests when coarse
// filesystems report the same size and mtime.
func TestRuntimeManifestCacheDetectsAtomicReplacementWithSameStat(t *testing.T) {
	service := newCacheTestCatalogService(t)
	pluginID := "cache-test-same-stat"
	artifactPath := writeCacheTestRuntimeArtifact(t, service.runtimeStorageDir, pluginID, "0.1.0")
	manifest, err := service.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("initial load failed: %v", err)
	}
	if manifest.Version != "0.1.0" {
		t.Fatalf("expected initial version 0.1.0, got %s", manifest.Version)
	}
	info, err := os.Stat(artifactPath)
	if err != nil {
		t.Fatalf("stat initial artifact: %v", err)
	}

	replacementPath := filepath.Join(service.runtimeStorageDir, pluginID+"-replacement.wasm")
	writeCacheTestRuntimeArtifactAtPath(t, replacementPath, pluginID, "0.2.0")
	if err = os.Rename(replacementPath, artifactPath); err != nil {
		t.Fatalf("rename replacement artifact: %v", err)
	}
	if err = os.Truncate(artifactPath, info.Size()); err != nil {
		t.Fatalf("truncate replacement artifact: %v", err)
	}
	if err = os.Chtimes(artifactPath, info.ModTime(), info.ModTime()); err != nil {
		t.Fatalf("restore replacement artifact mtime: %v", err)
	}
	reloaded, err := service.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("reload after same-stat replace failed: %v", err)
	}
	if reloaded.Version != "0.2.0" {
		t.Fatalf("expected replaced version 0.2.0, got %s", reloaded.Version)
	}
	if got := service.runtimeArtifactParseCount(artifactPath); got != 2 {
		t.Fatalf("expected same-stat replacement to parse twice, got %d", got)
	}
}

// TestRuntimeManifestCacheInvalidatesByPlugin verifies explicit plugin-scoped
// invalidation forces only the target artifact to be reparsed.
func TestRuntimeManifestCacheInvalidatesByPlugin(t *testing.T) {
	service := newCacheTestCatalogService(t)
	firstPath := writeCacheTestRuntimeArtifact(t, service.runtimeStorageDir, "cache-test-invalidate-a", "0.1.0")
	secondPath := writeCacheTestRuntimeArtifact(t, service.runtimeStorageDir, "cache-test-invalidate-b", "0.1.0")
	if _, err := service.ScanManifests(); err != nil {
		t.Fatalf("initial scan failed: %v", err)
	}

	service.InvalidateManifestCache("cache-test-invalidate-a")
	if _, err := service.ScanManifests(); err != nil {
		t.Fatalf("scan after invalidation failed: %v", err)
	}
	if got := service.runtimeArtifactParseCount(firstPath); got != 2 {
		t.Fatalf("expected invalidated artifact to parse twice, got %d", got)
	}
	if got := service.runtimeArtifactParseCount(secondPath); got != 1 {
		t.Fatalf("expected unaffected artifact to stay cached, got %d", got)
	}
}

// TestRuntimeManifestCacheReturnsDetachedCopies verifies callers cannot mutate
// cached manifest state shared with later readers.
func TestRuntimeManifestCacheReturnsDetachedCopies(t *testing.T) {
	service := newCacheTestCatalogService(t)
	artifactPath := writeCacheTestRuntimeArtifact(t, service.runtimeStorageDir, "cache-test-clone", "0.1.0")
	manifest, err := service.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("initial load failed: %v", err)
	}
	manifest.Name = "mutated"
	if manifest.RuntimeArtifact != nil && manifest.RuntimeArtifact.Manifest != nil {
		manifest.RuntimeArtifact.Manifest.Name = "mutated"
	}

	cached, err := service.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("cached load failed: %v", err)
	}
	if cached.Name == "mutated" {
		t.Fatal("expected cached manifest to be detached from caller mutation")
	}
	if cached.RuntimeArtifact != nil && cached.RuntimeArtifact.Manifest != nil &&
		cached.RuntimeArtifact.Manifest.Name == "mutated" {
		t.Fatal("expected cached artifact manifest to be detached from caller mutation")
	}
}

// TestSourceManifestCacheDetectsFileReplacement verifies filesystem-backed
// source plugins can be rescanned in the same process after plugin.yaml changes.
func TestSourceManifestCacheDetectsFileReplacement(t *testing.T) {
	var (
		service  = newCacheTestCatalogService(t)
		pluginID = "cache-test-source-replace"
		rootDir  = t.TempDir()
	)
	writeCacheTestSourcePlugin(t, rootDir, pluginID, "0.1.0")
	sourcePlugin := pluginhost.NewDeclarations(pluginID)
	sourcePlugin.Assets().UseEmbeddedFiles(os.DirFS(rootDir))
	cleanup, err := pluginhost.RegisterSourcePluginForTest(sourcePlugin)
	if err != nil {
		t.Fatalf("register source plugin fixture: %v", err)
	}
	t.Cleanup(cleanup)

	first, err := service.ScanManifests()
	if err != nil {
		t.Fatalf("scan first source manifests: %v", err)
	}
	if got := findCacheTestManifestVersion(first, pluginID); got != "0.1.0" {
		t.Fatalf("expected initial source version 0.1.0, got %q", got)
	}

	time.Sleep(time.Millisecond)
	writeCacheTestSourcePlugin(t, rootDir, pluginID, "0.2.0")
	second, err := service.ScanManifests()
	if err != nil {
		t.Fatalf("scan replaced source manifests: %v", err)
	}
	if got := findCacheTestManifestVersion(second, pluginID); got != "0.2.0" {
		t.Fatalf("expected replaced source version 0.2.0, got %q", got)
	}
}

// TestSourceManifestCacheReturnsDetachedCopies verifies source manifests are
// cloned before returning cached state to callers.
func TestSourceManifestCacheReturnsDetachedCopies(t *testing.T) {
	var (
		service  = newCacheTestCatalogService(t)
		pluginID = "cache-test-source-clone"
		rootDir  = t.TempDir()
	)
	writeCacheTestSourcePlugin(t, rootDir, pluginID, "0.1.0")
	sourcePlugin := pluginhost.NewDeclarations(pluginID)
	sourcePlugin.Assets().UseEmbeddedFiles(os.DirFS(rootDir))
	cleanup, err := pluginhost.RegisterSourcePluginForTest(sourcePlugin)
	if err != nil {
		t.Fatalf("register source plugin fixture: %v", err)
	}
	t.Cleanup(cleanup)

	first, err := service.ScanManifests()
	if err != nil {
		t.Fatalf("scan first source manifests: %v", err)
	}
	for _, manifest := range first {
		if manifest != nil && manifest.ID == pluginID {
			manifest.Name = "mutated"
		}
	}

	second, err := service.ScanManifests()
	if err != nil {
		t.Fatalf("scan cached source manifests: %v", err)
	}
	for _, manifest := range second {
		if manifest == nil || manifest.ID != pluginID {
			continue
		}
		if manifest.Name == "mutated" {
			t.Fatal("expected source manifest cache to return detached clones")
		}
		return
	}
	t.Fatalf("expected source plugin %s in scan result", pluginID)
}

// TestSourceManifestCacheDetectsSourcePluginReplacement verifies same-ID
// source-plugin registration replacements refresh callback/provider state even
// when plugin.yaml content is unchanged.
func TestSourceManifestCacheDetectsSourcePluginReplacement(t *testing.T) {
	var (
		service  = newCacheTestCatalogService(t)
		pluginID = "cache-test-source-registry-replace"
		rootDir  = t.TempDir()
	)
	writeCacheTestSourcePlugin(t, rootDir, pluginID, "0.1.0")
	firstPlugin := pluginhost.NewDeclarations(pluginID)
	firstPlugin.Assets().UseEmbeddedFiles(os.DirFS(rootDir))
	cleanup, err := pluginhost.RegisterSourcePluginForTest(firstPlugin)
	if err != nil {
		t.Fatalf("register first source plugin fixture: %v", err)
	}
	defer cleanup()

	first, err := service.ScanManifests()
	if err != nil {
		t.Fatalf("scan first source manifests: %v", err)
	}
	firstManifest := findCacheTestManifest(first, pluginID)
	if firstManifest == nil || firstManifest.SourcePlugin == nil {
		t.Fatalf("expected first source plugin manifest, got %#v", firstManifest)
	}

	secondPlugin := pluginhost.NewDeclarations(pluginID)
	secondPlugin.Assets().UseEmbeddedFiles(os.DirFS(rootDir))
	cleanupSecond, err := pluginhost.RegisterSourcePluginForTest(secondPlugin)
	if err != nil {
		t.Fatalf("register second source plugin fixture: %v", err)
	}
	defer cleanupSecond()

	second, err := service.ScanManifests()
	if err != nil {
		t.Fatalf("scan replaced source manifests: %v", err)
	}
	secondManifest := findCacheTestManifest(second, pluginID)
	if secondManifest == nil || secondManifest.SourcePlugin == nil {
		t.Fatalf("expected replaced source plugin manifest, got %#v", secondManifest)
	}
	if secondManifest.SourcePlugin == firstManifest.SourcePlugin {
		t.Fatal("expected source plugin registry replacement to refresh cached manifest")
	}
}

// createCacheTestRuntimeArtifacts creates count dynamic artifacts with stable
// plugin IDs for scan-boundary tests.
func createCacheTestRuntimeArtifacts(t *testing.T, storageDir string, count int, version string) []string {
	t.Helper()
	paths := make([]string, 0, count)
	for index := 0; index < count; index++ {
		pluginID := fmt.Sprintf("cache-test-plugin-%03d", index)
		paths = append(paths, writeCacheTestRuntimeArtifact(t, storageDir, pluginID, version))
	}
	return paths
}

// writeCacheTestRuntimeArtifact writes one dynamic artifact into test storage.
func writeCacheTestRuntimeArtifact(t *testing.T, storageDir string, pluginID string, version string) string {
	t.Helper()
	artifactPath := filepath.Join(storageDir, pluginID+".wasm")
	writeCacheTestRuntimeArtifactAtPath(t, artifactPath, pluginID, version)
	return artifactPath
}

// writeCacheTestRuntimeArtifactAtPath writes one dynamic artifact at an exact path.
func writeCacheTestRuntimeArtifactAtPath(t *testing.T, artifactPath string, pluginID string, version string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(artifactPath), 0o755); err != nil {
		t.Fatalf("failed to create artifact dir: %v", err)
	}
	supportsMultiTenant := true
	manifestContent, err := json.Marshal(&ArtifactManifest{
		ID:                  pluginID,
		Name:                "Cache Test " + pluginID,
		Version:             version,
		Type:                plugintypes.TypeDynamic.String(),
		ScopeNature:         plugintypes.ScopeNatureTenantAware.String(),
		SupportsMultiTenant: &supportsMultiTenant,
		DefaultInstallMode:  plugintypes.InstallModeTenantScoped.String(),
	})
	if err != nil {
		t.Fatalf("failed to marshal manifest: %v", err)
	}
	runtimeContent, err := json.Marshal(&ArtifactSpec{
		RuntimeKind: protocol.RuntimeKindWasm,
		ABIVersion:  protocol.SupportedABIVersion,
	})
	if err != nil {
		t.Fatalf("failed to marshal runtime metadata: %v", err)
	}
	wasm := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	wasm = appendCacheTestWasmCustomSection(wasm, protocol.WasmSectionManifest, manifestContent)
	wasm = appendCacheTestWasmCustomSection(wasm, protocol.WasmSectionRuntime, runtimeContent)
	if err := os.WriteFile(artifactPath, wasm, 0o644); err != nil {
		t.Fatalf("failed to write runtime artifact: %v", err)
	}
}

func writeCacheTestSourcePlugin(t *testing.T, rootDir string, pluginID string, version string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(rootDir, "backend"), 0o755); err != nil {
		t.Fatalf("create source backend dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rootDir, "go.mod"), []byte("module "+pluginID+"\n\ngo 1.25.0\n"), fs.FileMode(0o644)); err != nil {
		t.Fatalf("write source go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rootDir, "backend", "plugin.go"), []byte("package backend\n"), fs.FileMode(0o644)); err != nil {
		t.Fatalf("write source backend plugin.go: %v", err)
	}
	content := "id: " + pluginID + "\n" +
		"name: Cache Source " + pluginID + "\n" +
		"version: " + version + "\n" +
		"type: source\n" +
		"scope_nature: tenant_aware\n" +
		"supports_multi_tenant: true\n" +
		"default_install_mode: tenant_scoped\n"
	if err := os.WriteFile(filepath.Join(rootDir, resourcefs.EmbeddedManifestPath), []byte(content), fs.FileMode(0o644)); err != nil {
		t.Fatalf("write source plugin.yaml: %v", err)
	}
}

func findCacheTestManifestVersion(manifests []*Manifest, pluginID string) string {
	manifest := findCacheTestManifest(manifests, pluginID)
	if manifest == nil {
		return ""
	}
	return manifest.Version
}

func findCacheTestManifest(manifests []*Manifest, pluginID string) *Manifest {
	for _, manifest := range manifests {
		if manifest != nil && manifest.ID == pluginID {
			return manifest
		}
	}
	return nil
}

// cacheTestCatalogService carries the concrete catalog implementation and its
// isolated runtime storage directory.
type cacheTestCatalogService struct {
	*serviceImpl
	runtimeStorageDir string
}

// newCacheTestCatalogService returns a catalog service with isolated runtime
// storage so cache tests are independent from repository fixtures.
func newCacheTestCatalogService(t *testing.T) *cacheTestCatalogService {
	t.Helper()
	storageDir := t.TempDir()
	config := cacheTestConfigProvider{storageDir: storageDir}
	return &cacheTestCatalogService{
		serviceImpl:       New(config).(*serviceImpl),
		runtimeStorageDir: storageDir,
	}
}

// cacheTestConfigProvider supplies the runtime storage path to catalog tests.
type cacheTestConfigProvider struct {
	storageDir string
}

// GetPluginDynamicStoragePath returns the isolated runtime storage path.
func (p cacheTestConfigProvider) GetPluginDynamicStoragePath(context.Context) string {
	return p.storageDir
}

// appendCacheTestWasmCustomSection appends one WASM custom section.
func appendCacheTestWasmCustomSection(content []byte, name string, payload []byte) []byte {
	sectionPayload := append([]byte{}, encodeCacheTestWasmULEB128(uint32(len(name)))...)
	sectionPayload = append(sectionPayload, []byte(name)...)
	sectionPayload = append(sectionPayload, payload...)

	result := append([]byte{}, content...)
	result = append(result, 0x00)
	result = append(result, encodeCacheTestWasmULEB128(uint32(len(sectionPayload)))...)
	result = append(result, sectionPayload...)
	return result
}

// encodeCacheTestWasmULEB128 encodes a uint32 using WASM ULEB128 format.
func encodeCacheTestWasmULEB128(value uint32) []byte {
	result := make([]byte, 0, 5)
	for {
		current := byte(value & 0x7f)
		value >>= 7
		if value != 0 {
			current |= 0x80
		}
		result = append(result, current)
		if value == 0 {
			return result
		}
	}
}
