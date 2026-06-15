// This file verifies that repository-owned dynamic plugin samples are packaged
// into runtime artifacts without depending on lina-core host internals.

package wasmbuilder

import (
	"encoding/base64"
	"encoding/json"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	bridgeplugin "lina-core/pkg/plugin/pluginbridge"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// TestPluginDemoDynamicRuntimeArtifactEmbedsReviewedAssets verifies that the
// dynamic demo plugin artifact embeds the reviewed plugin source assets.
func TestPluginDemoDynamicRuntimeArtifactEmbedsReviewedAssets(t *testing.T) {
	repoRoot, ok := findRuntimeBuildRepoRoot(".")
	if !ok {
		t.Fatal("expected builder test to resolve repo root")
	}

	pluginDir := filepath.Join(repoRoot, "apps", "lina-plugins", "linapro-demo-dynamic")
	requireOfficialPluginDemoDynamic(t, pluginDir)
	prepareIgnoredPluginDemoDynamicRuntimeConfigForTest(t, pluginDir)
	prepareTemporaryPluginGoWorkForTest(t, repoRoot)
	expectedFrontendAssets := mustCollectSourceFrontendAssets(t, pluginDir)
	expectedInstallSQLAssets := mustCollectSourceSQLAssets(t, pluginDir, "manifest/sql")
	expectedUninstallSQLAssets := mustCollectSourceSQLAssets(t, pluginDir, "manifest/sql/uninstall")
	expectedMockSQLAssets := mustCollectSourceSQLAssets(t, pluginDir, "manifest/sql/mock-data")
	expectedManifestResourcePaths := mustCollectSourceManifestResourcePaths(t, pluginDir)

	out, err := buildRuntimeWasmArtifactFromSource(pluginDir, t.TempDir())
	if err != nil {
		t.Fatalf("expected dynamic demo plugin build to succeed, got error: %v", err)
	}
	assertGeneratedWasmDispatcherDidNotLeak(t, pluginDir)

	sections, err := parseWasmCustomSections(out.Content)
	if err != nil {
		t.Fatalf("expected wasm custom sections to parse, got error: %v", err)
	}

	manifest := &dynamicArtifactManifest{}
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionManifest], manifest); err != nil {
		t.Fatalf("expected manifest section json to unmarshal, got error: %v", err)
	}
	if manifest.ID != "linapro-demo-dynamic" {
		t.Fatalf("expected dynamic demo plugin id, got %s", manifest.ID)
	}
	if manifest.Type != pluginTypeDynamic {
		t.Fatalf("expected dynamic demo plugin type %s, got %s", pluginTypeDynamic, manifest.Type)
	}

	metadata := &protocol.RuntimeArtifactMetadata{}
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
	if metadata.ManifestResourceCount != len(expectedManifestResourcePaths) {
		t.Fatalf("expected manifest resource count %d, got %#v", len(expectedManifestResourcePaths), metadata)
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

	var manifestResources []*manifestResource
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionManifestResources], &manifestResources); err != nil {
		t.Fatalf("expected manifest resources section json to unmarshal, got error: %v", err)
	}
	assertManifestResourcesMatchSource(t, expectedManifestResourcePaths, manifestResourcePaths(manifestResources))

	var lifecycleContracts []*protocol.LifecycleContract
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionBackendLifecycle], &lifecycleContracts); err != nil {
		t.Fatalf("expected lifecycle section json to unmarshal, got error: %v", err)
	}
	assertPluginDemoDynamicLifecycleContracts(t, lifecycleContracts)
}

// TestPluginDemoDynamicGeneratedDispatcherIsZeroReflection verifies the
// generated dispatcher for the official dynamic demo plugin avoids runtime
// reflection and directly calls typed controllers.
func TestPluginDemoDynamicGeneratedDispatcherIsZeroReflection(t *testing.T) {
	repoRoot, ok := findRuntimeBuildRepoRoot(".")
	if !ok {
		t.Fatal("expected builder test to resolve repo root")
	}

	pluginDir := filepath.Join(repoRoot, "apps", "lina-plugins", "linapro-demo-dynamic")
	requireOfficialPluginDemoDynamic(t, pluginDir)
	prepareTemporaryPluginGoWorkForTest(t, repoRoot)

	routeSources, _, err := collectRouteContracts(pluginDir, "linapro-demo-dynamic")
	if err != nil {
		t.Fatalf("expected route contracts to collect, got error: %v", err)
	}
	lifecycleSpecs, err := collectLifecycleSpecs(pluginDir, "linapro-demo-dynamic")
	if err != nil {
		t.Fatalf("expected lifecycle specs to collect, got error: %v", err)
	}
	cleanup, err := prepareGeneratedWasmDispatcher(pluginDir, "linapro-demo-dynamic", routeSources, lifecycleSpecs)
	if err != nil {
		t.Fatalf("expected generated dispatcher to prepare, got error: %v", err)
	}
	t.Cleanup(func() {
		if cleanupErr := cleanup(); cleanupErr != nil {
			t.Fatalf("generated dispatcher cleanup failed: %v", cleanupErr)
		}
	})

	generatedPath := filepath.Join(pluginDir, "backend", generatedDispatcherFileName)
	content, err := os.ReadFile(generatedPath)
	if err != nil {
		t.Fatalf("expected generated dispatcher to exist, got error: %v", err)
	}
	generated := string(content)
	if !strings.Contains(generated, `"lina-core/pkg/plugin/pluginbridge/protocol"`) ||
		!strings.Contains(generated, `bridgeplugin "lina-core/pkg/plugin/pluginbridge"`) {
		t.Fatalf("generated dispatcher must use protocol and dynamic bridge imports:\n%s", generated)
	}
	for _, forbidden := range []string{
		`"reflect"`,
		"NewGuestControllerRouteDispatcher",
		"MustNewGuestControllerRouteDispatcher",
		"reflect.",
		`"lina-core/pkg/pluginbridge"`,
		`"lina-core/pkg/plugin/capability/guest"`,
		"capabilityguest",
	} {
		if strings.Contains(generated, forbidden) {
			t.Fatalf("generated dispatcher must not contain %q:\n%s", forbidden, generated)
		}
	}
	for _, forbidden := range []string{
		"protocol.Runtime(",
		"protocol.Storage(",
		"protocol.HTTP(",
		"protocol.Network(",
		"protocol.Data(",
		"protocol.Cache(",
		"protocol.Lock(",
		"protocol.Config(",
		"protocol.Notify(",
		"protocol.Cron(",
		"protocol.HostConfig(",
		"protocol.Manifest(",
		"protocol.Org(",
		"protocol.Tenant(",
	} {
		if strings.Contains(generated, forbidden) {
			t.Fatalf("generated dispatcher must not call root business client %q:\n%s", forbidden, generated)
		}
	}
	if !strings.Contains(generated, "func HandleRequest(") {
		t.Fatalf("generated dispatcher must expose HandleRequest:\n%s", generated)
	}
	if !strings.Contains(generated, "generatedController1().BackendSummary") {
		t.Fatalf("generated dispatcher must directly call typed controllers:\n%s", generated)
	}
	if strings.Contains(generated, "var generatedController1 ") {
		t.Fatalf("generated dispatcher must not initialize controllers at package init:\n%s", generated)
	}
	if !strings.Contains(generated, "sync.Once") ||
		!strings.Contains(generated, "func generatedController1()") {
		t.Fatalf("generated dispatcher must lazily initialize controllers once:\n%s", generated)
	}
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

// prepareIgnoredPluginDemoDynamicRuntimeConfigForTest mirrors the local plugin
// runtime config file that is intentionally ignored by git but required in the
// packaged dynamic artifact contract.
func prepareIgnoredPluginDemoDynamicRuntimeConfigForTest(t *testing.T, pluginDir string) {
	t.Helper()

	configPath := filepath.Join(pluginDir, "manifest", "config", "config.yaml")
	if info, err := os.Stat(configPath); err == nil {
		if info.IsDir() {
			t.Fatalf("dynamic demo runtime config path is a directory: %s", configPath)
		}
		return
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat dynamic demo runtime config failed: %v", err)
	}

	templatePath := filepath.Join(pluginDir, "manifest", "config", "config.example.yaml")
	content, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("read dynamic demo runtime config template failed: %v", err)
	}
	if err = os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("write dynamic demo runtime config fixture failed: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
			t.Fatalf("remove dynamic demo runtime config fixture failed: %v", err)
		}
	})
}

// assertGeneratedWasmDispatcherDidNotLeak verifies temporary generated source
// files are removed after the runtime build completes.
func assertGeneratedWasmDispatcherDidNotLeak(t *testing.T, pluginDir string) {
	t.Helper()

	generatedPath := filepath.Join(pluginDir, "backend", generatedDispatcherFileName)
	if _, err := os.Stat(generatedPath); err != nil {
		if os.IsNotExist(err) {
			return
		}
		t.Fatalf("stat generated dispatcher failed: %v", err)
	}
	t.Fatalf("generated dispatcher leaked into plugin source: %s", generatedPath)
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
		relativePath, err := filepath.Rel(pluginDir, filePath)
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

// mustCollectSourceManifestResourcePaths loads every file under plugin
// manifest/ using the same path contract exposed by the runtime artifact's raw
// manifest resource view.
func mustCollectSourceManifestResourcePaths(t *testing.T, pluginDir string) []string {
	t.Helper()

	manifestDir := filepath.Join(pluginDir, "manifest")
	paths := make([]string, 0)
	if err := filepath.WalkDir(manifestDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry == nil {
			return nil
		}
		if path != manifestDir && shouldSkipEmbeddedDirectoryEntry(entry.Name()) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		relativePath, err := filepath.Rel(pluginDir, path)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(relativePath))
		return nil
	}); err != nil {
		t.Fatalf("failed to collect dynamic demo manifest resources: %v", err)
	}
	sort.Strings(paths)
	return paths
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

// assertManifestResourcesMatchSource compares artifact manifest resource paths
// against the plugin source tree and locks the two demo manifest.get paths used
// by the runtime E2E.
func assertManifestResourcesMatchSource(t *testing.T, expected []string, actual []string) {
	t.Helper()

	if strings.Join(actual, ",") != strings.Join(expected, ",") {
		t.Fatalf("expected manifest resources %#v, got %#v", expected, actual)
	}

	required := []string{
		"manifest/config/config.yaml",
		"manifest/config/profile.yaml",
	}
	for _, path := range required {
		if !testStringSliceContains(actual, path) {
			t.Fatalf("expected manifest resource %s to be packaged, got %#v", path, actual)
		}
	}
}

// testStringSliceContains reports whether items contains target.
func testStringSliceContains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
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

// assertPluginDemoDynamicLifecycleContracts verifies the official dynamic demo
// relies on controller auto-discovery while preserving the reviewed lifecycle surface.
func assertPluginDemoDynamicLifecycleContracts(t *testing.T, actual []*protocol.LifecycleContract) {
	t.Helper()

	const beforeInstallTimeoutMs = 120000

	expected := []protocol.LifecycleOperation{
		protocol.LifecycleOperationBeforeInstall,
		protocol.LifecycleOperationAfterInstall,
		protocol.LifecycleOperationBeforeUpgrade,
		protocol.LifecycleOperationUpgrade,
		protocol.LifecycleOperationAfterUpgrade,
		protocol.LifecycleOperationBeforeDisable,
		protocol.LifecycleOperationAfterDisable,
		protocol.LifecycleOperationBeforeUninstall,
		protocol.LifecycleOperationUninstall,
		protocol.LifecycleOperationAfterUninstall,
		protocol.LifecycleOperationBeforeTenantDisable,
		protocol.LifecycleOperationAfterTenantDisable,
		protocol.LifecycleOperationBeforeTenantDelete,
		protocol.LifecycleOperationAfterTenantDelete,
		protocol.LifecycleOperationBeforeInstallModeChange,
		protocol.LifecycleOperationAfterInstallModeChange,
	}
	if len(actual) != len(expected) {
		t.Fatalf("expected %d lifecycle contracts, got %#v", len(expected), actual)
	}

	byOperation := make(map[protocol.LifecycleOperation]*protocol.LifecycleContract, len(actual))
	for _, contract := range actual {
		byOperation[contract.Operation] = contract
	}
	for _, operation := range expected {
		contract, ok := byOperation[operation]
		if !ok {
			t.Fatalf("expected lifecycle operation %s, got %#v", operation, actual)
		}
		expectedRequestType := operation.String() + "Req"
		expectedInternalPath := "/__lifecycle" + bridgeplugin.BuildGuestControllerInternalPath(operation.String())
		if contract.RequestType != expectedRequestType || contract.InternalPath != expectedInternalPath {
			t.Fatalf("unexpected lifecycle contract for %s: %#v", operation, contract)
		}
		expectedTimeoutMs := 0
		if operation == protocol.LifecycleOperationBeforeInstall {
			expectedTimeoutMs = beforeInstallTimeoutMs
		}
		if contract.TimeoutMs != expectedTimeoutMs {
			t.Fatalf(
				"unexpected lifecycle timeout for %s: got %d want %d",
				operation,
				contract.TimeoutMs,
				expectedTimeoutMs,
			)
		}
	}
}
