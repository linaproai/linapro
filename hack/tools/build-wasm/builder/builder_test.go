package builder

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"lina-core/pkg/pluginbridge"
)

func TestBuildRuntimeWasmArtifactFromSourceEmbedsDeclaredAssets(t *testing.T) {
	pluginDir := t.TempDir()

	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dynamic-builder\nname: Dynamic Builder\nversion: v0.1.0\ntype: dynamic\nscope_nature: tenant_aware\nsupports_multi_tenant: true\ndefault_install_mode: tenant_scoped\ndescription: standalone builder test\nhostServices:\n  - service: runtime\n    methods:\n      - log.write\n      - state.get\n      - state.set\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "frontend", "pages", "standalone.html"),
		"<!doctype html><html><body>it works</body></html>",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "manifest", "sql", "001-plugin-dynamic-builder.sql"),
		"SELECT 1;",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "manifest", "i18n", "en-US", "plugin.json"),
		"{\n  \"plugin.plugin-dynamic-builder.name\": \"Dynamic Builder\"\n}\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "manifest", "i18n", "zh-CN", "apidoc", "plugin-api-main.json"),
		"{\n  \"plugins.plugin_dynamic_builder.paths.get.review_summary.meta.summary\": \"查询摘要\"\n}\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "manifest", "sql", "uninstall", "001-plugin-dynamic-builder.sql"),
		"SELECT 2;",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "hooks", "001-login.yaml"),
		"event: auth.login.succeeded\naction: sleep\ntimeoutMs: 50\nsleepMs: 10\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "resources", "001-records.yaml"),
		"key: records\ntype: table-list\ntable: plugin_runtime_records\nfields:\n  - name: id\n    column: id\n  - name: status\n    column: status\nfilters:\n  - param: status\n    column: status\n    operator: eq\norderBy:\n  column: id\n  direction: asc\noperations:\n  - query\n  - get\n  - update\nkeyField: id\nwritableFields:\n  - status\naccess: both\ndataScope:\n  userColumn: owner_user_id\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "crons", "001-review-summary.yaml"),
		"name: review-summary-sync\ndisplayName: Review Summary Sync\ndescription: refreshes plugin review summary state\npattern: \"@every 5m\"\nscope: all_node\nconcurrency: singleton\ntimeoutSeconds: 30\nrequestType: ReviewSummaryReq\ninternalPath: /review-summary\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "api", "dynamic", "v1", "review_summary.go"),
		"package v1\n\nimport \"github.com/gogf/gf/v2/frame/g\"\n\ntype ReviewSummaryReq struct {\n\tg.Meta `path:\"/review-summary\" method:\"get\" tags:\"动态插件示例\" summary:\"查询摘要\" dc:\"返回一个动态插件摘要\" access:\"login\" permission:\"plugin-dynamic-builder:review:view\" operLog:\"other\"`\n}\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "main.go"),
		"package main\n\nfunc main() {}\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin_embed.go"),
		"package main\n\nimport \"embed\"\n\n//go:embed plugin.yaml frontend manifest\nvar EmbeddedFiles embed.FS\n",
	)

	out, err := BuildRuntimeWasmArtifactFromSource(pluginDir)
	if err != nil {
		t.Fatalf("expected dynamic build to succeed, got error: %v", err)
	}
	if out.RuntimePath != "" {
		t.Cleanup(func() {
			_ = os.RemoveAll(filepath.Dir(out.RuntimePath))
		})
	}
	if expected := filepath.Join(pluginDir, "temp", "plugin-dynamic-builder.wasm"); out.ArtifactPath != expected {
		t.Fatalf("expected artifact path %s, got %s", expected, out.ArtifactPath)
	}

	sections, err := parseWasmCustomSections(out.Content)
	if err != nil {
		t.Fatalf("expected wasm custom sections to parse, got error: %v", err)
	}

	manifest := &dynamicArtifactManifest{}
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionManifest], manifest); err != nil {
		t.Fatalf("expected manifest section json to unmarshal, got error: %v", err)
	}
	if manifest.ID != "plugin-dynamic-builder" {
		t.Fatalf("expected embedded manifest id plugin-dynamic-builder, got %s", manifest.ID)
	}
	if manifest.ScopeNature != pluginScopeNatureTenantAware {
		t.Fatalf("expected embedded scope nature tenant_aware, got %s", manifest.ScopeNature)
	}
	if manifest.SupportsMultiTenant == nil || !*manifest.SupportsMultiTenant {
		t.Fatalf("expected embedded supportsMultiTenant=true, got %#v", manifest.SupportsMultiTenant)
	}
	if manifest.DefaultInstallMode != pluginInstallModeTenantScoped {
		t.Fatalf("expected embedded default install mode tenant_scoped, got %s", manifest.DefaultInstallMode)
	}

	metadata := &dynamicArtifactMetadata{}
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionDynamic], metadata); err != nil {
		t.Fatalf("expected dynamic section json to unmarshal, got error: %v", err)
	}
	if metadata.FrontendAssetCount != 1 || metadata.I18NAssetCount != 1 || metadata.APIDocI18NAssetCount != 1 || metadata.SQLAssetCount != 2 {
		t.Fatalf("expected dynamic metadata counts 1/1/1/2, got %#v", metadata)
	}

	var frontend []*frontendAsset
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionFrontend], &frontend); err != nil {
		t.Fatalf("expected frontend section json to unmarshal, got error: %v", err)
	}
	if len(frontend) != 1 || frontend[0].Path != "standalone.html" {
		t.Fatalf("unexpected embedded frontend assets: %#v", frontend)
	}

	var i18n []*i18nAsset
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionI18N], &i18n); err != nil {
		t.Fatalf("expected i18n section json to unmarshal, got error: %v", err)
	}
	if len(i18n) != 1 || i18n[0].Locale != "en-US" || !strings.Contains(i18n[0].Content, "plugin.plugin-dynamic-builder.name") {
		t.Fatalf("unexpected embedded i18n assets: %#v", i18n)
	}

	var apiDocI18N []*i18nAsset
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionAPIDocI18N], &apiDocI18N); err != nil {
		t.Fatalf("expected apidoc i18n section json to unmarshal, got error: %v", err)
	}
	if len(apiDocI18N) != 1 || apiDocI18N[0].Locale != "zh-CN" || !strings.Contains(apiDocI18N[0].Content, "plugins.plugin_dynamic_builder") {
		t.Fatalf("unexpected embedded apidoc i18n assets: %#v", apiDocI18N)
	}

	var hooks []*hookSpec
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionBackendHooks], &hooks); err != nil {
		t.Fatalf("expected hook section json to unmarshal, got error: %v", err)
	}
	if len(hooks) != 1 || hooks[0].Action != hookActionSleep {
		t.Fatalf("unexpected embedded hook specs: %#v", hooks)
	}

	var resources []*resourceSpec
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionBackendRes], &resources); err != nil {
		t.Fatalf("expected resource section json to unmarshal, got error: %v", err)
	}
	if len(resources) != 1 || resources[0].DataScope == nil || resources[0].DataScope.UserColumn != "owner_user_id" {
		t.Fatalf("unexpected embedded resource specs: %#v", resources)
	}
	if resources[0].KeyField != "id" || len(resources[0].WritableFields) != 1 || resources[0].WritableFields[0] != "status" {
		t.Fatalf("unexpected embedded resource write contract: %#v", resources[0])
	}
	if resources[0].Access != "both" || len(resources[0].Operations) != 3 || resources[0].Operations[1] != "query" {
		t.Fatalf("unexpected embedded resource governance fields: %#v", resources[0])
	}

	if _, ok := sections[pluginDynamicWasmSectionBackendCrons]; ok {
		t.Fatalf("expected legacy cron declaration section to be omitted, got %#v", sections[pluginDynamicWasmSectionBackendCrons])
	}

	var routes []*pluginbridge.RouteContract
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionBackendRoutes], &routes); err != nil {
		t.Fatalf("expected route section json to unmarshal, got error: %v", err)
	}
	if len(routes) != 1 || routes[0].Permission != "plugin-dynamic-builder:review:view" {
		t.Fatalf("unexpected embedded route specs: %#v", routes)
	}
	if routes[0].Meta["operLog"] != "other" {
		t.Fatalf("expected custom route metadata to preserve operLog, got %#v", routes[0].Meta)
	}

	bridgeSpec := &pluginbridge.BridgeSpec{}
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionBackendBridge], bridgeSpec); err != nil {
		t.Fatalf("expected bridge section json to unmarshal, got error: %v", err)
	}
	if !bridgeSpec.RouteExecution || bridgeSpec.RequestCodec != pluginbridge.CodecProtobuf {
		t.Fatalf("unexpected embedded bridge spec: %#v", bridgeSpec)
	}

	if _, ok := sections[pluginbridge.WasmSectionBackendCapabilities]; ok {
		t.Fatalf("expected deprecated capabilities section to be omitted, got %#v", sections[pluginbridge.WasmSectionBackendCapabilities])
	}

	var hostServices []*pluginbridge.HostServiceSpec
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionBackendHostServices], &hostServices); err != nil {
		t.Fatalf("expected host services section json to unmarshal, got error: %v", err)
	}
	if len(hostServices) != 1 || hostServices[0].Service != pluginbridge.HostServiceRuntime {
		t.Fatalf("unexpected embedded host services: %#v", hostServices)
	}

	if out.RuntimePath == "" {
		t.Fatal("expected executable guest runtime path to be generated")
	}
	if _, err = os.Stat(filepath.Join(pluginDir, "temp", "runtime-plugin.wasm")); !os.IsNotExist(err) {
		t.Fatalf("expected guest runtime wasm to stop being written into plugin temp/, got err=%v", err)
	}
	runtimeStrings, err := readCommandOutput("strings", out.RuntimePath)
	if err != nil {
		t.Fatalf("expected runtime wasm strings inspection to succeed, got error: %v", err)
	}
	if !strings.Contains(runtimeStrings, "_initialize") {
		t.Fatalf("expected runtime guest wasm to expose _initialize, got output: %s", runtimeStrings)
	}
}

func TestBuildRuntimeWasmArtifactFromSourceFailsWhenEmbeddedResourcesOmitManifest(t *testing.T) {
	pluginDir := t.TempDir()

	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dynamic-missing-embed\nname: Dynamic Missing Embed\nversion: v0.1.0\ntype: dynamic\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "frontend", "pages", "standalone.html"),
		"<!doctype html><html><body>it works</body></html>",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin_embed.go"),
		"package main\n\nimport \"embed\"\n\n//go:embed frontend\nvar EmbeddedFiles embed.FS\n",
	)

	_, err := BuildRuntimeWasmArtifactFromSource(pluginDir)
	if err == nil {
		t.Fatal("expected embedded resource build without plugin.yaml to fail")
	}
	if !strings.Contains(err.Error(), "missing plugin.yaml") {
		t.Fatalf("expected missing embedded manifest error, got %v", err)
	}
}

func TestBuildRuntimeWasmArtifactFromSourceRejectsDeprecatedCapabilitiesDeclaration(t *testing.T) {
	pluginDir := t.TempDir()

	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dynamic-legacy-db\nname: Dynamic Legacy DB\nversion: v0.1.0\ntype: dynamic\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\ncapabilities:\n  - host:runtime\n  - host:db:query\nhostServices:\n  - service: runtime\n    methods:\n      - info.uuid\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "frontend", "pages", "standalone.html"),
		"<!doctype html><html><body>legacy capability</body></html>",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "main.go"),
		"package main\n\nfunc main() {}\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin_embed.go"),
		"package main\n\nimport \"embed\"\n\n//go:embed plugin.yaml frontend\nvar EmbeddedFiles embed.FS\n",
	)

	_, err := BuildRuntimeWasmArtifactFromSource(pluginDir)
	if err == nil {
		t.Fatal("expected deprecated top-level capabilities to be rejected")
	}
	if !strings.Contains(err.Error(), "host:db:query") || !strings.Contains(err.Error(), "hostServices") {
		t.Fatalf("expected deprecated capabilities error mentioning hostServices migration, got %v", err)
	}
}

func TestBuildRuntimeWasmArtifactFromSourceSkipsHiddenEmbeddedDirectoryEntries(t *testing.T) {
	pluginDir := t.TempDir()

	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dynamic-hidden\nname: Dynamic Hidden\nversion: v0.1.0\ntype: dynamic\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "frontend", "pages", "visible.html"),
		"<!doctype html><html><body>visible</body></html>",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "frontend", "pages", ".ignored.html"),
		"hidden",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "frontend", "pages", "_draft.html"),
		"draft",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "frontend", "pages", ".cache", "ghost.html"),
		"ghost",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "manifest", "sql", "001-plugin-dynamic-hidden.sql"),
		"SELECT 1;",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "manifest", "sql", ".ignored.sql"),
		"SELECT 0;",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "manifest", "sql", "_draft.sql"),
		"SELECT -1;",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "manifest", "sql", "mock-data", "001-plugin-dynamic-hidden-mock-data.sql"),
		"SELECT 99;",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "manifest", "sql", "uninstall", "001-plugin-dynamic-hidden.sql"),
		"SELECT 2;",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "manifest", "sql", "uninstall", ".ignored.sql"),
		"SELECT 3;",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin_embed.go"),
		"package main\n\nimport \"embed\"\n\n//go:embed plugin.yaml frontend manifest\nvar EmbeddedFiles embed.FS\n",
	)

	out, err := BuildRuntimeWasmArtifactFromSource(pluginDir)
	if err != nil {
		t.Fatalf("expected hidden-entry build to succeed, got error: %v", err)
	}

	sections, err := parseWasmCustomSections(out.Content)
	if err != nil {
		t.Fatalf("expected wasm custom sections to parse, got error: %v", err)
	}

	var frontend []*frontendAsset
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionFrontend], &frontend); err != nil {
		t.Fatalf("expected frontend section json to unmarshal, got error: %v", err)
	}
	if len(frontend) != 1 || frontend[0].Path != "visible.html" {
		t.Fatalf("expected only visible embedded frontend asset, got %#v", frontend)
	}

	var installSQL []*sqlAsset
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionInstallSQL], &installSQL); err != nil {
		t.Fatalf("expected install sql section json to unmarshal, got error: %v", err)
	}
	if len(installSQL) != 1 || installSQL[0].Key != "001-plugin-dynamic-hidden.sql" {
		t.Fatalf("expected only visible install sql asset, got %#v", installSQL)
	}

	var uninstallSQL []*sqlAsset
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionUninstallSQL], &uninstallSQL); err != nil {
		t.Fatalf("expected uninstall sql section json to unmarshal, got error: %v", err)
	}
	if len(uninstallSQL) != 1 || uninstallSQL[0].Key != "001-plugin-dynamic-hidden.sql" {
		t.Fatalf("expected only visible uninstall sql asset, got %#v", uninstallSQL)
	}

	// Mock-data SQL ships in its own dedicated section so the host can detect
	// mock-data presence without scanning the install section, and the
	// optional mock-data load phase can pull from it independently.
	var mockSQL []*sqlAsset
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionMockSQL], &mockSQL); err != nil {
		t.Fatalf("expected mock sql section json to unmarshal, got error: %v", err)
	}
	if len(mockSQL) != 1 || mockSQL[0].Key != "001-plugin-dynamic-hidden-mock-data.sql" {
		t.Fatalf("expected mock-data sql asset to land in the mock section, got %#v", mockSQL)
	}
}

func TestBuildRuntimeWasmArtifactFromSourceCleansTemporaryGoMod(t *testing.T) {
	pluginDir := t.TempDir()
	outputDir := t.TempDir()

	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dynamic-temp-gomod\nname: Dynamic Temp GoMod\nversion: v0.1.0\ntype: dynamic\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "main.go"),
		"package main\n\nfunc main() {}\n",
	)

	out, err := buildRuntimeWasmArtifactFromSource(pluginDir, outputDir)
	if err != nil {
		t.Fatalf("expected build without go.mod to succeed, got error: %v", err)
	}
	if out.RuntimePath != "" {
		t.Cleanup(func() {
			_ = os.RemoveAll(filepath.Dir(out.RuntimePath))
		})
	}

	if _, err = os.Stat(filepath.Join(pluginDir, "go.mod")); !os.IsNotExist(err) {
		t.Fatalf("expected temporary go.mod to be cleaned up, got err=%v", err)
	}
	if _, err = os.Stat(filepath.Join(pluginDir, "go.sum")); !os.IsNotExist(err) {
		t.Fatalf("expected temporary go.sum to be cleaned up, got err=%v", err)
	}
}

func TestWriteRuntimeWasmArtifactFromSourceWritesGeneratedFile(t *testing.T) {
	pluginDir := t.TempDir()
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dynamic-write\nname: Dynamic Write\nversion: v0.1.0\ntype: dynamic\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\n",
	)

	repoRoot, ok := findRuntimeBuildRepoRoot(".")
	if !ok {
		t.Fatal("expected builder test to resolve repo root")
	}
	out, err := WriteRuntimeWasmArtifactFromSource(pluginDir, "")
	if err != nil {
		t.Fatalf("expected dynamic artifact write to succeed, got error: %v", err)
	}
	expectedPath := filepath.Join(repoRoot, defaultRuntimeOutputDir, "plugin-dynamic-write.wasm")
	if out.ArtifactPath != expectedPath {
		t.Fatalf("expected generated dynamic artifact path %s, got %s", expectedPath, out.ArtifactPath)
	}
	t.Cleanup(func() {
		_ = os.Remove(out.ArtifactPath)
		_ = os.RemoveAll(filepath.Join(repoRoot, defaultRuntimeOutputDir, runtimeWorkspaceDirName, "plugin-dynamic-write"))
	})

	content, err := os.ReadFile(out.ArtifactPath)
	if err != nil {
		t.Fatalf("expected generated dynamic artifact to exist, got error: %v", err)
	}
	if len(content) == 0 {
		t.Fatalf("expected generated dynamic artifact to contain bytes")
	}
	if _, err = os.Stat(filepath.Join(pluginDir, "temp", "plugin-dynamic-write.wasm")); !os.IsNotExist(err) {
		t.Fatalf("expected generated dynamic artifact to stop being written into plugin temp/, got err=%v", err)
	}
}

func TestWriteRuntimeWasmArtifactFromSourceSupportsExternalOutputDir(t *testing.T) {
	pluginDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "output")
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dynamic-output\nname: Dynamic Output\nversion: v0.1.0\ntype: dynamic\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\n",
	)

	out, err := WriteRuntimeWasmArtifactFromSource(pluginDir, outputDir)
	if err != nil {
		t.Fatalf("expected dynamic artifact write to external dir to succeed, got error: %v", err)
	}
	if expected := filepath.Join(outputDir, "plugin-dynamic-output.wasm"); out.ArtifactPath != expected {
		t.Fatalf("expected generated dynamic artifact path %s, got %s", expected, out.ArtifactPath)
	}
	if _, err = os.Stat(out.ArtifactPath); err != nil {
		t.Fatalf("expected generated dynamic artifact to exist in external dir, got error: %v", err)
	}
	if _, err = os.Stat(filepath.Join(pluginDir, "temp", "runtime-plugin.wasm")); !os.IsNotExist(err) {
		t.Fatalf("expected guest runtime wasm to stop being written into plugin temp/, got err=%v", err)
	}
}

func TestWriteRuntimeWasmArtifactFromSourceSupportsRelativeOutputDir(t *testing.T) {
	pluginDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "output")
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dynamic-relative-output\nname: Dynamic Relative Output\nversion: v0.1.0\ntype: dynamic\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "main.go"),
		"package main\n\nfunc main() {}\n",
	)

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("expected builder test to resolve current working directory, got error: %v", err)
	}
	relativeOutputDir, err := filepath.Rel(workingDir, outputDir)
	if err != nil {
		t.Fatalf("expected builder test to compute relative output dir, got error: %v", err)
	}

	out, err := WriteRuntimeWasmArtifactFromSource(pluginDir, relativeOutputDir)
	if err != nil {
		t.Fatalf("expected dynamic artifact write to relative dir to succeed, got error: %v", err)
	}
	if expected := filepath.Join(outputDir, "plugin-dynamic-relative-output.wasm"); out.ArtifactPath != expected {
		t.Fatalf("expected generated dynamic artifact path %s, got %s", expected, out.ArtifactPath)
	}
	if expected := filepath.Join(outputDir, runtimeWorkspaceDirName, "plugin-dynamic-relative-output", "runtime-plugin.wasm"); out.RuntimePath != expected {
		t.Fatalf("expected generated guest runtime path %s, got %s", expected, out.RuntimePath)
	}
	if _, err = os.Stat(out.ArtifactPath); err != nil {
		t.Fatalf("expected generated dynamic artifact to exist in relative output dir, got error: %v", err)
	}
	if _, err = os.Stat(out.RuntimePath); err != nil {
		t.Fatalf("expected generated guest runtime to exist in relative output dir, got error: %v", err)
	}
}

func mustWriteFile(t *testing.T, filePath string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("failed to create directory for %s: %v", filePath, err)
	}
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write file %s: %v", filePath, err)
	}
}

func readCommandOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func parseWasmCustomSections(content []byte) (map[string][]byte, error) {
	if len(content) < 8 {
		return nil, os.ErrInvalid
	}
	if string(content[:4]) != "\x00asm" {
		return nil, os.ErrInvalid
	}

	sections := make(map[string][]byte)
	cursor := 8
	for cursor < len(content) {
		sectionID := content[cursor]
		cursor++

		sectionSize, nextCursor, err := readULEB128(content, cursor)
		if err != nil {
			return nil, err
		}
		cursor = nextCursor

		end := cursor + int(sectionSize)
		if end > len(content) {
			return nil, os.ErrInvalid
		}

		if sectionID == 0 {
			nameLength, nameCursor, err := readULEB128(content, cursor)
			if err != nil {
				return nil, err
			}
			nameEnd := nameCursor + int(nameLength)
			if nameEnd > end {
				return nil, os.ErrInvalid
			}
			name := string(content[nameCursor:nameEnd])
			sections[name] = append([]byte(nil), content[nameEnd:end]...)
		}
		cursor = end
	}

	return sections, nil
}

func readULEB128(content []byte, cursor int) (uint32, int, error) {
	var (
		shift uint
		value uint32
	)

	for {
		if cursor >= len(content) {
			return 0, cursor, os.ErrInvalid
		}
		part := content[cursor]
		cursor++

		value |= uint32(part&0x7f) << shift
		if part&0x80 == 0 {
			return value, cursor, nil
		}
		shift += 7
		if shift > 28 {
			return 0, cursor, os.ErrInvalid
		}
	}
}
