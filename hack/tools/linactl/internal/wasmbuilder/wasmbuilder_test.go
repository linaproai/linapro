// This file verifies dynamic WASM builder resource packaging contracts.

package wasmbuilder

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"lina-core/pkg/plugin/pluginbridge/protocol"
)

func TestRenderWasmDispatcherBoolParserUsesStandardFormsOnly(t *testing.T) {
	content, err := renderWasmDispatcher(&wasmDispatcherSpec{
		PluginID: "plugin-dev-bool",
		APIControllers: []*wasmAPIControllerSpec{
			{
				APIPackage:     "dynamic/v1",
				ImportAlias:    "controller1",
				PackagePath:    "plugin-dev-bool/backend/internal/controller/dynamic/v1",
				InterfacePath:  "plugin-dev-bool/backend/api/dynamic/v1",
				Constructor:    "controller1.New()",
				ConcreteType:   "*controller1.Controller",
				InterfaceName:  "IReview",
				InterfaceAlias: "dynamicv1",
			},
		},
		Routes: []*wasmRouteHandlerSpec{
			{
				RequestType:     "ReviewReq",
				Method:          "GET",
				Path:            "/review",
				APIPackage:      "dynamic/v1",
				ControllerAlias: "controller1",
				MethodName:      "Review",
				DTOImportAlias:  "dtoV1",
				RequestTypeExpr: "dtoV1.ReviewReq",
				Fields: []*wasmDTOFieldSpec{
					{GoName: "Enabled", JSONName: "enabled", GoType: "bool", Required: true},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("renderWasmDispatcher returned error: %v", err)
	}

	assertGeneratedWasmBoolParserUsesStandardFormsOnly(t, content)
}

func assertGeneratedWasmBoolParserUsesStandardFormsOnly(t *testing.T, content []byte) {
	t.Helper()

	fileNode, err := parser.ParseFile(token.NewFileSet(), generatedDispatcherFileName, content, 0)
	if err != nil {
		t.Fatalf("parse generated dispatcher: %v", err)
	}

	allowedCases := map[string]struct{}{
		"1":     {},
		"true":  {},
		"0":     {},
		"false": {},
		"":      {},
	}
	seenCases := make(map[string]struct{}, len(allowedCases))
	foundParser := false
	ast.Inspect(fileNode, func(node ast.Node) bool {
		switch item := node.(type) {
		case *ast.FuncDecl:
			if item.Name == nil || item.Name.Name != "generatedWasmParseBool" {
				return true
			}
			foundParser = true
			params := item.Type.Params
			if params == nil || len(params.List) != 1 || len(params.List[0].Names) != 1 {
				t.Fatalf("generated bool parser must accept exactly one value parameter")
			}
		case *ast.CallExpr:
			switch fn := item.Fun.(type) {
			case *ast.SelectorExpr:
				if ident, ok := fn.X.(*ast.Ident); ok && ident.Name == "strconv" && fn.Sel.Name == "ParseBool" {
					t.Fatalf("generated dispatcher must not call the standard library bool parser")
				}
			case *ast.Ident:
				if fn.Name == "generatedWasmParseBool" && len(item.Args) != 1 {
					t.Fatalf("generated bool parser call must pass exactly one argument")
				}
			}
		case *ast.CaseClause:
			if len(item.List) == 0 {
				return true
			}
			for _, expr := range item.List {
				literal, ok := expr.(*ast.BasicLit)
				if !ok || literal.Kind != token.STRING {
					continue
				}
				value, unquoteErr := strconv.Unquote(literal.Value)
				if unquoteErr != nil {
					t.Fatalf("unquote generated bool case %s: %v", literal.Value, unquoteErr)
				}
				if _, ok = allowedCases[value]; ok {
					seenCases[value] = struct{}{}
					continue
				}
				if value == "POST" || value == "PUT" || value == "PATCH" || value == "GET" || value == "ReviewReq" {
					continue
				}
				t.Fatalf("generated dispatcher contains unexpected string switch case %q", value)
			}
		}
		return true
	})
	if !foundParser {
		t.Fatalf("generated dispatcher must define generatedWasmParseBool")
	}
	for value := range allowedCases {
		if _, ok := seenCases[value]; !ok {
			t.Fatalf("generated bool parser missing standard case %q", value)
		}
	}
}

func TestBuildRuntimeWasmArtifactFromSourceEmbedsDeclaredAssets(t *testing.T) {
	pluginDir := t.TempDir()

	mustWriteFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: plugin-dev-dynamic-builder\nname: Dynamic Builder\nversion: v0.1.0\ntype: dynamic\nscope_nature: tenant_aware\nsupports_multi_tenant: true\ndefault_install_mode: tenant_scoped\ndescription: standalone builder test\ndependencies:\n  framework:\n    version: \">=0.1.0 <1.0.0\"\n  plugins:\n    - id: linapro-ai-core\n      version: \">=0.1.0\"\n    - id: linapro-tenant-core\n      version: \">=0.1.0\"\nhostServices:\n  - service: runtime\n    methods:\n      - log.write\n      - state.get\n      - state.set\n  - service: ai\n    owner: linapro-ai-core\n    version: v1\n    methods:\n      - text.method_status.get\n",
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
		filepath.Join(pluginDir, "hack", "config.yaml"),
		`wasm:
  hooks:
    - event: auth.login.succeeded
      action: sleep
      timeout: 50ms
      sleep: 10ms
  lifecycle:
    timeouts:
      BeforeInstall: 3s
  resources:
    - key: records
      type: table-list
      table: plugin_runtime_records
      fields:
        - name: id
          column: id
        - name: status
          column: status
      filters:
        - param: status
          column: status
          operator: eq
      orderBy:
        column: id
        direction: asc
      operations:
        - query
        - get
        - update
      keyField: id
      writableFields:
        - status
      access: both
      dataScope:
        userColumn: owner_user_id
`,
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
		"package backend\n\nimport bridgeplugin \"lina-core/pkg/plugin/pluginbridge\"\n\nfunc RegisterPlugin(plugin bridgeplugin.Declarations) error {\n\treturn plugin.Routes().Group(\"/api/v1\", \"dynamic/v1\")\n}\n",
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
	if len(manifest.Dependencies.Plugins) != 2 {
		t.Fatalf("expected two embedded plugin dependencies, got %#v", manifest.Dependencies.Plugins)
	}
	pluginDependencyVersions := make(map[string]string, len(manifest.Dependencies.Plugins))
	for _, dependency := range manifest.Dependencies.Plugins {
		pluginDependencyVersions[dependency.ID] = dependency.Version
	}
	if pluginDependencyVersions["linapro-ai-core"] != ">=0.1.0" ||
		pluginDependencyVersions["linapro-tenant-core"] != ">=0.1.0" {
		t.Fatalf("unexpected embedded plugin dependencies: %#v", manifest.Dependencies.Plugins)
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
	if len(hooks) != 1 ||
		hooks[0].Action != hookActionSleep ||
		hooks[0].TimeoutMs != 50 ||
		hooks[0].SleepMs != 10 {
		t.Fatalf("unexpected embedded hook specs: %#v", hooks)
	}

	var lifecycle []*protocol.LifecycleContract
	if err = json.Unmarshal(sections[pluginDynamicWasmSectionBackendLifecycle], &lifecycle); err != nil {
		t.Fatalf("expected lifecycle section json to unmarshal, got error: %v", err)
	}
	if len(lifecycle) != 1 ||
		lifecycle[0].Operation != protocol.LifecycleOperationBeforeInstall ||
		lifecycle[0].RequestType != "BeforeInstallReq" ||
		lifecycle[0].InternalPath != "/__lifecycle/before-install" ||
		lifecycle[0].TimeoutMs != 3000 {
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
	if len(hostServices) != 2 {
		t.Fatalf("unexpected embedded host services: %#v", hostServices)
	}
	var ownerAIService *protocol.HostServiceSpec
	for _, item := range hostServices {
		if item.Owner == "linapro-ai-core" && item.Service == "ai" {
			ownerAIService = item
			break
		}
	}
	if ownerAIService == nil || ownerAIService.Version != "v1" ||
		len(ownerAIService.Methods) != 1 ||
		ownerAIService.Methods[0] != "text.method_status.get" {
		t.Fatalf("expected embedded owner-aware AI host service, got %#v", hostServices)
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

func TestCollectHookSpecsReadsPluginHackConfig(t *testing.T) {
	pluginDir := t.TempDir()
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "hack", "config.yaml"),
		`wasm:
  hooks:
    - event: auth.login.succeeded
      action: sleep
      mode: async
      timeout: 50ms
      sleep: 10ms
`,
	)

	items, err := collectHookSpecs(pluginDir, "plugin-dev-dynamic-hooks")
	if err != nil {
		t.Fatalf("expected hook config collection to succeed, got error: %v", err)
	}
	if len(items) != 1 ||
		items[0].Event != extensionPointAuthLoginSucceeded ||
		items[0].Action != hookActionSleep ||
		items[0].Mode != callbackExecutionModeAsync ||
		items[0].TimeoutMs != 50 ||
		items[0].SleepMs != 10 {
		t.Fatalf("unexpected hook config result: %#v", items)
	}
}

func TestCollectHookSpecsRejectsConfigDurationWithoutUnit(t *testing.T) {
	pluginDir := t.TempDir()
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "hack", "config.yaml"),
		`wasm:
  hooks:
    - event: auth.login.succeeded
      action: sleep
      timeout: "50"
      sleep: 10ms
`,
	)

	_, err := collectHookSpecs(pluginDir, "plugin-dev-dynamic-hooks")
	if err == nil || !strings.Contains(err.Error(), "duration must include a valid unit") {
		t.Fatalf("expected invalid hook duration error, got %v", err)
	}
}

func TestCollectHookSpecsRejectsUnsupportedConfigField(t *testing.T) {
	pluginDir := t.TempDir()
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "hack", "config.yaml"),
		`wasm:
  hooks:
    - event: auth.login.succeeded
      action: sleep
      timeoutMs: 50
      sleep: 10ms
`,
	)

	_, err := collectHookSpecs(pluginDir, "plugin-dev-dynamic-hooks")
	if err == nil || !strings.Contains(err.Error(), "plugin hook config field is not supported: timeoutMs") {
		t.Fatalf("expected unsupported hook config field error, got %v", err)
	}
}

func TestCollectResourceSpecsReadsPluginHackConfig(t *testing.T) {
	pluginDir := t.TempDir()
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "hack", "config.yaml"),
		`wasm:
  resources:
    - key: records
      type: table-list
      table: plugin_dev_dynamic_records
      fields:
        - name: id
          column: id
        - name: status
          column: status
      filters:
        - param: status
          column: status
          operator: eq
      orderBy:
        column: id
        direction: desc
      operations:
        - query
        - get
        - update
      keyField: id
      writableFields:
        - status
      access: both
      dataScope:
        userColumn: owner_user_id
`,
	)

	items, err := collectResourceSpecs(pluginDir, "plugin-dev-dynamic-resources")
	if err != nil {
		t.Fatalf("expected resource config collection to succeed, got error: %v", err)
	}
	if len(items) != 1 ||
		items[0].Key != "records" ||
		items[0].OrderBy.Direction != "desc" ||
		items[0].Access != "both" ||
		items[0].DataScope == nil ||
		items[0].DataScope.UserColumn != "owner_user_id" {
		t.Fatalf("unexpected resource config result: %#v", items)
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

func TestCollectLifecycleSpecsReadsPluginHackConfigTimeout(t *testing.T) {
	pluginDir := t.TempDir()
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "controller.go"),
		lifecycleControllerSourceForTest("BeforeInstall"),
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "hack", "config.yaml"),
		"wasm:\n  lifecycle:\n    timeouts:\n      BeforeInstall: 2m\n",
	)

	items, err := collectLifecycleSpecs(pluginDir, "plugin-dev-dynamic-lifecycle")
	if err != nil {
		t.Fatalf("expected lifecycle config timeout discovery to succeed, got error: %v", err)
	}
	if len(items) != 1 ||
		items[0].Operation != protocol.LifecycleOperationBeforeInstall ||
		items[0].TimeoutMs != 120000 {
		t.Fatalf("unexpected lifecycle config timeout result: %#v", items)
	}
}

func TestCollectLifecycleSpecsRejectsConfigTimeoutWithoutHandler(t *testing.T) {
	pluginDir := t.TempDir()
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "hack", "config.yaml"),
		"wasm:\n  lifecycle:\n    timeouts:\n      BeforeInstall: 2m\n",
	)

	_, err := collectLifecycleSpecs(pluginDir, "plugin-dev-dynamic-lifecycle")
	if err == nil || !strings.Contains(err.Error(), "has no matching handler") {
		t.Fatalf("expected missing handler config timeout error, got %v", err)
	}
}

func TestCollectLifecycleSpecsRejectsUnsupportedConfigTimeoutOperation(t *testing.T) {
	pluginDir := t.TempDir()
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "controller.go"),
		lifecycleControllerSourceForTest("BeforeInstall"),
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "hack", "config.yaml"),
		"wasm:\n  lifecycle:\n    timeouts:\n      CheckInstall: 2m\n",
	)

	_, err := collectLifecycleSpecs(pluginDir, "plugin-dev-dynamic-lifecycle")
	if err == nil || !strings.Contains(err.Error(), "operation is unsupported") {
		t.Fatalf("expected unsupported lifecycle config timeout operation error, got %v", err)
	}
}

func TestCollectLifecycleSpecsRejectsInvalidConfigTimeoutDuration(t *testing.T) {
	pluginDir := t.TempDir()
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "backend", "controller.go"),
		lifecycleControllerSourceForTest("BeforeInstall"),
	)
	mustWriteFile(
		t,
		filepath.Join(pluginDir, "hack", "config.yaml"),
		"wasm:\n  lifecycle:\n    timeouts:\n      BeforeInstall: \"120000\"\n",
	)

	_, err := collectLifecycleSpecs(pluginDir, "plugin-dev-dynamic-lifecycle")
	if err == nil || !strings.Contains(err.Error(), "timeout duration must include a valid unit") {
		t.Fatalf("expected invalid lifecycle config timeout duration error, got %v", err)
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
