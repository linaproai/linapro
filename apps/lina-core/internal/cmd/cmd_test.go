// This file verifies shared command helpers and command-package governance
// checks for confirmations, source plugin route middleware, and panic usage.
// Official plugin backend/plugin.go init registration fail-fast panics are
// auto-allowed by AST pattern without enumerating plugin IDs in the host
// allowlist; non-registration panic patterns remain rejected.

package cmd

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"unicode"

	_ "lina-core/pkg/dbdriver"

	"lina-core/internal/utility/testsupport"
)

// TestRequireCommandConfirmation verifies sensitive command confirmation tokens
// are enforced for init and mock operations.
func TestRequireCommandConfirmation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		commandName    string
		confirmValue   string
		wantErr        bool
		wantSubstrings []string
	}{
		{
			name:         "init accepts matching confirmation",
			commandName:  initCommandName,
			confirmValue: initCommandName,
		},
		{
			name:         "mock accepts matching confirmation",
			commandName:  mockCommandName,
			confirmValue: mockCommandName,
		},
		{
			name:         "upgrade accepts matching confirmation",
			commandName:  upgradeCommandName,
			confirmValue: upgradeCommandName,
		},
		{
			name:         "init rejects missing confirmation",
			commandName:  initCommandName,
			confirmValue: "",
			wantErr:      true,
			wantSubstrings: []string{
				"command init performs sensitive upgrade or database operations",
				makeConfirmationExample(initCommandName),
				goRunConfirmationExample(initCommandName),
			},
		},
		{
			name:         "mock rejects wrong confirmation",
			commandName:  mockCommandName,
			confirmValue: initCommandName,
			wantErr:      true,
			wantSubstrings: []string{
				"command mock performs sensitive upgrade or database operations",
				makeConfirmationExample(mockCommandName),
				goRunConfirmationExample(mockCommandName),
			},
		},
		{
			name:         "upgrade rejects missing confirmation",
			commandName:  upgradeCommandName,
			confirmValue: "",
			wantErr:      true,
			wantSubstrings: []string{
				"command upgrade performs sensitive upgrade or database operations",
				makeConfirmationExample(upgradeCommandName),
				goRunConfirmationExample(upgradeCommandName),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := requireCommandConfirmation(tt.commandName, tt.confirmValue)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for command %q", tt.commandName)
				}
				for _, substring := range tt.wantSubstrings {
					if !strings.Contains(err.Error(), substring) {
						t.Fatalf("expected error %q to contain %q", err.Error(), substring)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

// TestCommandPackageHasNoHanText verifies CLI diagnostics in this package stay
// as English developer-facing source text.
func TestCommandPackageHasNoHanText(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read command package directory: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
			continue
		}
		content, readErr := os.ReadFile(entry.Name())
		if readErr != nil {
			t.Fatalf("read %s: %v", entry.Name(), readErr)
		}
		for _, r := range string(content) {
			if unicode.Is(unicode.Han, r) {
				t.Fatalf("%s contains Han text; command diagnostics must use English source text", entry.Name())
			}
		}
	}
}

// TestParseInitRebuildFlag verifies the optional rebuild flag accepts common
// boolean spellings and rejects ambiguous values.
func TestParseInitRebuildFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    bool
		wantErr bool
	}{
		{name: "empty defaults to false", input: "", want: false},
		{name: "true enables rebuild", input: "true", want: true},
		{name: "one enables rebuild", input: "1", want: true},
		{name: "yes enables rebuild", input: "yes", want: true},
		{name: "false disables rebuild", input: "false", want: false},
		{name: "zero disables rebuild", input: "0", want: false},
		{name: "no disables rebuild", input: "no", want: false},
		{name: "reject unknown value", input: "maybe", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseInitRebuildFlag(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("parse rebuild flag: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

// TestSourcePluginProtectedRoutesIncludeTenancy verifies source plugin APIs use
// the same Auth -> Tenancy -> Permission chain as host protected static APIs.
func TestSourcePluginProtectedRoutesIncludeTenancy(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", "..", "..", ".."))
	if !testsupport.OfficialPluginsWorkspaceReady(repoRoot) {
		t.Skip("official plugin workspace is not initialized")
	}
	pluginFiles, err := filepath.Glob(filepath.Join(repoRoot, "apps", "lina-plugins", "*", "backend", "plugin.go"))
	if err != nil {
		t.Fatalf("scan source plugin route files failed: %v", err)
	}
	if len(pluginFiles) == 0 {
		t.Fatal("expected source plugin route files")
	}

	for _, file := range pluginFiles {
		t.Run(filepath.Base(filepath.Dir(filepath.Dir(file))), func(t *testing.T) {
			content, readErr := os.ReadFile(file)
			if readErr != nil {
				t.Fatalf("read source plugin route file failed: %v", readErr)
			}
			text := string(content)
			if !strings.Contains(text, "middlewares.Auth()") {
				return
			}
			var (
				authIndex       = strings.Index(text, "middlewares.Auth()")
				tenancyIndex    = strings.Index(text, "middlewares.Tenancy()")
				permissionIndex = strings.Index(text, "middlewares.Permission()")
			)
			if tenancyIndex < 0 {
				t.Fatalf("protected source plugin route must include Tenancy middleware: %s", file)
			}
			if !(authIndex < tenancyIndex && tenancyIndex < permissionIndex) {
				t.Fatalf("protected source plugin route must use Auth -> Tenancy -> Permission order: %s", file)
			}
		})
	}
}

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
			Path:     "apps/lina-core/internal/service/config/config_workspace.go",
			Function: "mustNormalizeWorkspaceBasePath",
			Count:    7,
			Category: panicCategoryStaticConfig,
			Reason:   "invalid static workspace basePath would make frontend fallback route binding ambiguous, so startup must fail before serving HTTP traffic",
		},
		{
			Path:     "apps/lina-core/internal/service/config/config_raw.go",
			Function: "configureRuntimeParamCacheDomain",
			Count:    1,
			Category: panicCategoryStaticConfig,
			Reason:   "runtime-config cachecoord domain registration is a static consistency contract and failures make protected config freshness undefined",
		},
		{
			Path:     "apps/lina-core/internal/service/cachecoord/revisionctrl/revisionctrl_controller.go",
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
			Path:     "apps/lina-core/pkg/plugin/pluginbridge/pluginbridge_router.go",
			Function: "MustNewGuestControllerRouteDispatcher",
			Count:    1,
			Category: panicCategoryMustConstructor,
			Reason:   "Must constructor documents fail-fast behavior and has a non-Must alternative",
		},
		{
			Path:     "apps/lina-core/pkg/plugin/pluginbridge/internal/hostservice/hostservice_validation.go",
			Function: "MustNormalizeHostServiceSpecs",
			Count:    1,
			Category: panicCategoryMustConstructor,
			Reason:   "Must helper is reserved for compile-time host service declarations",
		},
		{
			Path:     "apps/lina-core/pkg/plugin/pluginbridge/internal/hostservice/hostservice_validation.go",
			Function: "MustNormalizeHostServiceSpecsForPlugin",
			Count:    1,
			Category: panicCategoryMustConstructor,
			Reason:   "Must helper is reserved for compile-time plugin-scoped host service declarations",
		},
		{
			Path:     "apps/lina-core/internal/service/plugin/internal/datahost/internal/host/host_db.go",
			Function: "registerPluginDataDrivers",
			Count:    1,
			Category: panicCategoryStartup,
			Reason:   "plugin data DB drivers must register once before plugin data access can work",
		},
		// Official source plugins under apps/lina-plugins/*/backend/plugin.go are not
		// enumerated here. Their init registration fail-fast panics are auto-allowed
		// by AST pattern matching (see classifyPluginRegistrationPanic).
	},
}

// TestProductionPanicsMatchAllowlist verifies production panic usage stays
// narrow and documented. Host code uses an explicit Path/Function/Count
// allowlist; official plugin backend/plugin.go init registration fail-fast
// panics are auto-allowed by AST pattern without enumerating plugin IDs.
func TestProductionPanicsMatchAllowlist(t *testing.T) {
	repoRoot := repoRootFromTest(t)
	if !testsupport.OfficialPluginsWorkspaceReady(repoRoot) {
		t.Skip("official plugin workspace is not initialized")
	}
	found := scanProductionPanicCalls(t, repoRoot, productionPanicPolicy)
	allowlist := buildPanicAllowlist(t, productionPanicPolicy.Allowances)

	assertNoUnexpectedPanics(t, repoRoot, found, allowlist)
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

// assertNoUnexpectedPanics verifies every found panic is explicitly allowlisted
// or matches the official plugin init registration fail-fast auto-allow pattern.
func assertNoUnexpectedPanics(
	t *testing.T,
	repoRoot string,
	found map[panicKey]int,
	allowlist map[panicKey]panicAllowance,
) {
	t.Helper()

	for _, key := range sortedPanicKeys(found) {
		count := found[key]
		allowance, ok := allowlist[key]
		if ok {
			if allowance.Count != count {
				t.Errorf("panic count changed for %s: want %d, got %d", key, allowance.Count, count)
			}
			continue
		}
		matched, allowed, reason := classifyPluginRegistrationPanic(repoRoot, key)
		if matched {
			if allowed {
				continue
			}
			t.Errorf(
				"unexpected panic pattern in plugin init %s category=%s count=%d: %s",
				key,
				panicCategoryPluginRegistration,
				count,
				reason,
			)
			continue
		}
		t.Errorf("panic call is not allowlisted: %s count=%d", key, count)
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

// isOfficialPluginBackendPluginGo reports whether relPath is a source plugin
// entry file at apps/lina-plugins/<plugin-id>/backend/plugin.go.
func isOfficialPluginBackendPluginGo(relPath string) bool {
	parts := strings.Split(filepath.ToSlash(relPath), "/")
	return len(parts) == 5 &&
		parts[0] == "apps" &&
		parts[1] == "lina-plugins" &&
		parts[2] != "" &&
		parts[2] != "." &&
		parts[2] != ".." &&
		parts[3] == "backend" &&
		parts[4] == "plugin.go"
}

// classifyPluginRegistrationPanic reports whether a scanned panic key is under
// an official plugin backend/plugin.go init and every panic in that init is a
// registration fail-fast of the form panic(errIdent). When the path/function
// matches the plugin entry convention but the AST pattern fails, matched is
// true and allowed is false so callers can emit a pattern-specific diagnostic.
func classifyPluginRegistrationPanic(repoRoot string, key panicKey) (matched bool, allowed bool, reason string) {
	if key.Function != "init" || !isOfficialPluginBackendPluginGo(key.Path) {
		return false, false, ""
	}
	ok, failReason := pluginInitRegistrationFailFast(repoRoot, key.Path)
	if !ok {
		return true, false, failReason
	}
	return true, true, ""
}

// pluginInitRegistrationFailFast parses one plugin.go file and reports whether
// its package-level init only uses registration fail-fast panics.
func pluginInitRegistrationFailFast(repoRoot string, relPath string) (bool, string) {
	absPath := filepath.Join(repoRoot, filepath.FromSlash(relPath))
	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, absPath, nil, 0)
	if err != nil {
		return false, "parse failed: " + err.Error()
	}
	return initFuncRegistrationFailFast(parsed)
}

// initFuncRegistrationFailFast inspects the package-level init function body.
// It requires at least one panic and that every panic is panic(<ident>).
func initFuncRegistrationFailFast(file *ast.File) (bool, string) {
	var initFn *ast.FuncDecl
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || fn.Name == nil || fn.Name.Name != "init" || fn.Body == nil {
			continue
		}
		initFn = fn
		break
	}
	if initFn == nil {
		return false, "missing package-level init"
	}

	panicCount := 0
	var unexpected string
	ast.Inspect(initFn.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		ident, ok := call.Fun.(*ast.Ident)
		if !ok || ident.Name != "panic" {
			return true
		}
		panicCount++
		if !isRegistrationFailFastPanicCall(call) {
			unexpected = "panic is not registration fail-fast panic(<ident>)"
			return false
		}
		return true
	})
	if unexpected != "" {
		return false, unexpected
	}
	if panicCount == 0 {
		return false, "init has no panic calls"
	}
	return true, ""
}

// isRegistrationFailFastPanicCall reports whether a panic call is the standard
// registration fail-fast form panic(err) / panic(someIdent). Literal messages,
// formatted strings, and constructed errors are rejected.
func isRegistrationFailFastPanicCall(call *ast.CallExpr) bool {
	if call == nil || len(call.Args) != 1 {
		return false
	}
	_, ok := call.Args[0].(*ast.Ident)
	return ok
}

// TestPluginRegistrationFailFastHelpers verifies auto-allow path checks and
// AST pattern acceptance / rejection without depending on the full workspace scan.
func TestPluginRegistrationFailFastHelpers(t *testing.T) {
	t.Parallel()

	t.Run("path matching", func(t *testing.T) {
		t.Parallel()
		cases := []struct {
			path string
			want bool
		}{
			{path: "apps/lina-plugins/linapro-demo-source/backend/plugin.go", want: true},
			{path: "apps/lina-plugins/acme-demo/backend/plugin.go", want: true},
			{path: "apps/lina-core/internal/cmd/cmd.go", want: false},
			{path: "apps/lina-plugins/linapro-demo-source/backend/other.go", want: false},
			{path: "apps/lina-plugins/linapro-demo-source/plugin.go", want: false},
			{path: "apps/lina-plugins/../evil/backend/plugin.go", want: false},
		}
		for _, tc := range cases {
			if got := isOfficialPluginBackendPluginGo(tc.path); got != tc.want {
				t.Fatalf("isOfficialPluginBackendPluginGo(%q)=%v, want %v", tc.path, got, tc.want)
			}
		}
	})

	t.Run("accepts panic ident fail-fast", func(t *testing.T) {
		t.Parallel()
		src := `package plugin
func init() {
	if err := register(); err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
}
func register() error { return nil }
`
		file := mustParseTestSource(t, src)
		ok, reason := initFuncRegistrationFailFast(file)
		if !ok {
			t.Fatalf("expected fail-fast init accepted, got reason=%q", reason)
		}
	})

	t.Run("rejects literal panic", func(t *testing.T) {
		t.Parallel()
		src := `package plugin
func init() {
	if err := register(); err != nil {
		panic(err)
	}
	panic("unexpected")
}
func register() error { return nil }
`
		file := mustParseTestSource(t, src)
		ok, reason := initFuncRegistrationFailFast(file)
		if ok {
			t.Fatal("expected literal panic to be rejected")
		}
		if reason == "" {
			t.Fatal("expected rejection reason")
		}
	})

	t.Run("rejects formatted panic", func(t *testing.T) {
		t.Parallel()
		src := `package plugin
import "fmt"
func init() {
	panic(fmt.Sprintf("bad %s", "x"))
}
`
		file := mustParseTestSource(t, src)
		ok, _ := initFuncRegistrationFailFast(file)
		if ok {
			t.Fatal("expected formatted panic to be rejected")
		}
	})

	t.Run("rejects non-init production panic path for auto-allow key", func(t *testing.T) {
		t.Parallel()
		key := panicKey{
			Path:     "apps/lina-plugins/demo/backend/plugin.go",
			Function: "registerRoutes",
		}
		matched, allowed, _ := classifyPluginRegistrationPanic(t.TempDir(), key)
		if matched || allowed {
			t.Fatal("non-init function must not match plugin registration auto-allow")
		}
	})
}

// mustParseTestSource parses an in-memory Go source file for helper unit tests.
func mustParseTestSource(t *testing.T, src string) *ast.File {
	t.Helper()
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "plugin.go", src, 0)
	if err != nil {
		t.Fatalf("parse test source: %v", err)
	}
	return file
}
