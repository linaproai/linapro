// This file verifies dynamic WASM builder resource packaging contracts.

package wasmbuilder

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"lina-core/pkg/plugin/pluginbridge/protocol"
)

func TestBuildRuntimeWasmArtifactFromSourceEmbedsDeclaredAssets(t *testing.T) {
	pluginDir := t.TempDir()

	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dev-dynamic-builder\nname: Dynamic Builder\nversion: v0.1.0\ntype: dynamic\nscope_nature: tenant_aware\nsupports_multi_tenant: true\ndefault_install_mode: tenant_scoped\ndescription: standalone builder test\ndependencies:\n  framework:\n    version: \">=0.1.0 <1.0.0\"\n  plugins:\n    - id: linapro-tenant-core\n      version: \">=0.1.0\"\nhostServices:\n  - service: runtime\n    methods:\n      - log.write\n      - state.get\n      - state.set\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "frontend", "pages", "standalone.html"),
		"<!doctype html><html><body>it works</body></html>",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "manifest", "sql", "001-plugin-dev-dynamic-builder.sql"),
		"SELECT 1;",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "manifest", "config", "config.example.yaml"),
		"monitor:\n  interval: 30s\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "manifest", "config", "config.yaml"),
		"monitor:\n  interval: 45s\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "manifest", "metadata.yaml"),
		"title: Dynamic Builder\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "manifest", "resources", "policy.yaml"),
		"enabled: true\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "manifest", "i18n", "en-US", "plugin.json"),
		"{\n  \"plugin.plugin-dev-dynamic-builder.name\": \"Dynamic Builder\"\n}\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "manifest", "i18n", "zh-CN", "apidoc", "plugin-api-main.json"),
		"{\n  \"plugins.plugin_dev_dynamic_builder.paths.get.review_summary.meta.summary\": \"查询摘要\"\n}\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "manifest", "sql", "uninstall", "001-plugin-dev-dynamic-builder.sql"),
		"SELECT 2;",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "hooks", "001-login.yaml"),
		"event: auth.login.succeeded\naction: sleep\ntimeoutMs: 50\nsleepMs: 10\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "lifecycle", "001-before-install.yaml"),
		"operation: BeforeInstall\nrequestType: BeforeInstallReq\ninternalPath: /__lifecycle/before-install\ntimeoutMs: 3000\n",
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
		filepath.Join(pluginDir, "backend", "api", "dynamic", "dynamic.go"),
		"package dynamicapi\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "api", "dynamic", "v1", "review_summary.go"),
		"package v1\n\nimport \"github.com/gogf/gf/v2/frame/g\"\n\ntype ReviewSummaryReq struct {\n\tg.Meta `path:\"/review-summary\" method:\"get\" tags:\"动态插件示例\" summary:\"查询摘要\" dc:\"返回一个动态插件摘要\" access:\"login\" permission:\"plugin-dev-dynamic-builder:review:view\" operLog:\"other\"`\n}\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "plugin.go"),
		"package backend\n\nimport bridgeguest \"lina-core/pkg/plugin/pluginbridge/guest\"\n\nfunc RegisterRoutes(registrar bridgeguest.DynamicRouteRegistrar) error {\n\treturn registrar.Group(\"/api/v1\", \"dynamic/v1\")\n}\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "controller.go"),
		lifecycleControllerSourceForTest("BeforeInstall"),
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
	if expected := filepath.Join(pluginDir, "temp", "plugin-dev-dynamic-builder.wasm"); out.ArtifactPath != expected {
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
	if manifest.ID != "plugin-dev-dynamic-builder" {
		t.Fatalf("expected embedded manifest id plugin-dev-dynamic-builder, got %s", manifest.ID)
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
	if manifest.Dependencies == nil || manifest.Dependencies.Framework == nil {
		t.Fatalf("expected embedded dependencies, got %#v", manifest.Dependencies)
	}
	if manifest.Dependencies.Framework.Version != ">=0.1.0 <1.0.0" {
		t.Fatalf("unexpected framework dependency: %#v", manifest.Dependencies.Framework)
	}
	if len(manifest.Dependencies.Plugins) != 1 {
		t.Fatalf("expected one embedded plugin dependency, got %#v", manifest.Dependencies.Plugins)
	}
	if manifest.Dependencies.Plugins[0].ID != "linapro-tenant-core" ||
		manifest.Dependencies.Plugins[0].Version != ">=0.1.0" {
		t.Fatalf("unexpected embedded plugin dependency: %#v", manifest.Dependencies.Plugins[0])
	}

	metadata := &protocol.RuntimeArtifactMetadata{}
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionDynamic], metadata); err != nil {
		t.Fatalf("expected dynamic section json to unmarshal, got error: %v", err)
	}
	if metadata.FrontendAssetCount != 1 ||
		metadata.I18NAssetCount != 1 ||
		metadata.APIDocI18NAssetCount != 1 ||
		metadata.SQLAssetCount != 2 ||
		metadata.ManifestResourceCount != 8 {
		t.Fatalf("expected dynamic metadata counts 1/1/1/2/8, got %#v", metadata)
	}

	var frontend []*frontendAsset
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionFrontend], &frontend); err != nil {
		t.Fatalf("expected frontend section json to unmarshal, got error: %v", err)
	}
	if len(frontend) != 1 || frontend[0].Path != "frontend/pages/standalone.html" {
		t.Fatalf("unexpected embedded frontend assets: %#v", frontend)
	}

	var i18n []*i18nAsset
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionI18N], &i18n); err != nil {
		t.Fatalf("expected i18n section json to unmarshal, got error: %v", err)
	}
	if len(i18n) != 1 || i18n[0].Locale != "en-US" || !strings.Contains(i18n[0].Content, "plugin.plugin-dev-dynamic-builder.name") {
		t.Fatalf("unexpected embedded i18n assets: %#v", i18n)
	}

	var apiDocI18N []*i18nAsset
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionAPIDocI18N], &apiDocI18N); err != nil {
		t.Fatalf("expected apidoc i18n section json to unmarshal, got error: %v", err)
	}
	if len(apiDocI18N) != 1 || apiDocI18N[0].Locale != "zh-CN" || !strings.Contains(apiDocI18N[0].Content, "plugins.plugin_dev_dynamic_builder") {
		t.Fatalf("unexpected embedded apidoc i18n assets: %#v", apiDocI18N)
	}

	var manifestResources []*manifestResource
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionManifestResources], &manifestResources); err != nil {
		t.Fatalf("expected manifest resource section json to unmarshal, got error: %v", err)
	}
	expectedManifestPaths := []string{
		"manifest/config/config.example.yaml",
		"manifest/config/config.yaml",
		"manifest/i18n/en-US/plugin.json",
		"manifest/i18n/zh-CN/apidoc/plugin-api-main.json",
		"manifest/metadata.yaml",
		"manifest/resources/policy.yaml",
		"manifest/sql/001-plugin-dev-dynamic-builder.sql",
		"manifest/sql/uninstall/001-plugin-dev-dynamic-builder.sql",
	}
	if got := manifestResourcePaths(manifestResources); strings.Join(got, ",") != strings.Join(expectedManifestPaths, ",") {
		t.Fatalf("expected manifest resource paths %#v, got %#v", expectedManifestPaths, got)
	}

	var hooks []*hookSpec
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionBackendHooks], &hooks); err != nil {
		t.Fatalf("expected hook section json to unmarshal, got error: %v", err)
	}
	if len(hooks) != 1 || hooks[0].Action != hookActionSleep {
		t.Fatalf("unexpected embedded hook specs: %#v", hooks)
	}

	var lifecycle []*protocol.LifecycleContract
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionBackendLifecycle], &lifecycle); err != nil {
		t.Fatalf("expected lifecycle section json to unmarshal, got error: %v", err)
	}
	if len(lifecycle) != 1 ||
		lifecycle[0].Operation != protocol.LifecycleOperationBeforeInstall ||
		lifecycle[0].RequestType != "BeforeInstallReq" ||
		lifecycle[0].InternalPath != "/__lifecycle/before-install" {
		t.Fatalf("unexpected embedded lifecycle specs: %#v", lifecycle)
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

	var routes []*protocol.RouteContract
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionBackendRoutes], &routes); err != nil {
		t.Fatalf("expected route section json to unmarshal, got error: %v", err)
	}
	if len(routes) != 1 || routes[0].Permission != "plugin-dev-dynamic-builder:review:view" {
		t.Fatalf("unexpected embedded route specs: %#v", routes)
	}
	if routes[0].Path != "/api/v1/review-summary" {
		t.Fatalf("expected route group prefix to be composed into route path, got %#v", routes[0])
	}
	if routes[0].Meta["operLog"] != "other" {
		t.Fatalf("expected custom route metadata to preserve operLog, got %#v", routes[0].Meta)
	}

	bridgeSpec := &protocol.BridgeSpec{}
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionBackendBridge], bridgeSpec); err != nil {
		t.Fatalf("expected bridge section json to unmarshal, got error: %v", err)
	}
	if !bridgeSpec.RouteExecution || bridgeSpec.RequestCodec != protocol.CodecProtobuf {
		t.Fatalf("unexpected embedded bridge spec: %#v", bridgeSpec)
	}

	var hostServices []*protocol.HostServiceSpec
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionBackendHostServices], &hostServices); err != nil {
		t.Fatalf("expected host services section json to unmarshal, got error: %v", err)
	}
	if len(hostServices) != 1 || hostServices[0].Service != protocol.HostServiceRuntime {
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

func TestCollectManifestResourcesScansDirectoryFallback(t *testing.T) {
	pluginDir := t.TempDir()
	mustWriteFile(t, filepath.Join(pluginDir, "manifest", "config", "config.example.yaml"), "name: example\n")
	mustWriteFile(t, filepath.Join(pluginDir, "manifest", "config", "config.yaml"), "name: actual\n")
	mustWriteFile(t, filepath.Join(pluginDir, "manifest", "metadata.yaml"), "title: demo\n")
	mustWriteFile(t, filepath.Join(pluginDir, "manifest", "resources", "policy.yaml"), "enabled: true\n")
	mustWriteFile(t, filepath.Join(pluginDir, "manifest", "resources", "ignored.json"), "{}\n")
	mustWriteFile(t, filepath.Join(pluginDir, "manifest", "sql", "001-demo.sql"), "SELECT 1;\n")
	mustWriteFile(t, filepath.Join(pluginDir, "manifest", "i18n", "zh-CN", "plugin.json"), "{}\n")

	resources, err := collectManifestResources(pluginDir, nil)
	if err != nil {
		t.Fatalf("expected manifest resource collection to succeed, got error: %v", err)
	}
	expectedPaths := []string{
		"manifest/config/config.example.yaml",
		"manifest/config/config.yaml",
		"manifest/i18n/zh-CN/plugin.json",
		"manifest/metadata.yaml",
		"manifest/resources/ignored.json",
		"manifest/resources/policy.yaml",
		"manifest/sql/001-demo.sql",
	}
	if got := manifestResourcePaths(resources); strings.Join(got, ",") != strings.Join(expectedPaths, ",") {
		t.Fatalf("expected manifest resources %#v, got %#v", expectedPaths, got)
	}
}

func TestCollectLifecycleSpecsAutoDiscoversBackendHandlers(t *testing.T) {
	pluginDir := t.TempDir()
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "controller.go"),
		lifecycleControllerSourceForTest("BeforeInstall", "AfterInstall"),
	)

	items, err := collectLifecycleSpecs(pluginDir, "plugin-dev-dynamic-lifecycle")
	if err != nil {
		t.Fatalf("expected lifecycle auto discovery to succeed, got error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 lifecycle specs, got %#v", items)
	}
	if items[0].Operation != protocol.LifecycleOperationBeforeInstall ||
		items[0].RequestType != "BeforeInstallReq" ||
		items[0].InternalPath != "/__lifecycle/before-install" {
		t.Fatalf("unexpected before-install lifecycle spec: %#v", items[0])
	}
	if items[1].Operation != protocol.LifecycleOperationAfterInstall ||
		items[1].RequestType != "AfterInstallReq" ||
		items[1].InternalPath != "/__lifecycle/after-install" {
		t.Fatalf("unexpected after-install lifecycle spec: %#v", items[1])
	}
}

func TestCollectLifecycleSpecsAppliesOverride(t *testing.T) {
	pluginDir := t.TempDir()
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "controller.go"),
		lifecycleControllerSourceForTest("BeforeInstall"),
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "lifecycle", "001-before-install.yaml"),
		"operation: BeforeInstall\nrequestType: BeforeInstallReq\ninternalPath: /before-install\ntimeoutMs: 3000\n",
	)

	items, err := collectLifecycleSpecs(pluginDir, "plugin-dev-dynamic-lifecycle")
	if err != nil {
		t.Fatalf("expected lifecycle override merge to succeed, got error: %v", err)
	}
	if len(items) != 1 ||
		items[0].RequestType != "BeforeInstallReq" ||
		items[0].InternalPath != "/before-install" ||
		items[0].TimeoutMs != 3000 {
		t.Fatalf("unexpected lifecycle override result: %#v", items)
	}
}

func TestCollectLifecycleSpecsRejectsOverrideWithoutHandler(t *testing.T) {
	pluginDir := t.TempDir()
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "lifecycle", "001-before-install.yaml"),
		"operation: BeforeInstall\nrequestType: BeforeInstallReq\ninternalPath: /__lifecycle/before-install\n",
	)

	_, err := collectLifecycleSpecs(pluginDir, "plugin-dev-dynamic-lifecycle")
	if err == nil || !strings.Contains(err.Error(), "has no matching handler") {
		t.Fatalf("expected missing handler override error, got %v", err)
	}
}

func TestCollectLifecycleSpecsRejectsDuplicateOverride(t *testing.T) {
	pluginDir := t.TempDir()
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "controller.go"),
		lifecycleControllerSourceForTest("BeforeInstall"),
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "lifecycle", "001-before-install.yaml"),
		"operation: BeforeInstall\nrequestType: BeforeInstallReq\ninternalPath: /__lifecycle/before-install\n",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "lifecycle", "002-before-install.yaml"),
		"operation: BeforeInstall\nrequestType: BeforeInstallReq\ninternalPath: /__lifecycle/before-install\n",
	)

	_, err := collectLifecycleSpecs(pluginDir, "plugin-dev-dynamic-lifecycle")
	if err == nil || !strings.Contains(err.Error(), "operation is duplicated") {
		t.Fatalf("expected duplicate override error, got %v", err)
	}
}

func TestCollectLifecycleSpecsIgnoresNonLifecycleHandlerName(t *testing.T) {
	pluginDir := t.TempDir()
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "controller.go"),
		lifecycleControllerSourceForTest("CheckInstall"),
	)

	items, err := collectLifecycleSpecs(pluginDir, "plugin-dev-dynamic-lifecycle")
	if err != nil {
		t.Fatalf("expected non-lifecycle handler name to be ignored, got %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected non-lifecycle handler name to be ignored, got %#v", items)
	}
}

func TestCollectLifecycleSpecsRejectsUnreachableOverride(t *testing.T) {
	pluginDir := t.TempDir()
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "controller.go"),
		lifecycleControllerSourceForTest("BeforeInstall"),
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "lifecycle", "001-before-install.yaml"),
		"operation: BeforeInstall\nrequestType: CustomBeforeInstallReq\ninternalPath: /custom/before-install\n",
	)

	_, err := collectLifecycleSpecs(pluginDir, "plugin-dev-dynamic-lifecycle")
	if err == nil || !strings.Contains(err.Error(), "not reachable by guest dispatcher") {
		t.Fatalf("expected unreachable override error, got %v", err)
	}
}

func TestCollectLifecycleSpecsIgnoresServiceMethods(t *testing.T) {
	pluginDir := t.TempDir()
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "internal", "service", "dynamic", "dynamic_lifecycle.go"),
		lifecycleServiceSourceForTest("BeforeInstall"),
	)

	items, err := collectLifecycleSpecs(pluginDir, "plugin-dev-dynamic-lifecycle")
	if err != nil {
		t.Fatalf("expected service method scan to be ignored without error, got %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected service lifecycle-like method to be ignored, got %#v", items)
	}
}

func TestBuildRuntimeWasmArtifactFromSourceFailsWhenEmbeddedResourcesOmitManifest(t *testing.T) {
	pluginDir := t.TempDir()

	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dev-dynamic-missing-embed\nname: Dynamic Missing Embed\nversion: v0.1.0\ntype: dynamic\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\n",
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

func TestBuildRuntimeWasmArtifactFromSourceRejectsDependencyPolicyFields(t *testing.T) {
	tests := []struct {
		name     string
		fragment string
		want     string
	}{
		{
			name:     "required",
			fragment: "required: false\n",
			want:     "dependencies.plugins[0].required",
		},
		{
			name:     "install",
			fragment: "install: auto\n",
			want:     "dependencies.plugins[0].install",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pluginDir := t.TempDir()

			mustWriteFile(
				t,
				filepath.Join(pluginDir, "plugin.yaml"),
				"id: plugin-dev-dynamic-dependency-policy\nname: Dynamic Dependency Policy\nversion: v0.1.0\ntype: dynamic\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\ndependencies:\n  plugins:\n    - id: linapro-tenant-core\n      "+tt.fragment,
			)
			mustWriteFile(
				t,
				filepath.Join(pluginDir, "frontend", "pages", "standalone.html"),
				"<!doctype html><html><body>dependency policy</body></html>",
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
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected dependency schema field error containing %q, got %v", tt.want, err)
			}
		})
	}
}

func TestBuildRuntimeWasmArtifactFromSourceSkipsHiddenEmbeddedDirectoryEntries(t *testing.T) {
	pluginDir := t.TempDir()

	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dev-dynamic-hidden\nname: Dynamic Hidden\nversion: v0.1.0\ntype: dynamic\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\n",
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
		filepath.Join(pluginDir, "manifest", "sql", "001-plugin-dev-dynamic-hidden.sql"),
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
		filepath.Join(pluginDir, "manifest", "sql", "mock-data", "001-plugin-dev-dynamic-hidden-mock-data.sql"),
		"SELECT 99;",
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "manifest", "sql", "uninstall", "001-plugin-dev-dynamic-hidden.sql"),
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
	if len(frontend) != 1 || frontend[0].Path != "frontend/pages/visible.html" {
		t.Fatalf("expected only visible embedded frontend asset, got %#v", frontend)
	}

	var installSQL []*sqlAsset
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionInstallSQL], &installSQL); err != nil {
		t.Fatalf("expected install sql section json to unmarshal, got error: %v", err)
	}
	if len(installSQL) != 1 || installSQL[0].Key != "001-plugin-dev-dynamic-hidden.sql" {
		t.Fatalf("expected only visible install sql asset, got %#v", installSQL)
	}

	var uninstallSQL []*sqlAsset
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionUninstallSQL], &uninstallSQL); err != nil {
		t.Fatalf("expected uninstall sql section json to unmarshal, got error: %v", err)
	}
	if len(uninstallSQL) != 1 || uninstallSQL[0].Key != "001-plugin-dev-dynamic-hidden.sql" {
		t.Fatalf("expected only visible uninstall sql asset, got %#v", uninstallSQL)
	}

	// Mock-data SQL ships in its own dedicated section so the host can detect
	// mock-data presence without scanning the install section, and the
	// optional mock-data load phase can pull from it independently.
	var mockSQL []*sqlAsset
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionMockSQL], &mockSQL); err != nil {
		t.Fatalf("expected mock sql section json to unmarshal, got error: %v", err)
	}
	if len(mockSQL) != 1 || mockSQL[0].Key != "001-plugin-dev-dynamic-hidden-mock-data.sql" {
		t.Fatalf("expected mock-data sql asset to land in the mock section, got %#v", mockSQL)
	}
}

func TestBuildRuntimeWasmArtifactFromSourceRejectsNonOwnedDataTables(t *testing.T) {
	cases := []struct {
		name         string
		table        string
		errorMessage string
	}{
		{
			name:         "host core table",
			table:        "sys_user",
			errorMessage: "host service data cannot declare host core table",
		},
		{
			name:         "other plugin table",
			table:        "plugin_other_plugin_record",
			errorMessage: "host service data table must belong to plugin plugin-dev-dynamic-data namespace",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pluginDir := t.TempDir()
			mustWriteFile(
				t,
				filepath.Join(pluginDir, "plugin.yaml"),
				"id: plugin-dev-dynamic-data\nname: Dynamic Data Boundary\nversion: v0.1.0\ntype: dynamic\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\nhostServices:\n  - service: data\n    methods:\n      - list\n    resources:\n      tables:\n        - "+tc.table+"\n",
			)
			mustWriteFile(
				t,
				filepath.Join(pluginDir, "main.go"),
				"package main\n\nfunc main() {}\n",
			)

			_, err := BuildRuntimeWasmArtifactFromSource(pluginDir)
			if err == nil {
				t.Fatalf("expected dynamic build to reject table %s", tc.table)
			}
			if !strings.Contains(err.Error(), tc.errorMessage) {
				t.Fatalf("expected error to contain %q, got %v", tc.errorMessage, err)
			}
		})
	}
}

func TestBuildRuntimeWasmArtifactFromSourceCleansTemporaryGoMod(t *testing.T) {
	pluginDir := t.TempDir()
	outputDir := t.TempDir()

	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dev-dynamic-temp-gomod\nname: Dynamic Temp GoMod\nversion: v0.1.0\ntype: dynamic\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\n",
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
		"id: plugin-dev-dynamic-write\nname: Dynamic Write\nversion: v0.1.0\ntype: dynamic\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\n",
	)

	repoRoot, ok := findRuntimeBuildRepoRoot(".")
	if !ok {
		t.Fatal("expected builder test to resolve repo root")
	}
	out, err := WriteRuntimeWasmArtifactFromSource(pluginDir, "")
	if err != nil {
		t.Fatalf("expected dynamic artifact write to succeed, got error: %v", err)
	}
	expectedPath := filepath.Join(repoRoot, defaultRuntimeOutputDir, "plugin-dev-dynamic-write.wasm")
	if out.ArtifactPath != expectedPath {
		t.Fatalf("expected generated dynamic artifact path %s, got %s", expectedPath, out.ArtifactPath)
	}
	t.Cleanup(func() {
		_ = os.Remove(out.ArtifactPath)
		_ = os.RemoveAll(filepath.Join(repoRoot, defaultRuntimeOutputDir, runtimeWorkspaceDirName, "plugin-dev-dynamic-write"))
	})

	content, err := os.ReadFile(out.ArtifactPath)
	if err != nil {
		t.Fatalf("expected generated dynamic artifact to exist, got error: %v", err)
	}
	if len(content) == 0 {
		t.Fatalf("expected generated dynamic artifact to contain bytes")
	}
	if _, err = os.Stat(filepath.Join(pluginDir, "temp", "plugin-dev-dynamic-write.wasm")); !os.IsNotExist(err) {
		t.Fatalf("expected generated dynamic artifact to stop being written into plugin temp/, got err=%v", err)
	}
}

func TestWriteRuntimeWasmArtifactFromSourceSupportsExternalOutputDir(t *testing.T) {
	pluginDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "output")
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dev-dynamic-output\nname: Dynamic Output\nversion: v0.1.0\ntype: dynamic\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\n",
	)

	out, err := WriteRuntimeWasmArtifactFromSource(pluginDir, outputDir)
	if err != nil {
		t.Fatalf("expected dynamic artifact write to external dir to succeed, got error: %v", err)
	}
	if expected := filepath.Join(outputDir, "plugin-dev-dynamic-output.wasm"); out.ArtifactPath != expected {
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
		"id: plugin-dev-dynamic-relative-output\nname: Dynamic Relative Output\nversion: v0.1.0\ntype: dynamic\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\n",
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
	if expected := filepath.Join(outputDir, "plugin-dev-dynamic-relative-output.wasm"); out.ArtifactPath != expected {
		t.Fatalf("expected generated dynamic artifact path %s, got %s", expected, out.ArtifactPath)
	}
	if expected := filepath.Join(outputDir, runtimeWorkspaceDirName, "plugin-dev-dynamic-relative-output", "runtime-plugin.wasm"); out.RuntimePath != expected {
		t.Fatalf("expected generated guest runtime path %s, got %s", expected, out.RuntimePath)
	}
	if _, err = os.Stat(out.ArtifactPath); err != nil {
		t.Fatalf("expected generated dynamic artifact to exist in relative output dir, got error: %v", err)
	}
	if _, err = os.Stat(out.RuntimePath); err != nil {
		t.Fatalf("expected generated guest runtime to exist in relative output dir, got error: %v", err)
	}
}

func TestSelectGuestRuntimeGoWorkUsesPluginWorkspaceOnlyForOfficialPlugins(t *testing.T) {
	repoRoot, ok := findRuntimeBuildRepoRoot(".")
	if !ok {
		t.Fatal("expected builder test to resolve repo root")
	}

	officialPluginDir := filepath.Join(repoRoot, "apps", "lina-plugins", "linapro-demo-dynamic")
	if got := selectGuestRuntimeGoWork(officialPluginDir); got != filepath.Join(repoRoot, "temp", "go.work.plugins") {
		t.Fatalf("expected official plugin dir to use temporary plugin workspace, got %q", got)
	}

	syntheticPluginDir := t.TempDir()
	if got := selectGuestRuntimeGoWork(syntheticPluginDir); got != "off" {
		t.Fatalf("expected synthetic plugin dir to use workspace off, got %q", got)
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

func lifecycleControllerSourceForTest(methodNames ...string) string {
	return lifecycleCallbackSourceForTest("backend", "Controller", "c", methodNames...)
}

func lifecycleServiceSourceForTest(methodNames ...string) string {
	return lifecycleCallbackSourceForTest("dynamic", "Service", "s", methodNames...)
}

func lifecycleCallbackSourceForTest(packageName string, receiverType string, receiverName string, methodNames ...string) string {
	var builder strings.Builder
	builder.WriteString("package ")
	builder.WriteString(packageName)
	builder.WriteString("\n\nimport \"context\"\n\ntype ")
	builder.WriteString(receiverType)
	builder.WriteString(" struct{}\n\ntype LifecycleDecisionRes struct {\n\tOK bool `json:\"ok\"`\n}\n")
	for _, methodName := range methodNames {
		builder.WriteString("\ntype ")
		builder.WriteString(methodName)
		builder.WriteString("Req struct{}\n\nfunc (")
		builder.WriteString(receiverName)
		builder.WriteString(" *")
		builder.WriteString(receiverType)
		builder.WriteString(") ")
		builder.WriteString(methodName)
		builder.WriteString("(_ context.Context, _ *")
		builder.WriteString(methodName)
		builder.WriteString("Req) (*LifecycleDecisionRes, error) {\n\treturn &LifecycleDecisionRes{OK: true}, nil\n}\n")
	}
	return builder.String()
}

func manifestResourcePaths(resources []*manifestResource) []string {
	paths := make([]string, 0, len(resources))
	for _, resource := range resources {
		if resource == nil {
			continue
		}
		paths = append(paths, resource.Path)
	}
	sort.Strings(paths)
	return paths
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
