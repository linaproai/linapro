// This file verifies that repository-owned dynamic plugin samples are packaged
// into runtime artifacts without depending on lina-core host internals.

package builder

import (
	"encoding/base64"
	"encoding/json"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// TestPluginDemoDynamicRuntimeArtifactEmbedsReviewedAssets verifies that the
// dynamic demo plugin artifact embeds the reviewed plugin source assets.
func TestPluginDemoDynamicRuntimeArtifactEmbedsReviewedAssets(t *testing.T) {
	repoRoot, ok := findRuntimeBuildRepoRoot(".")
	if !ok {
		t.Fatal("expected builder test to resolve repo root")
	}

	pluginDir := filepath.Join(repoRoot, "apps", "lina-plugins", "plugin-demo-dynamic")
	expectedFrontendAssets := mustCollectSourceFrontendAssets(t, pluginDir)
	expectedInstallSQLAssets := mustCollectSourceSQLAssets(t, pluginDir, "manifest/sql")
	expectedUninstallSQLAssets := mustCollectSourceSQLAssets(t, pluginDir, "manifest/sql/uninstall")
	expectedMockSQLAssets := mustCollectSourceSQLAssets(t, pluginDir, "manifest/sql/mock-data")

	out, err := buildRuntimeWasmArtifactFromSource(pluginDir, t.TempDir())
	if err != nil {
		t.Fatalf("expected dynamic demo plugin build to succeed, got error: %v", err)
	}

	sections, err := parseWasmCustomSections(out.Content)
	if err != nil {
		t.Fatalf("expected wasm custom sections to parse, got error: %v", err)
	}

	manifest := &dynamicArtifactManifest{}
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionManifest], manifest); err != nil {
		t.Fatalf("expected manifest section json to unmarshal, got error: %v", err)
	}
	if manifest.ID != "plugin-demo-dynamic" {
		t.Fatalf("expected dynamic demo plugin id, got %s", manifest.ID)
	}
	if manifest.Type != pluginTypeDynamic {
		t.Fatalf("expected dynamic demo plugin type %s, got %s", pluginTypeDynamic, manifest.Type)
	}

	metadata := &dynamicArtifactMetadata{}
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionDynamic], metadata); err != nil {
		t.Fatalf("expected runtime metadata section json to unmarshal, got error: %v", err)
	}
	if metadata.FrontendAssetCount != len(expectedFrontendAssets) {
		t.Fatalf("expected frontend asset count %d, got %d", len(expectedFrontendAssets), metadata.FrontendAssetCount)
	}
	expectedSQLAssetCount := len(expectedInstallSQLAssets) + len(expectedUninstallSQLAssets) + len(expectedMockSQLAssets)
	if metadata.SQLAssetCount != expectedSQLAssetCount || metadata.MockSQLAssetCount != len(expectedMockSQLAssets) {
		t.Fatalf("expected sql/mock asset counts %d/%d, got %#v", expectedSQLAssetCount, len(expectedMockSQLAssets), metadata)
	}

	var frontendAssets []*frontendAsset
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionFrontend], &frontendAssets); err != nil {
		t.Fatalf("expected frontend section json to unmarshal, got error: %v", err)
	}
	assertFrontendAssetsMatchSource(t, expectedFrontendAssets, frontendAssets)

	var installSQLAssets []*sqlAsset
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionInstallSQL], &installSQLAssets); err != nil {
		t.Fatalf("expected install sql section json to unmarshal, got error: %v", err)
	}
	assertSQLAssetsMatchSource(t, expectedInstallSQLAssets, installSQLAssets)

	var uninstallSQLAssets []*sqlAsset
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionUninstallSQL], &uninstallSQLAssets); err != nil {
		t.Fatalf("expected uninstall sql section json to unmarshal, got error: %v", err)
	}
	assertSQLAssetsMatchSource(t, expectedUninstallSQLAssets, uninstallSQLAssets)

	var mockSQLAssets []*sqlAsset
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionMockSQL], &mockSQLAssets); err != nil {
		t.Fatalf("expected mock sql section json to unmarshal, got error: %v", err)
	}
	assertSQLAssetsMatchSource(t, expectedMockSQLAssets, mockSQLAssets)
}

// mustCollectSourceFrontendAssets loads plugin frontend source files using the
// same path and content-type contract exposed by the runtime artifact.
func mustCollectSourceFrontendAssets(t *testing.T, pluginDir string) []*frontendAsset {
	t.Helper()

	frontendDir := filepath.Join(pluginDir, "frontend", "pages")
	paths := make([]string, 0)
	if err := filepath.WalkDir(frontendDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry == nil || entry.IsDir() {
			return nil
		}
		paths = append(paths, path)
		return nil
	}); err != nil {
		t.Fatalf("failed to collect dynamic demo frontend assets: %v", err)
	}
	sort.Strings(paths)

	assets := make([]*frontendAsset, 0, len(paths))
	for _, filePath := range paths {
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read dynamic demo frontend asset %s: %v", filePath, err)
		}
		relativePath, err := filepath.Rel(frontendDir, filePath)
		if err != nil {
			t.Fatalf("failed to resolve dynamic demo frontend asset path %s: %v", filePath, err)
		}
		contentType := mime.TypeByExtension(filepath.Ext(filePath))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		assets = append(assets, &frontendAsset{
			Path:          filepath.ToSlash(relativePath),
			ContentBase64: base64.StdEncoding.EncodeToString(content),
			ContentType:   contentType,
		})
	}
	return assets
}

// mustCollectSourceSQLAssets loads direct SQL files from a plugin source
// directory and preserves the artifact ordering contract.
func mustCollectSourceSQLAssets(t *testing.T, pluginDir string, relativeDir string) []*sqlAsset {
	t.Helper()

	searchDir := filepath.Join(pluginDir, filepath.FromSlash(relativeDir))
	entries, err := os.ReadDir(searchDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*sqlAsset{}
		}
		t.Fatalf("failed to collect dynamic demo SQL assets from %s: %v", searchDir, err)
	}

	fileNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		fileNames = append(fileNames, entry.Name())
	}
	sort.Strings(fileNames)

	assets := make([]*sqlAsset, 0, len(fileNames))
	for _, fileName := range fileNames {
		content, err := os.ReadFile(filepath.Join(searchDir, fileName))
		if err != nil {
			t.Fatalf("failed to read dynamic demo SQL asset %s: %v", fileName, err)
		}
		assets = append(assets, &sqlAsset{
			Key:     fileName,
			Content: strings.TrimSpace(string(content)),
		})
	}
	return assets
}

// assertFrontendAssetsMatchSource compares artifact frontend asset payloads
// against the source files by path, content type, and encoded content.
func assertFrontendAssetsMatchSource(t *testing.T, expected []*frontendAsset, actual []*frontendAsset) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf("expected %d frontend assets, got %d", len(expected), len(actual))
	}

	expectedByPath := make(map[string]*frontendAsset, len(expected))
	for _, asset := range expected {
		expectedByPath[asset.Path] = asset
	}
	for _, asset := range actual {
		expectedAsset, ok := expectedByPath[asset.Path]
		if !ok {
			t.Fatalf("unexpected frontend asset path: %s", asset.Path)
		}
		if asset.ContentType != expectedAsset.ContentType {
			t.Fatalf("expected frontend asset %s content type %s, got %s", asset.Path, expectedAsset.ContentType, asset.ContentType)
		}
		if asset.ContentBase64 != expectedAsset.ContentBase64 {
			t.Fatalf("unexpected frontend asset content for %s", asset.Path)
		}
	}
}

// assertSQLAssetsMatchSource compares ordered SQL artifact entries against the
// source files by file name and trimmed SQL content.
func assertSQLAssetsMatchSource(t *testing.T, expected []*sqlAsset, actual []*sqlAsset) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf("expected %d SQL assets, got %d", len(expected), len(actual))
	}
	for index := range expected {
		if actual[index].Key != expected[index].Key {
			t.Fatalf("expected SQL asset key %s, got %s", expected[index].Key, actual[index].Key)
		}
		if strings.TrimSpace(actual[index].Content) != strings.TrimSpace(expected[index].Content) {
			t.Fatalf("unexpected SQL content for asset %s", expected[index].Key)
		}
	}
}
