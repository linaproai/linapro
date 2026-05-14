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
	requireOfficialPluginDemoDynamic(t, pluginDir)
	prepareTemporaryPluginGoWorkForTest(t, repoRoot)
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

// requireOfficialPluginDemoDynamic skips plugin-full fixture checks when the
// official plugin submodule is not initialized in a host-only checkout.
func requireOfficialPluginDemoDynamic(t *testing.T, pluginDir string) {
	t.Helper()

	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	if _, err := os.Stat(manifestPath); err != nil {
		if os.IsNotExist(err) {
			t.Skip("official plugin workspace is not initialized")
		}
		t.Fatalf("stat dynamic demo plugin manifest failed: %v", err)
	}
}

// prepareTemporaryPluginGoWorkForTest mirrors linactl's ignored plugin
// workspace generation for tests that call the builder package directly.
func prepareTemporaryPluginGoWorkForTest(t *testing.T, repoRoot string) {
	t.Helper()

	rootContent, err := os.ReadFile(filepath.Join(repoRoot, "go.work"))
	if err != nil {
		t.Fatalf("failed to read root go.work: %v", err)
	}
	version := testGoWorkVersion(string(rootContent))
	if version == "" {
		t.Fatal("root go.work is missing a go version directive")
	}

	workspacePath := filepath.Join(repoRoot, "temp", "go.work.plugins")
	uses := make([]string, 0)
	seen := make(map[string]struct{})
	addUse := func(use string) {
		normalized := strings.TrimPrefix(filepath.ToSlash(filepath.Clean(use)), "./")
		if normalized == "" || normalized == "apps/lina-plugins" || strings.HasPrefix(normalized, "apps/lina-plugins/") {
			return
		}
		if _, ok := seen[normalized]; ok {
			return
		}
		seen[normalized] = struct{}{}
		uses = append(uses, normalized)
	}
	for _, use := range testGoWorkUses(string(rootContent)) {
		addUse(use)
	}

	pluginUses := testPluginGoWorkUses(t, repoRoot)
	for _, use := range pluginUses {
		if _, ok := seen[use]; ok {
			continue
		}
		seen[use] = struct{}{}
		uses = append(uses, use)
	}

	var builder strings.Builder
	builder.WriteString("go ")
	builder.WriteString(version)
	builder.WriteString("\n\nuse (\n")
	for _, use := range uses {
		modulePath := filepath.Join(repoRoot, filepath.FromSlash(use))
		relativePath, err := filepath.Rel(filepath.Dir(workspacePath), modulePath)
		if err != nil {
			t.Fatalf("failed to render test plugin workspace path %s: %v", use, err)
		}
		builder.WriteString("\t")
		builder.WriteString(filepath.ToSlash(relativePath))
		builder.WriteString("\n")
	}
	builder.WriteString(")\n")

	if err = os.MkdirAll(filepath.Dir(workspacePath), 0o755); err != nil {
		t.Fatalf("failed to create test plugin workspace directory: %v", err)
	}
	if err = os.WriteFile(workspacePath, []byte(builder.String()), 0o644); err != nil {
		t.Fatalf("failed to write test plugin workspace: %v", err)
	}
}

// testPluginGoWorkUses discovers plugin modules for the test workspace.
func testPluginGoWorkUses(t *testing.T, repoRoot string) []string {
	t.Helper()

	pluginRoot := filepath.Join(repoRoot, "apps", "lina-plugins")
	uses := make([]string, 0)
	if err := filepath.WalkDir(pluginRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || entry.Name() != "go.mod" {
			return nil
		}
		relativePath, relErr := filepath.Rel(repoRoot, filepath.Dir(path))
		if relErr != nil {
			return relErr
		}
		uses = append(uses, filepath.ToSlash(relativePath))
		return nil
	}); err != nil {
		t.Fatalf("failed to scan official plugin modules: %v", err)
	}
	sort.Slice(uses, func(left int, right int) bool {
		leftDepth := strings.Count(uses[left], "/")
		rightDepth := strings.Count(uses[right], "/")
		if leftDepth != rightDepth {
			return leftDepth < rightDepth
		}
		return uses[left] < uses[right]
	})
	return uses
}

// testGoWorkVersion extracts the go directive from test workspace content.
func testGoWorkVersion(content string) string {
	for _, line := range strings.Split(content, "\n") {
		fields := strings.Fields(testStripGoWorkComment(line))
		if len(fields) >= 2 && fields[0] == "go" {
			return fields[1]
		}
	}
	return ""
}

// testGoWorkUses extracts use directives from test workspace content.
func testGoWorkUses(content string) []string {
	var (
		uses       []string
		inUseBlock bool
	)
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(testStripGoWorkComment(line))
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "use (") {
			inUseBlock = true
			continue
		}
		if inUseBlock {
			if trimmed == ")" {
				inUseBlock = false
				continue
			}
			if use := testFirstGoWorkField(trimmed); use != "" {
				uses = append(uses, use)
			}
			continue
		}
		if strings.HasPrefix(trimmed, "use ") {
			if use := testFirstGoWorkField(strings.TrimSpace(strings.TrimPrefix(trimmed, "use"))); use != "" && use != "(" {
				uses = append(uses, use)
			}
		}
	}
	return uses
}

// testStripGoWorkComment removes simple line comments from go.work syntax.
func testStripGoWorkComment(line string) string {
	if index := strings.Index(line, "//"); index >= 0 {
		return line[:index]
	}
	return line
}

// testFirstGoWorkField returns the first path-like token from one go.work line.
func testFirstGoWorkField(line string) string {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return ""
	}
	return strings.Trim(fields[0], "\"")
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
