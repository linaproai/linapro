// This file verifies backend production panic usage against the documented
// governance boundary for startup, registration, Must helper, and rethrow paths.

package cmd

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"lina-core/pkg/testsupport"
)

// panicCategory names the approved semantic boundary for a production panic.
type panicCategory string

const (
	// panicCategoryMustConstructor allows explicit Must helper fail-fast behavior.
	panicCategoryMustConstructor panicCategory = "must-constructor"
	// panicCategoryPanicRethrow allows rethrowing unknown panics after normalization.
	panicCategoryPanicRethrow panicCategory = "panic-rethrow"
	// panicCategoryPluginRegistration allows top-level source plugin registration fail-fast after error-returning APIs.
	panicCategoryPluginRegistration panicCategory = "plugin-registration"
	// panicCategoryStaticConfig allows invalid static configuration to fail during startup.
	panicCategoryStaticConfig panicCategory = "static-config"
	// panicCategoryStartup allows unrecoverable process bootstrap failures.
	panicCategoryStartup panicCategory = "startup"
)

// panicAuditPolicy describes where production panic governance should scan and
// which call boundaries are approved.
type panicAuditPolicy struct {
	ScanRoots  []string
	SkipDirs   []string
	Allowances []panicAllowance
}

// panicAllowance describes one approved panic boundary in production code.
type panicAllowance struct {
	Path     string
	Function string
	Count    int
	Category panicCategory
	Reason   string
}

// panicKey identifies panic calls by source file and enclosing function.
type panicKey struct {
	Path     string
	Function string
}

// productionPanicPolicy enumerates approved panic use in backend production
// files. Counts are intentionally strict so adding another panic to an
// already-approved function still requires updating this review point.
var productionPanicPolicy = panicAuditPolicy{
	ScanRoots: []string{
		"apps/lina-core",
		"apps/lina-plugins",
	},
	SkipDirs: []string{
		"node_modules",
		"testdata",
		"testutil",
		"vendor",
	},
	Allowances: []panicAllowance{
		{
			Path:     "apps/lina-core/main.go",
			Function: "main",
			Count:    1,
			Category: panicCategoryStartup,
			Reason:   "the command tree cannot be constructed, so the process cannot continue",
		},
		{
			Path:     "apps/lina-core/pkg/bizerr/bizerr_code.go",
			Function: "MustDefineWithKey",
			Count:    3,
			Category: panicCategoryMustConstructor,
			Reason:   "invalid business error definitions must fail during startup or tests",
		},
		{
			Path:     "apps/lina-core/internal/service/config/config_duration.go",
			Function: "mustScanConfig",
			Count:    2,
			Category: panicCategoryStaticConfig,
			Reason:   "invalid static configuration must fail before the dependent component runs",
		},
		{
			Path:     "apps/lina-core/internal/service/config/config_duration.go",
			Function: "mustParsePositiveDuration",
			Count:    2,
			Category: panicCategoryStaticConfig,
			Reason:   "invalid static duration configuration has no safe runtime meaning",
		},
		{
			Path:     "apps/lina-core/internal/service/config/config_duration.go",
			Function: "mustValidateSecondAlignedDuration",
			Count:    2,
			Category: panicCategoryStaticConfig,
			Reason:   "static scheduler intervals must be valid before cron registration",
		},
		{
			Path:     "apps/lina-core/internal/service/config/config_cluster.go",
			Function: "mustValidateClusterConfig",
			Count:    3,
			Category: panicCategoryStaticConfig,
			Reason:   "cluster mode must fail fast when the required Redis coordination backend is missing or unsupported",
		},
		{
			Path:     "apps/lina-core/internal/service/config/config_i18n.go",
			Function: "normalizeAndValidateI18nConfig",
			Count:    2,
			Category: panicCategoryStaticConfig,
			Reason:   "missing packaged i18n defaults makes locale resolution undefined",
		},
		{
			Path:     "apps/lina-core/internal/service/config/config_metadata.go",
			Function: "(*serviceImpl).GetMetadata",
			Count:    2,
			Category: panicCategoryStaticConfig,
			Reason:   "packaged delivery metadata must be readable and parseable",
		},
		{
			Path:     "apps/lina-core/internal/service/config/config_metadata.go",
			Function: "mustScanMetadataConfig",
			Count:    3,
			Category: panicCategoryStaticConfig,
			Reason:   "embedded metadata scan failures indicate a broken build artifact",
		},
		{
			Path:     "apps/lina-core/internal/service/config/config_plugin.go",
			Function: "(*serviceImpl).GetPlugin",
			Count:    2,
			Category: panicCategoryStaticConfig,
			Reason:   "static plugin.autoEnable validation surfaces from helpers as errors and is converted to a single fail-fast panic at the cache-load boundary so startup terminates with a clear message before dependent components run",
		},
		{
			Path:     "apps/lina-core/internal/service/config/config_plugin.go",
			Function: "SetPluginAutoEnableOverride",
			Count:    1,
			Category: panicCategoryStaticConfig,
			Reason:   "test override helpers receive already-validated IDs; a normalization failure indicates broken test fixtures and must surface immediately",
		},
		{
			Path:     "apps/lina-core/internal/service/config/config_plugin.go",
			Function: "SetPluginAutoEnableEntriesOverride",
			Count:    1,
			Category: panicCategoryStaticConfig,
			Reason:   "test override helpers receive already-validated entries; a normalization failure indicates broken test fixtures and must surface immediately",
		},
		{
			Path:     "apps/lina-core/internal/service/config/config_runtime_params_revision.go",
			Function: "configureRuntimeParamCacheDomain",
			Count:    1,
			Category: panicCategoryStaticConfig,
			Reason:   "runtime-config cachecoord domain registration is a static consistency contract and failures make protected config freshness undefined",
		},
		{
			Path:     "apps/lina-core/internal/service/pluginruntimecache/pluginruntimecache_controller.go",
			Function: "configureRuntimeCacheDomain",
			Count:    1,
			Category: panicCategoryStaticConfig,
			Reason:   "plugin-runtime cachecoord domain registration is a static consistency contract and failures make plugin cache freshness undefined",
		},
		{
			Path:     "apps/lina-core/internal/service/role/role_access_revision.go",
			Function: "configureAccessTopologyCacheDomain",
			Count:    1,
			Category: panicCategoryStaticConfig,
			Reason:   "permission-access cachecoord domain registration is a static consistency contract and failures must fail closed before serving authorization checks",
		},
		{
			Path:     "apps/lina-core/internal/service/middleware/middleware_request_body_limit.go",
			Function: "(*serviceImpl).RequestBodyLimit",
			Count:    1,
			Category: panicCategoryPanicRethrow,
			Reason:   "unknown framework panic is rethrown after known request-size errors are normalized",
		},
		{
			Path:     "apps/lina-core/pkg/pluginbridge/guest/guest_router.go",
			Function: "MustNewGuestControllerRouteDispatcher",
			Count:    1,
			Category: panicCategoryMustConstructor,
			Reason:   "Must constructor documents fail-fast behavior and has a non-Must alternative",
		},
		{
			Path:     "apps/lina-core/pkg/pluginbridge/hostservice/hostservice_validation.go",
			Function: "MustNormalizeHostServiceSpecs",
			Count:    1,
			Category: panicCategoryMustConstructor,
			Reason:   "Must helper is reserved for compile-time host service declarations",
		},
		{
			Path:     "apps/lina-core/pkg/plugindb/host/db.go",
			Function: "registerPluginDataDrivers",
			Count:    1,
			Category: panicCategoryStartup,
			Reason:   "plugin data DB drivers must register once before plugin data access can work",
		},
		{
			Path:     "apps/lina-plugins/content-notice/backend/plugin.go",
			Function: "init",
			Count:    2,
			Category: panicCategoryPluginRegistration,
			Reason:   "top-level source plugin registration chooses fail-fast after the error-returning registration API rejects invalid static declarations",
		},
		{
			Path:     "apps/lina-plugins/demo-control/backend/plugin.go",
			Function: "init",
			Count:    2,
			Category: panicCategoryPluginRegistration,
			Reason:   "top-level source plugin registration chooses fail-fast after the error-returning registration API rejects invalid static declarations",
		},
		{
			Path:     "apps/lina-plugins/monitor-loginlog/backend/plugin.go",
			Function: "init",
			Count:    5,
			Category: panicCategoryPluginRegistration,
			Reason:   "top-level source plugin registration chooses fail-fast after the error-returning registration API rejects invalid static declarations",
		},
		{
			Path:     "apps/lina-plugins/monitor-online/backend/plugin.go",
			Function: "init",
			Count:    2,
			Category: panicCategoryPluginRegistration,
			Reason:   "top-level source plugin registration chooses fail-fast after the error-returning registration API rejects invalid static declarations",
		},
		{
			Path:     "apps/lina-plugins/monitor-operlog/backend/plugin.go",
			Function: "init",
			Count:    2,
			Category: panicCategoryPluginRegistration,
			Reason:   "top-level source plugin registration chooses fail-fast after the error-returning registration API rejects invalid static declarations",
		},
		{
			Path:     "apps/lina-plugins/monitor-server/backend/plugin.go",
			Function: "init",
			Count:    4,
			Category: panicCategoryPluginRegistration,
			Reason:   "top-level source plugin registration chooses fail-fast after the error-returning registration API rejects invalid static declarations",
		},
		{
			Path:     "apps/lina-plugins/multi-tenant/backend/plugin.go",
			Function: "init",
			Count:    5,
			Category: panicCategoryPluginRegistration,
			Reason:   "top-level source plugin registration chooses fail-fast after the error-returning registration API rejects invalid static declarations",
		},
		{
			Path:     "apps/lina-plugins/org-center/backend/plugin.go",
			Function: "init",
			Count:    2,
			Category: panicCategoryPluginRegistration,
			Reason:   "top-level source plugin registration chooses fail-fast after the error-returning registration API rejects invalid static declarations",
		},
		{
			Path:     "apps/lina-plugins/plugin-demo-source/backend/plugin.go",
			Function: "init",
			Count:    5,
			Category: panicCategoryPluginRegistration,
			Reason:   "top-level source plugin registration chooses fail-fast after the error-returning registration API rejects invalid static declarations",
		},
	},
}

// TestProductionPanicsMatchAllowlist verifies production panic usage stays
// narrow and documented.
func TestProductionPanicsMatchAllowlist(t *testing.T) {
	repoRoot := repoRootFromTest(t)
	if !testsupport.OfficialPluginsWorkspaceReady(repoRoot) {
		t.Skip("official plugin workspace is not initialized")
	}
	found := scanProductionPanicCalls(t, repoRoot, productionPanicPolicy)
	allowlist := buildPanicAllowlist(t, productionPanicPolicy.Allowances)

	assertNoUnexpectedPanics(t, found, allowlist)
	assertNoStalePanicAllowances(t, found, allowlist)
}

// key returns the stable lookup key for one approved panic boundary.
func (allowance panicAllowance) key() panicKey {
	return panicKey{
		Path:     filepath.ToSlash(allowance.Path),
		Function: allowance.Function,
	}
}

// String formats a panic boundary for deterministic test diagnostics.
func (key panicKey) String() string {
	return key.Path + ":" + key.Function
}

// repoRootFromTest returns the repository root for this command package test.
func repoRootFromTest(t *testing.T) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file")
	}
	root, err := filepath.Abs(filepath.Join(filepath.Dir(currentFile), "..", "..", "..", ".."))
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	return root
}

// buildPanicAllowlist validates configured allowances and indexes them by key.
func buildPanicAllowlist(t *testing.T, allowances []panicAllowance) map[panicKey]panicAllowance {
	t.Helper()

	allowlist := make(map[panicKey]panicAllowance, len(allowances))
	for index, allowance := range allowances {
		key := allowance.key()
		if strings.TrimSpace(key.Path) == "" {
			t.Fatalf("panic allowance %d must include path", index+1)
		}
		if strings.TrimSpace(key.Function) == "" {
			t.Fatalf("panic allowance %d must include function", index+1)
		}
		if allowance.Count <= 0 {
			t.Fatalf("panic allowance %s must expect at least one call", key)
		}
		if strings.TrimSpace(string(allowance.Category)) == "" {
			t.Fatalf("panic allowance %s must include category", key)
		}
		if strings.TrimSpace(allowance.Reason) == "" {
			t.Fatalf("panic allowance %s must include reason", key)
		}
		if _, exists := allowlist[key]; exists {
			t.Fatalf("panic allowance %s is duplicated", key)
		}
		allowlist[key] = allowance
	}
	return allowlist
}

// scanProductionPanicCalls records panic call counts by file and enclosing function.
func scanProductionPanicCalls(t *testing.T, repoRoot string, policy panicAuditPolicy) map[panicKey]int {
	t.Helper()

	found := make(map[panicKey]int)
	skipDirs := skipDirSet(policy.SkipDirs)
	for _, root := range policy.ScanRoots {
		scanRootForPanicCalls(t, repoRoot, filepath.Join(repoRoot, filepath.FromSlash(root)), skipDirs, found)
	}
	return found
}

// skipDirSet builds a lookup set for directories excluded from production scanning.
func skipDirSet(names []string) map[string]struct{} {
	skipDirs := make(map[string]struct{}, len(names))
	for _, name := range names {
		skipDirs[name] = struct{}{}
	}
	return skipDirs
}

// scanRootForPanicCalls scans one source root for production panic calls.
func scanRootForPanicCalls(
	t *testing.T,
	repoRoot string,
	scanRoot string,
	skipDirs map[string]struct{},
	found map[panicKey]int,
) {
	t.Helper()

	err := filepath.WalkDir(scanRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if _, skip := skipDirs[entry.Name()]; skip {
				return filepath.SkipDir
			}
			return nil
		}
		if !isProductionGoFile(path) {
			return nil
		}
		relPath, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		scanFilePanicCalls(t, path, filepath.ToSlash(relPath), found)
		return nil
	})
	if err != nil {
		t.Fatalf("scan panic usage under %s: %v", scanRoot, err)
	}
}

// isProductionGoFile reports whether the path is a non-test Go source file.
func isProductionGoFile(path string) bool {
	return strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go")
}

// scanFilePanicCalls parses one Go file and records panic calls by enclosing function.
func scanFilePanicCalls(t *testing.T, path string, relPath string, found map[panicKey]int) {
	t.Helper()

	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", relPath, err)
	}
	for _, decl := range parsed.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		key := panicKey{
			Path:     relPath,
			Function: functionAllowlistName(fn),
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			ident, ok := call.Fun.(*ast.Ident)
			if ok && ident.Name == "panic" {
				found[key]++
			}
			return true
		})
	}
}

// assertNoUnexpectedPanics verifies every found panic is explicitly allowlisted.
func assertNoUnexpectedPanics(
	t *testing.T,
	found map[panicKey]int,
	allowlist map[panicKey]panicAllowance,
) {
	t.Helper()

	for _, key := range sortedPanicKeys(found) {
		count := found[key]
		allowance, ok := allowlist[key]
		if !ok {
			t.Errorf("panic call is not allowlisted: %s count=%d", key, count)
			continue
		}
		if allowance.Count != count {
			t.Errorf("panic count changed for %s: want %d, got %d", key, allowance.Count, count)
		}
	}
}

// assertNoStalePanicAllowances verifies every allowlist entry still matches source code.
func assertNoStalePanicAllowances(
	t *testing.T,
	found map[panicKey]int,
	allowlist map[panicKey]panicAllowance,
) {
	t.Helper()

	for _, key := range sortedPanicAllowanceKeys(allowlist) {
		if _, ok := found[key]; !ok {
			t.Errorf("panic allowance no longer matches any call: %s", key)
		}
	}
}

// sortedPanicKeys returns deterministic key ordering for count maps.
func sortedPanicKeys(items map[panicKey]int) []panicKey {
	keys := make([]panicKey, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sortPanicKeys(keys)
	return keys
}

// sortedPanicAllowanceKeys returns deterministic key ordering for allowance maps.
func sortedPanicAllowanceKeys(items map[panicKey]panicAllowance) []panicKey {
	keys := make([]panicKey, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sortPanicKeys(keys)
	return keys
}

// sortPanicKeys sorts panic keys by path and function for stable diagnostics.
func sortPanicKeys(keys []panicKey) {
	sort.Slice(keys, func(i int, j int) bool {
		if keys[i].Path == keys[j].Path {
			return keys[i].Function < keys[j].Function
		}
		return keys[i].Path < keys[j].Path
	})
}

// functionAllowlistName formats top-level and method declarations for stable
// allowlist keys.
func functionAllowlistName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return fn.Name.Name
	}
	return receiverName(fn.Recv.List[0].Type) + "." + fn.Name.Name
}

// receiverName formats one method receiver type without using source positions.
func receiverName(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return "(*" + receiverName(typed.X) + ")"
	case *ast.IndexExpr:
		return receiverName(typed.X)
	case *ast.IndexListExpr:
		return receiverName(typed.X)
	case *ast.SelectorExpr:
		return receiverName(typed.X) + "." + typed.Sel.Name
	default:
		return "unknown"
	}
}
