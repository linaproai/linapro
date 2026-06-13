// This file verifies static ownership boundaries between plugin catalog, store,
// and plugintypes packages.

package plugin

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPluginInternalImportBoundaries verifies leaf and manifest packages do not
// grow forbidden sibling or generated-model imports.
func TestPluginInternalImportBoundaries(t *testing.T) {
	checkForbiddenImports(t, "internal/plugintypes", map[string]string{
		"lina-core/internal/service/plugin/internal/catalog":     "plugintypes must remain independent from catalog",
		"lina-core/internal/service/plugin/internal/store":       "plugintypes must remain independent from store",
		"lina-core/internal/service/plugin/internal/runtime":     "plugintypes must remain independent from runtime",
		"lina-core/internal/service/plugin/internal/integration": "plugintypes must remain independent from integration",
		"lina-core/internal/dao":                                 "plugintypes must not depend on generated DAO models",
		"lina-core/internal/model/do":                            "plugintypes must not depend on generated DO models",
		"lina-core/internal/model/entity":                        "plugintypes must not depend on generated entity models",
	})
	checkForbiddenImports(t, "internal/catalog", map[string]string{
		"lina-core/internal/service/plugin/internal/runtime":     "catalog must not depend on runtime implementation",
		"lina-core/internal/service/plugin/internal/integration": "catalog must not depend on integration implementation",
		"lina-core/internal/dao":                                 "catalog must not own governance persistence",
		"lina-core/internal/model/do":                            "catalog must not own generated DO writes",
		"lina-core/internal/model/entity":                        "catalog must not expose generated entity reads",
	})
}

// TestPluginStoreExportedSurfaceDoesNotLeakGeneratedModels verifies store owns
// generated model access internally without returning those types as API.
func TestPluginStoreExportedSurfaceDoesNotLeakGeneratedModels(t *testing.T) {
	files := parseGoFiles(t, "internal/store")
	for path, file := range files {
		for _, decl := range file.Decls {
			switch typed := decl.(type) {
			case *ast.GenDecl:
				if typed.Tok != token.TYPE {
					continue
				}
				for _, spec := range typed.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok || !typeSpec.Name.IsExported() {
						continue
					}
					if usesGeneratedModel(typeSpec.Type) {
						t.Fatalf("%s: exported store type %s leaks generated DAO/DO/entity model types", path, typeSpec.Name.Name)
					}
				}
			case *ast.FuncDecl:
				if !exportedFuncDecl(typed) {
					continue
				}
				if usesGeneratedModel(typed.Type) {
					t.Fatalf("%s: exported store function %s leaks generated DAO/DO/entity model types", path, typed.Name.Name)
				}
			}
		}
	}
}

// TestCatalogSetterWiringRemoved verifies catalog no longer exposes or stores
// the old runtime/integration callback wiring.
func TestCatalogSetterWiringRemoved(t *testing.T) {
	files := parseGoFiles(t, "internal/catalog")
	for path, file := range files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch typed := node.(type) {
			case *ast.FuncDecl:
				if strings.HasPrefix(typed.Name.Name, "Set") {
					t.Fatalf("%s: catalog must not define Set* wiring method/function %s", path, typed.Name.Name)
				}
			case *ast.TypeSpec:
				if typed.Name.Name != "serviceImpl" {
					return true
				}
				structType, ok := typed.Type.(*ast.StructType)
				if !ok {
					return true
				}
				for _, field := range structType.Fields.List {
					for _, name := range field.Names {
						if forbiddenCatalogCallbackField(name.Name) {
							t.Fatalf("%s: catalog serviceImpl still stores callback field %s", path, name.Name)
						}
					}
				}
			}
			return true
		})
	}

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "plugin.go", nil, 0)
	if err != nil {
		t.Fatalf("parse plugin.go: %v", err)
	}
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || !strings.HasPrefix(selector.Sel.Name, "Set") {
			return true
		}
		ident, ok := selector.X.(*ast.Ident)
		if ok && ident.Name == "catalogSvc" {
			position := fileSet.Position(selector.Pos())
			t.Fatalf("%s: plugin.New must not call catalog Set* wiring method %s", position, selector.Sel.Name)
		}
		return true
	})
}

// TestPluginRuntimeCacheOldImportRemoved verifies plugin runtime cache
// coordination is imported from the cache coordination boundary, not the old
// plugin-owned package path.
func TestPluginRuntimeCacheOldImportRemoved(t *testing.T) {
	checkForbiddenImports(t, ".", map[string]string{
		"lina-core/internal/service/plugin/runtimecache": "plugin runtime cache revision control must live under cachecoord/revisionctrl",
	})
	checkForbiddenImports(t, "../i18n", map[string]string{
		"lina-core/internal/service/plugin/runtimecache": "i18n must use cachecoord/revisionctrl instead of the plugin package tree",
	})
	checkForbiddenImports(t, "../../cmd", map[string]string{
		"lina-core/internal/service/plugin/runtimecache": "startup tests must use cachecoord/revisionctrl instead of the old plugin package tree",
	})
}

// TestPluginWiringStateStaticBoundaries verifies the B-stage wiring and
// mutable-state cleanup cannot regress silently.
func TestPluginWiringStateStaticBoundaries(t *testing.T) {
	forbiddenFiles := []string{
		filepath.Join("internal", "integration", "integration_wiring.go"),
		filepath.Join("internal", "lifecycle", "lifecycle_wiring.go"),
	}
	for _, path := range forbiddenFiles {
		if _, err := os.Stat(path); err == nil {
			t.Fatalf("%s: old production wiring file must not be restored", path)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", path, err)
		}
	}

	checkForbiddenText(t, "internal/runtime", map[string]string{
		"ValidateRequiredDependencies": "runtime dependencies must be provided by constructor parameters",
		"SetDependencyValidator":       "runtime dependency validator must be provided by constructor parameters",
		"SetMenuManager":               "runtime menu manager must be provided by constructor parameters",
		"SetHookDispatcher":            "runtime hook dispatcher must be provided by constructor parameters",
		"SetStorageCleanupServices":    "runtime storage cleanup services must be provided by constructor parameters",
	})
	checkForbiddenText(t, "internal/integration", map[string]string{
		"defaultSharedState":     "integration shared state must be injected by the composition root",
		"SetDynamicJobExecutor":  "integration dynamic job executor must be provided by constructor parameters",
		"SetOrganizationService": "integration organization capability must be provided by constructor parameters",
	})
	checkForbiddenText(t, ".", map[string]string{
		"lifecycleObserverByID":              "lifecycle observers must stay on service instances",
		"pluginRuntimeCacheObservedRevision": "runtime cache observed revision must stay in revision controllers",
	})
	checkForbiddenText(t, "internal/wasm", map[string]string{
		"atomic.Pointer": "WASM host service dependencies must be held by runtime instances",
		"func Configure": "WASM host service production Configure* entrypoints must not be restored",
	})
	checkForbiddenText(t, "internal/testutil/testutil_services.go", map[string]string{
		"SetDependencyValidator": "testutil must not replicate old runtime setter wiring",
		"SetMenuManager":         "testutil must not replicate old runtime/integration setter wiring",
		"SetHookDispatcher":      "testutil must not replicate old runtime/integration setter wiring",
		"SetReconciler":          "testutil must not replicate old lifecycle setter wiring",
	})
	checkForbiddenPluginRootWiringMethods(t)
	checkForbiddenHTTPStartupPluginSetterCalls(t)
	checkForbiddenImports(t, "../middleware", map[string]string{
		"lina-core/internal/service/plugin": "middleware must not depend on the complete plugin facade when publishing route middleware",
	})
}

// TestCapabilityHostInternalMicroPackagesRemoved verifies host capability
// domain adapters stay inside the capabilityhost package instead of re-growing
// single-file internal micro packages.
func TestCapabilityHostInternalMicroPackagesRemoved(t *testing.T) {
	internalDir := filepath.Join("internal", "capabilityhost", "internal")
	if _, err := os.Stat(internalDir); os.IsNotExist(err) {
		return
	} else if err != nil {
		t.Fatalf("stat %s: %v", internalDir, err)
	}
	if err := filepath.WalkDir(internalDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".go") {
			t.Fatalf("%s: capabilityhost internal micro package file must be merged into the capabilityhost package", path)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk %s: %v", internalDir, err)
	}
}

// TestPluginLifecycleOrchestrationStaticBoundaries captures C-stage ownership
// boundaries while lifecycle orchestration is being migrated in batches.
func TestPluginLifecycleOrchestrationStaticBoundaries(t *testing.T) {
	checkForbiddenText(t, "internal/lifecycle", map[string]string{
		"ResolveSQLAssets":       "SQL asset resolution belongs to internal/migration",
		"ResolvePluginSQLAssets": "SQL asset resolution belongs to internal/migration",
		"SysPluginMigration":     "migration ledger writes belong to internal/migration",
		"ReconcileProvider":      "lifecycle must receive the complete orchestration runtime seam, not the old dynamic-only reconciler seam",
	})
	if _, err := os.Stat(filepath.Join("internal", "migration", "migration.go")); err != nil {
		t.Fatalf("internal/migration/migration.go must own SQL migration execution: %v", err)
	}
	checkForbiddenText(t, ".", map[string]string{
		"lifecycleReconcilerProvider": "lifecycle must be constructed after runtime with explicit orchestration dependencies",
	})
	checkLifecycleConstructorWiring(t)

	transitionalRootDAO := map[string]map[string]string{
		"plugin.go": {
			"lina-core/internal/model/entity": "root facade still exposes route projection contracts that use generated menu entity until route/presentation cleanup",
		},
		"plugin_integration.go": {
			"lina-core/internal/model/entity": "integration projection facade still exposes menu entity until projection contracts are narrowed",
		},
		"plugin_runtime_delegates.go": {
			"lina-core/internal/model/entity": "runtime delegate still bridges permission menu filtering until projection contracts are narrowed",
		},
	}
	checkPluginRootGeneratedModelImports(t, transitionalRootDAO)
	checkPluginProjectionBuilderBoundary(t)
	checkPluginChangePublisherBoundary(t)
}

// TestPluginUpgradeOrchestrationStaticBoundaries captures the D-stage upgrade
// ownership boundaries now that source and dynamic upgrade orchestration live
// under the unified upgrade owner.
func TestPluginUpgradeOrchestrationStaticBoundaries(t *testing.T) {
	upgradeDir := filepath.Join("internal", "upgrade")
	sourceUpgradeDir := filepath.Join("internal", "sourceupgrade")
	runtimeUpgradeDir := filepath.Join("internal", "runtimeupgrade")

	if _, err := os.Stat(upgradeDir); err != nil {
		t.Fatalf("internal/upgrade must own plugin upgrade orchestration: %v", err)
	}
	checkPackageName(t, upgradeDir, "upgrade")

	for _, legacyDir := range []string{sourceUpgradeDir, runtimeUpgradeDir} {
		if _, err := os.Stat(legacyDir); err == nil {
			t.Fatalf("%s: old upgrade package must not be restored", legacyDir)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", legacyDir, err)
		}
	}

	checkForbiddenImports(t, ".", map[string]string{
		"lina-core/internal/service/plugin/internal/sourceupgrade":  "source upgrade orchestration must live under internal/upgrade",
		"lina-core/internal/service/plugin/internal/runtimeupgrade": "runtime upgrade planning must live under internal/upgrade",
	})
	checkForbiddenText(t, "internal/upgrade", map[string]string{
		"recordSourceUpgradeFailureMigration": "runtime upgrade failure ledger writes must use the unified owner helper",
	})
	checkRootRuntimeUpgradeDoesNotReenterSourcePublicMethod(t)
}

// checkPluginProjectionBuilderBoundary keeps management list, summary, detail,
// and dependency snapshot paths on the same projection builder.
func checkPluginProjectionBuilderBoundary(t *testing.T) {
	t.Helper()

	projectionCallers := map[string][]string{
		"plugin_list.go": {
			"syncAndList",
			"buildManagementList",
			"buildManagementSummaryList",
			"buildManagementDetail",
		},
		"plugin_dependency.go": {
			"buildDependencySnapshots",
		},
	}
	for path, functions := range projectionCallers {
		fileSet := token.NewFileSet()
		file, err := parser.ParseFile(fileSet, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, name := range functions {
			decl := findFuncDecl(file, name)
			if decl == nil {
				t.Fatalf("%s: expected function %s", path, name)
			}
			if !funcDeclCalls(decl, "buildPluginProjection") {
				t.Fatalf("%s: %s must use buildPluginProjection", path, name)
			}
			if path == "plugin_list.go" && funcDeclCalls(decl, "ScanManifests") {
				t.Fatalf("%s: %s must not own a separate manifest scan pipeline", path, name)
			}
		}
	}
}

// checkPluginChangePublisherBoundary keeps runtime revision publication,
// management read-model invalidation, and derived cache invalidation behind the
// same root plugin change publisher.
func checkPluginChangePublisherBoundary(t *testing.T) {
	t.Helper()

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "plugin_runtime_cache.go", nil, 0)
	if err != nil {
		t.Fatalf("parse plugin_runtime_cache.go: %v", err)
	}
	publisher := findFuncDecl(file, "publishPluginChange")
	if publisher == nil {
		t.Fatalf("plugin_runtime_cache.go: expected publishPluginChange")
	}
	for _, name := range []string{
		"invalidateRuntimeUpgradeCaches",
		"InvalidateManagementListCache",
		"MarkChanged",
	} {
		if !funcDeclCalls(publisher, name) {
			t.Fatalf("plugin_runtime_cache.go: publishPluginChange must call %s", name)
		}
	}
	for _, name := range []string{
		"markRuntimeCacheChanged",
		"MarkRuntimeCacheChanged",
		"PublishPluginChange",
		"syncEnabledSnapshotAndPublishRuntimeChange",
	} {
		decl := findFuncDecl(file, name)
		if decl == nil {
			t.Fatalf("plugin_runtime_cache.go: expected function %s", name)
		}
		if !funcDeclCalls(decl, "publishPluginChange") {
			t.Fatalf("plugin_runtime_cache.go: %s must delegate to publishPluginChange", name)
		}
	}

	listFile := parseSingleFile(t, "plugin_list.go")
	for _, name := range []string{"SyncSourcePlugins", "SyncSourcePluginsStrict", "SyncAndList"} {
		decl := findFuncDecl(listFile, name)
		if decl == nil {
			t.Fatalf("plugin_list.go: expected function %s", name)
		}
		if !funcDeclCalls(decl, "publishPluginChange") {
			t.Fatalf("plugin_list.go: %s must publish through publishPluginChange", name)
		}
		if funcDeclCalls(decl, "markRuntimeCacheChanged") || funcDeclCalls(decl, "Invalidate") {
			t.Fatalf("plugin_list.go: %s must not bypass publishPluginChange for cache invalidation", name)
		}
	}

	lifecycleFile := parseSingleFile(t, filepath.Join("internal", "lifecycle", "lifecycle_cache.go"))
	lifecyclePublisher := findFuncDecl(lifecycleFile, "syncEnabledSnapshotAndPublishRuntimeChange")
	if lifecyclePublisher == nil {
		t.Fatalf("internal/lifecycle/lifecycle_cache.go: expected syncEnabledSnapshotAndPublishRuntimeChange")
	}
	if !funcDeclCalls(lifecyclePublisher, "PublishPluginChange") {
		t.Fatalf("internal/lifecycle/lifecycle_cache.go: syncEnabledSnapshotAndPublishRuntimeChange must use scoped PublishPluginChange")
	}
	if funcDeclCalls(lifecyclePublisher, "MarkRuntimeCacheChanged") {
		t.Fatalf("internal/lifecycle/lifecycle_cache.go: syncEnabledSnapshotAndPublishRuntimeChange must not use legacy MarkRuntimeCacheChanged")
	}

	runtimeFile := parseSingleFile(t, filepath.Join("internal", "runtime", "runtime_wiring.go"))
	runtimePublisher := findFuncDecl(runtimeFile, "notifyRuntimeCacheChanged")
	if runtimePublisher == nil {
		t.Fatalf("internal/runtime/runtime_wiring.go: expected notifyRuntimeCacheChanged")
	}
	if !funcDeclCalls(runtimePublisher, "PublishPluginChange") {
		t.Fatalf("internal/runtime/runtime_wiring.go: notifyRuntimeCacheChanged must use scoped PublishPluginChange")
	}
	if funcDeclCalls(runtimePublisher, "MarkRuntimeCacheChanged") {
		t.Fatalf("internal/runtime/runtime_wiring.go: notifyRuntimeCacheChanged must not use legacy MarkRuntimeCacheChanged")
	}
}

// checkLifecycleConstructorWiring ensures C-stage lifecycle orchestration gets
// explicit owner dependencies through its constructor instead of a dynamic-only
// post-construction bridge.
func checkLifecycleConstructorWiring(t *testing.T) {
	t.Helper()

	fileSet := token.NewFileSet()
	lifecycleFile, err := parser.ParseFile(fileSet, filepath.Join("internal", "lifecycle", "lifecycle.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse lifecycle.go: %v", err)
	}
	newArgCount := -1
	for _, decl := range lifecycleFile.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "New" || fn.Recv != nil {
			continue
		}
		newArgCount = fn.Type.Params.NumFields()
	}
	if newArgCount != 11 {
		t.Fatalf("internal/lifecycle.New must explicitly receive 11 orchestration dependencies, got %d", newArgCount)
	}

	pluginFile, err := parser.ParseFile(fileSet, "plugin.go", nil, 0)
	if err != nil {
		t.Fatalf("parse plugin.go: %v", err)
	}
	callArgCount := -1
	ast.Inspect(pluginFile, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "New" {
			return true
		}
		ident, ok := selector.X.(*ast.Ident)
		if !ok || ident.Name != "lifecycle" {
			return true
		}
		callArgCount = len(call.Args)
		return true
	})
	if callArgCount != 11 {
		t.Fatalf("plugin.New must construct lifecycle with 11 explicit dependencies, got %d", callArgCount)
	}
}

// findFuncDecl returns the named package-level function declaration.
func findFuncDecl(file *ast.File, name string) *ast.FuncDecl {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && fn.Name.Name == name {
			return fn
		}
	}
	return nil
}

// parseSingleFile parses one Go source file from the plugin package tree.
func parseSingleFile(t *testing.T, path string) *ast.File {
	t.Helper()
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return file
}

// funcDeclCalls reports whether function body calls a selector or function name.
func funcDeclCalls(decl *ast.FuncDecl, name string) bool {
	if decl == nil || decl.Body == nil {
		return false
	}
	found := false
	ast.Inspect(decl.Body, func(node ast.Node) bool {
		if found {
			return false
		}
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fun := call.Fun.(type) {
		case *ast.Ident:
			found = fun.Name == name
		case *ast.SelectorExpr:
			found = fun.Sel.Name == name
		}
		return !found
	})
	return found
}

func checkForbiddenImports(t *testing.T, dir string, forbidden map[string]string) {
	t.Helper()
	for path, file := range parseGoFiles(t, dir) {
		for _, importSpec := range file.Imports {
			importPath := strings.Trim(importSpec.Path.Value, `"`)
			if reason, ok := forbidden[importPath]; ok {
				t.Fatalf("%s imports %s: %s", path, importPath, reason)
			}
		}
	}
}

// checkPackageName verifies all production Go files in dir use the expected package.
func checkPackageName(t *testing.T, dir string, expected string) {
	t.Helper()
	for path, file := range parseGoFiles(t, dir) {
		if file.Name == nil || file.Name.Name != expected {
			t.Fatalf("%s: expected package %s, got %v", path, expected, file.Name)
		}
	}
}

// pathExists reports whether path exists.
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// checkRootRuntimeUpgradeDoesNotReenterSourcePublicMethod blocks the old source
// branch where runtime upgrade execution called the public source upgrade entry.
func checkRootRuntimeUpgradeDoesNotReenterSourcePublicMethod(t *testing.T) {
	t.Helper()
	file := parseSingleFile(t, "plugin_runtime_upgrade_execute.go")
	decl := findFuncDecl(file, "executeRuntimeUpgradeByType")
	if decl == nil {
		return
	}
	if funcDeclCalls(decl, "UpgradeSourcePlugin") {
		t.Fatalf("plugin_runtime_upgrade_execute.go: executeRuntimeUpgradeByType must call the upgrade owner directly, not public UpgradeSourcePlugin")
	}
}

// checkPluginRootGeneratedModelImports blocks new root-facade DAO/DO/Entity
// imports while documenting the C-stage files that are intentionally migrated
// later in this change.
func checkPluginRootGeneratedModelImports(t *testing.T, allowlist map[string]map[string]string) {
	t.Helper()
	for path, file := range parseGoFiles(t, ".") {
		if strings.HasPrefix(path, "internal"+string(filepath.Separator)) ||
			strings.Contains(path, string(filepath.Separator)+"internal"+string(filepath.Separator)) {
			continue
		}
		base := filepath.Base(path)
		for _, importSpec := range file.Imports {
			importPath := strings.Trim(importSpec.Path.Value, `"`)
			if importPath != "lina-core/internal/dao" &&
				importPath != "lina-core/internal/model/do" &&
				importPath != "lina-core/internal/model/entity" {
				continue
			}
			if allowedForFile, ok := allowlist[base]; ok {
				if _, allowed := allowedForFile[importPath]; allowed {
					continue
				}
			}
			t.Fatalf("%s imports %s: root plugin facade must not gain new generated model access during C migration", path, importPath)
		}
	}
}

// checkForbiddenPluginRootWiringMethods verifies the root plugin facade does
// not reintroduce startup-only wiring methods.
func checkForbiddenPluginRootWiringMethods(t *testing.T) {
	t.Helper()
	forbidden := map[string]string{
		"SetCapabilities":                       "capability services must be passed into plugin.New",
		"SetOrganizationCapability":             "organization capability must be passed into plugin.New",
		"SetTenantStartupCapability":            "tenant startup capability must be passed into plugin.New",
		"SetTenantProvisioningCapability":       "tenant provisioning capability must be passed into plugin.New",
		"SetTenantPlatformGovernanceCapability": "tenant governance capability must be passed into plugin.New",
		"ConfigureWasmHostServices":             "WASM host service runtime must be constructed during plugin.New",
	}
	for path, file := range parseGoFiles(t, ".") {
		ast.Inspect(file, func(node ast.Node) bool {
			switch typed := node.(type) {
			case *ast.FuncDecl:
				if reason, ok := forbidden[typed.Name.Name]; ok {
					t.Fatalf("%s: root plugin production method %s is forbidden: %s", path, typed.Name.Name, reason)
				}
			case *ast.TypeSpec:
				interfaceType, ok := typed.Type.(*ast.InterfaceType)
				if !ok || typed.Name.Name != "SourceIntegrationService" && typed.Name.Name != "SourceUpgradeGovernanceService" && typed.Name.Name != "Service" {
					return true
				}
				for _, method := range interfaceType.Methods.List {
					for _, name := range method.Names {
						if reason, ok := forbidden[name.Name]; ok {
							t.Fatalf("%s: plugin interface %s exposes forbidden method %s: %s", path, typed.Name.Name, name.Name, reason)
						}
					}
				}
			}
			return true
		})
	}
}

// checkForbiddenHTTPStartupPluginSetterCalls verifies the root composition
// passes startup dependencies through constructors instead of post-construction
// plugin service setters.
func checkForbiddenHTTPStartupPluginSetterCalls(t *testing.T) {
	t.Helper()
	forbidden := map[string]string{
		"SetCapabilities":                       "capability services must be passed into plugin.New",
		"SetOrganizationCapability":             "organization capability must be passed into plugin.New",
		"SetTenantStartupCapability":            "tenant startup capability must be passed into plugin.New",
		"SetTenantProvisioningCapability":       "tenant provisioning capability must be passed into plugin.New",
		"SetTenantPlatformGovernanceCapability": "tenant governance capability must be passed into plugin.New",
		"ConfigureWasmHostServices":             "WASM host service runtime must be constructed during plugin.New",
	}
	fileSet := token.NewFileSet()
	path := filepath.Join("..", "..", "cmd", "internal", "httpstartup", "http_runtime.go")
	file, err := parser.ParseFile(fileSet, path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		reason, forbiddenCall := forbidden[selector.Sel.Name]
		if !forbiddenCall {
			return true
		}
		position := fileSet.Position(selector.Pos())
		t.Fatalf("%s: http startup must not call plugin service wiring method %s: %s", position, selector.Sel.Name, reason)
		return true
	})
}

// checkForbiddenText scans Go source files under path for forbidden text.
func checkForbiddenText(t *testing.T, path string, forbidden map[string]string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if !info.IsDir() {
		checkForbiddenTextFile(t, path, forbidden)
		return
	}
	if err = filepath.WalkDir(path, func(filePath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if strings.HasSuffix(filePath, ".go") && !strings.HasSuffix(filePath, "_test.go") {
			checkForbiddenTextFile(t, filePath, forbidden)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk %s: %v", path, err)
	}
}

// checkForbiddenTextFile scans one file for forbidden text.
func checkForbiddenTextFile(t *testing.T, path string, forbidden map[string]string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	text := string(content)
	for pattern, reason := range forbidden {
		if strings.Contains(text, pattern) {
			t.Fatalf("%s contains %q: %s", path, pattern, reason)
		}
	}
}

// parseGoFiles parses all non-test Go source files under dir.
func parseGoFiles(t *testing.T, dir string) map[string]*ast.File {
	t.Helper()
	files := make(map[string]*ast.File)
	fileSet := token.NewFileSet()
	if err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		files[path] = file
		return nil
	}); err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	if len(files) == 0 {
		t.Fatalf("expected Go files under %s", dir)
	}
	return files
}

// exportedFuncDecl reports whether decl contributes to the package API surface.
func exportedFuncDecl(decl *ast.FuncDecl) bool {
	if decl.Recv == nil {
		return decl.Name.IsExported()
	}
	if len(decl.Recv.List) == 0 {
		return false
	}
	return decl.Name.IsExported() && exportedReceiverType(decl.Recv.List[0].Type)
}

// exportedReceiverType reports whether a method receiver is externally named.
func exportedReceiverType(expr ast.Expr) bool {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.IsExported()
	case *ast.StarExpr:
		return exportedReceiverType(typed.X)
	default:
		return false
	}
}

// usesGeneratedModel reports whether node references generated DAO/DO/Entity selectors.
func usesGeneratedModel(node ast.Node) bool {
	found := false
	ast.Inspect(node, func(child ast.Node) bool {
		if found {
			return false
		}
		selector, ok := child.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := selector.X.(*ast.Ident)
		if !ok {
			return true
		}
		switch ident.Name {
		case "dao", "do", "entity":
			found = true
			return false
		default:
			return true
		}
	})
	return found
}

// forbiddenCatalogCallbackField reports whether name is an old catalog callback slot.
func forbiddenCatalogCallbackField(name string) bool {
	switch name {
	case "backendLoader", "artifactParser", "dynamicManifestLoader",
		"nodeStateSyncer", "menuSyncer", "resourceRefSyncer",
		"releaseStateSyncer", "hookDispatcher":
		return true
	default:
		return strings.HasSuffix(name, "Syncer") || strings.HasSuffix(name, "Dispatcher")
	}
}
