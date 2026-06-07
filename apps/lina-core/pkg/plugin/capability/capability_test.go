// This file verifies public capability contracts and repository-wide plugin
// capability boundary governance using cross-platform Go tests.

package capability

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/tenantcap"
)

// transientWasmDispatcherFileName is generated during WASM packaging and removed
// after the build, so repository governance scans must not treat it as stable source.
const transientWasmDispatcherFileName = "zz_generated_wasm_dispatcher.go"

// TestServicesDoesNotExposeTenantFilter verifies database-query-builder
// tenant filtering stays out of the ordinary capability services.
func TestServicesDoesNotExposeTenantFilter(t *testing.T) {
	servicesType := reflect.TypeOf((*Services)(nil)).Elem()
	if _, ok := servicesType.MethodByName("TenantFilter"); ok {
		t.Fatal("capability.Services must not expose TenantFilter")
	}
}

// TestServicesExposeUniqueDomainEntries verifies the plugin-facing capability
// directory does not publish duplicate legacy domain objects alongside the new
// typed domain services.
func TestServicesExposeUniqueDomainEntries(t *testing.T) {
	servicesType := reflect.TypeOf((*Services)(nil)).Elem()
	for _, removed := range []string{"Notify", "Session", "RuntimeConfig", "Config", "PluginConfig", "PluginLifecycle", "PluginState"} {
		if _, ok := servicesType.MethodByName(removed); ok {
			t.Fatalf("capability.Services must not expose removed %s entry", removed)
		}
	}
	for _, required := range []string{"HostConfig", "Notifications", "Plugins", "Sessions"} {
		if _, ok := servicesType.MethodByName(required); !ok {
			t.Fatalf("capability.Services must expose %s", required)
		}
	}

	adminType := reflect.TypeOf((*AdminServices)(nil)).Elem()
	if _, ok := adminType.MethodByName("Notify"); ok {
		t.Fatal("capability.AdminServices must not expose removed Notify entry")
	}
	if _, ok := adminType.MethodByName("Notifications"); !ok {
		t.Fatal("capability.AdminServices must expose Notifications")
	}
}

// TestOrgServiceDoesNotExposeWorkspaceProjections verifies ordinary
// organization capability consumption stays independent from user-management
// workspace projections.
func TestOrgServiceDoesNotExposeWorkspaceProjections(t *testing.T) {
	serviceType := reflect.TypeOf((*orgcap.Service)(nil)).Elem()
	for _, method := range []string{"UserDeptTree", "ListPostOptions"} {
		if _, ok := serviceType.MethodByName(method); ok {
			t.Fatalf("orgcap.Service must not expose host workspace projection method %s", method)
		}
	}
}

// TestTenantRuntimeServiceDoesNotExposeFallback verifies unused platform
// fallback helpers stay out of the broad runtime combination interface.
func TestTenantRuntimeServiceDoesNotExposeFallback(t *testing.T) {
	serviceType := reflect.TypeOf((*tenantcap.RuntimeService)(nil)).Elem()
	if _, ok := serviceType.MethodByName("ReadWithPlatformFallback"); ok {
		t.Fatal("tenantcap.RuntimeService must not expose ReadWithPlatformFallback")
	}
}

// TestRepositoryPluginCapabilityBoundaries scans production Go code for removed
// plugin capability access paths that would bypass the public capability seams.
func TestRepositoryPluginCapabilityBoundaries(t *testing.T) {
	t.Parallel()

	root := findCapabilityGovernanceRepositoryRoot(t)
	var findings []string
	findings = append(findings, scanCapabilityPackageInternalPluginImports(t, root)...)
	findings = append(findings, scanPluginCodeHostInternalImports(t, root)...)
	findings = append(findings, scanRemovedPluginCapabilityProductionReferences(t, root)...)
	findings = append(findings, scanPluginbridgeProtocolOwnership(t, root)...)
	if len(findings) == 0 {
		return
	}
	sort.Strings(findings)
	t.Fatalf("plugin capability boundary violations:\n%s", strings.Join(findings, "\n"))
}

// TestCollectGoSourcePathsSkipsTransientWasmDispatcher verifies concurrent WASM
// packaging artifacts do not enter repository-wide governance scans.
func TestCollectGoSourcePathsSkipsTransientWasmDispatcher(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	backendDir := filepath.Join(root, "apps", "lina-plugins", "demo", "backend")
	if err := os.MkdirAll(backendDir, 0755); err != nil {
		t.Fatalf("create backend dir: %v", err)
	}
	stableSourcePath := filepath.Join(backendDir, "plugin.go")
	if err := os.WriteFile(stableSourcePath, []byte("package backend\n"), 0644); err != nil {
		t.Fatalf("write stable source: %v", err)
	}
	transientSourcePath := filepath.Join(backendDir, transientWasmDispatcherFileName)
	if err := os.WriteFile(transientSourcePath, []byte("package backend\n"), 0644); err != nil {
		t.Fatalf("write transient source: %v", err)
	}

	paths := collectGoSourcePaths(t, root, false, "apps/lina-plugins")
	if len(paths) != 1 {
		t.Fatalf("expected only stable source, got %#v", paths)
	}
	if paths[0] != stableSourcePath {
		t.Fatalf("expected stable source %q, got %#v", stableSourcePath, paths)
	}
}

// scanCapabilityPackageInternalPluginImports verifies public plugin packages do
// not import host plugin-service internals.
func scanCapabilityPackageInternalPluginImports(t *testing.T, root string) []string {
	t.Helper()

	var findings []string
	for _, file := range parseGoSources(t, root, true, "apps/lina-core/pkg/plugin") {
		for _, importPath := range file.importPaths {
			if importPath == "lina-core/internal/service/plugin" || strings.HasPrefix(importPath, "lina-core/internal/service/plugin/") {
				findings = append(findings, fmt.Sprintf("%s imports host plugin service internal package %q", file.relPath, importPath))
			}
		}
	}
	return findings
}

// scanPluginCodeHostInternalImports verifies official plugin code only depends
// on published host contracts.
func scanPluginCodeHostInternalImports(t *testing.T, root string) []string {
	t.Helper()

	var findings []string
	for _, file := range parseGoSources(t, root, true, "apps/lina-plugins") {
		for _, importPath := range file.importPaths {
			if importPath == "lina-core/internal" || strings.HasPrefix(importPath, "lina-core/internal/") {
				findings = append(findings, fmt.Sprintf("%s imports host internal package %q", file.relPath, importPath))
			}
		}
	}
	return findings
}

// scanRemovedPluginCapabilityProductionReferences verifies production code does
// not reintroduce removed plugin host-service or bridge business-client entries.
func scanRemovedPluginCapabilityProductionReferences(t *testing.T, root string) []string {
	t.Helper()

	var findings []string
	for _, file := range parseGoSources(t, root, false, "apps/lina-core", "apps/lina-plugins", "hack/tools/linactl") {
		findings = append(findings, scanRemovedPluginCapabilityImports(file)...)
		findings = append(findings, scanRemovedProviderEnvReferences(file)...)
		findings = append(findings, scanRemovedHostServicesReferences(file)...)
		findings = append(findings, scanRemovedPluginbridgeBusinessReferences(file)...)
	}
	return findings
}

// scanRemovedPluginCapabilityImports blocks deleted pre-boundary package paths.
func scanRemovedPluginCapabilityImports(file parsedCapabilityGovernanceFile) []string {
	var findings []string
	for _, importPath := range file.importPaths {
		if removedPluginCapabilityImport(importPath) {
			findings = append(findings, fmt.Sprintf("%s imports removed plugin package %q", file.relPath, importPath))
		}
	}
	return findings
}

// scanRemovedProviderEnvReferences blocks the removed ProviderEnv.Services
// escape hatch.
func scanRemovedProviderEnvReferences(file parsedCapabilityGovernanceFile) []string {
	var findings []string
	if offset := bytes.Index(file.content, []byte("ProviderEnv.Services")); offset >= 0 {
		findings = append(findings, fmt.Sprintf("%s:%d references removed ProviderEnv.Services", file.relPath, lineForOffset(file.content, offset)))
	}
	ast.Inspect(file.astFile, func(node ast.Node) bool {
		typeSpec, ok := node.(*ast.TypeSpec)
		if !ok || typeSpec.Name.Name != "ProviderEnv" {
			return true
		}
		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}
		for _, field := range structType.Fields.List {
			for _, name := range field.Names {
				if name.Name == "Services" {
					findings = append(findings, fmt.Sprintf("%s declares removed ProviderEnv.Services field", file.position(name.Pos())))
				}
			}
		}
		return true
	})
	return findings
}

// scanRemovedHostServicesReferences blocks removed source-plugin host service
// directory accessors and misleading pluginhost service-directory names while
// allowing manifest HostServices data fields.
func scanRemovedHostServicesReferences(file parsedCapabilityGovernanceFile) []string {
	var findings []string
	pluginhostImports := file.importNamesForPath("lina-core/pkg/plugin/pluginhost")
	inPluginhostPackage := strings.HasPrefix(file.relPath, "apps/lina-core/pkg/plugin/pluginhost/")
	ast.Inspect(file.astFile, func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.FuncDecl:
			if typed.Name.Name == "HostServices" || typed.Name.Name == "HostServicesForPlugin" {
				findings = append(findings, fmt.Sprintf("%s declares removed %s function or method", file.position(typed.Name.Pos()), typed.Name.Name))
			}
			if inPluginhostPackage && typed.Name.Name == "Capabilities" {
				findings = append(findings, fmt.Sprintf("%s declares misleading pluginhost Capabilities() service-directory accessor", file.position(typed.Name.Pos())))
			}
		case *ast.InterfaceType:
			for _, method := range typed.Methods.List {
				for _, name := range method.Names {
					if name.Name == "HostServices" {
						findings = append(findings, fmt.Sprintf("%s declares removed HostServices() interface method", file.position(name.Pos())))
					}
					if inPluginhostPackage && name.Name == "Capabilities" {
						findings = append(findings, fmt.Sprintf("%s declares misleading pluginhost Capabilities() service-directory interface method", file.position(name.Pos())))
					}
				}
			}
		case *ast.CallExpr:
			switch fun := typed.Fun.(type) {
			case *ast.Ident:
				if fun.Name == "HostServicesForPlugin" {
					findings = append(findings, fmt.Sprintf("%s calls removed HostServicesForPlugin helper", file.position(fun.Pos())))
				}
			case *ast.SelectorExpr:
				if fun.Sel.Name == "HostServicesForPlugin" {
					findings = append(findings, fmt.Sprintf("%s calls removed HostServicesForPlugin helper", file.position(fun.Sel.Pos())))
				}
				if fun.Sel.Name == "HostServices" {
					findings = append(findings, fmt.Sprintf("%s calls removed HostServices() accessor", file.position(fun.Sel.Pos())))
				}
				if fun.Sel.Name == "Capabilities" && (inPluginhostPackage || len(pluginhostImports) > 0) {
					findings = append(findings, fmt.Sprintf("%s calls misleading pluginhost Capabilities() service-directory accessor", file.position(fun.Sel.Pos())))
				}
			}
		case *ast.SelectorExpr:
			if typed.Sel.Name == "HostServices" && selectorUsesImport(typed, pluginhostImports) {
				findings = append(findings, fmt.Sprintf("%s references removed pluginhost.HostServices type", file.position(typed.Sel.Pos())))
			}
		}
		return true
	})
	return findings
}

// scanRemovedPluginbridgeBusinessReferences blocks root pluginbridge business
// clients; dynamic plugin business code must use pluginbridge/guest instead.
func scanRemovedPluginbridgeBusinessReferences(file parsedCapabilityGovernanceFile) []string {
	var findings []string
	pluginbridgeImports := file.importNamesForPath("lina-core/pkg/plugin/pluginbridge")
	forbiddenFunctions := map[string]struct{}{
		"Runtime":    {},
		"Storage":    {},
		"HTTP":       {},
		"Network":    {},
		"Data":       {},
		"Cache":      {},
		"Lock":       {},
		"Config":     {},
		"Notify":     {},
		"Cron":       {},
		"HostConfig": {},
		"Manifest":   {},
		"Org":        {},
		"Tenant":     {},
	}
	forbiddenTypes := map[string]struct{}{
		"RuntimeHostService":    {},
		"StorageHostService":    {},
		"HTTPHostService":       {},
		"NetworkHostService":    {},
		"DataHostService":       {},
		"CacheHostService":      {},
		"LockHostService":       {},
		"ConfigHostService":     {},
		"NotifyHostService":     {},
		"CronHostService":       {},
		"HostConfigHostService": {},
		"ManifestHostService":   {},
		"OrgService":            {},
		"TenantService":         {},
	}
	ast.Inspect(file.astFile, func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.CallExpr:
			selector, ok := typed.Fun.(*ast.SelectorExpr)
			if !ok || !selectorUsesImport(selector, pluginbridgeImports) {
				return true
			}
			if _, forbidden := forbiddenFunctions[selector.Sel.Name]; forbidden {
				findings = append(findings, fmt.Sprintf("%s calls removed pluginbridge.%s business client", file.position(selector.Sel.Pos()), selector.Sel.Name))
			}
		case *ast.SelectorExpr:
			if !selectorUsesImport(typed, pluginbridgeImports) {
				return true
			}
			if _, forbidden := forbiddenTypes[typed.Sel.Name]; forbidden {
				findings = append(findings, fmt.Sprintf("%s references removed pluginbridge.%s business type", file.position(typed.Sel.Pos()), typed.Sel.Name))
			}
		case *ast.TypeSpec:
			if _, forbidden := forbiddenTypes[typed.Name.Name]; forbidden && file.relPath == "apps/lina-core/pkg/plugin/pluginbridge/pluginbridge.go" {
				findings = append(findings, fmt.Sprintf("%s declares removed pluginbridge business type %s", file.position(typed.Name.Pos()), typed.Name.Name))
			}
		case *ast.FuncDecl:
			if _, forbidden := forbiddenFunctions[typed.Name.Name]; forbidden && file.relPath == "apps/lina-core/pkg/plugin/pluginbridge/pluginbridge.go" {
				findings = append(findings, fmt.Sprintf("%s declares removed pluginbridge business client %s", file.position(typed.Name.Pos()), typed.Name.Name))
			}
		}
		return true
	})
	return findings
}

// scanPluginbridgeProtocolOwnership verifies bridge wire contracts are owned by
// pluginbridge/protocol, while higher-level guest SDK packages keep their
// intentional runtime transport seams.
func scanPluginbridgeProtocolOwnership(t *testing.T, root string) []string {
	t.Helper()

	var findings []string
	findings = append(findings, scanRecordStoreGuestImports(t, root)...)
	findings = append(findings, scanPluginbridgeRootInternalProtocolImports(t, root)...)
	findings = append(findings, scanPluginbridgeRootFacadeAliases(t, root)...)
	findings = append(findings, scanGuestRuntimeProtocolAliases(t, root)...)
	return findings
}

// scanRecordStoreGuestImports blocks the record store facade from importing
// pluginbridge/guest directly. pluginbridge/guest injects a narrow invoker so
// the guest SDK can own the public directory without creating an import cycle.
func scanRecordStoreGuestImports(t *testing.T, root string) []string {
	t.Helper()

	var findings []string
	for _, file := range parseGoSources(t, root, true, "apps/lina-core/pkg/plugin/capability/recordstore") {
		for _, importPath := range file.importPaths {
			switch importPath {
			case "lina-core/pkg/plugin/pluginbridge/guest":
				findings = append(findings, fmt.Sprintf("%s imports pluginbridge/guest instead of using an injected host-service invoker", file.relPath))
			case "lina-core/pkg/plugin/pluginbridge/contract":
				findings = append(findings, fmt.Sprintf("%s imports pluginbridge/contract instead of pluginbridge/protocol", file.relPath))
			}
		}
	}
	return findings
}

// scanPluginbridgeRootInternalProtocolImports blocks the root facade from
// directly importing internal protocol subcomponents. The public protocol package
// owns those aliases so the facade cannot drift into a second protocol surface.
func scanPluginbridgeRootInternalProtocolImports(t *testing.T, root string) []string {
	t.Helper()

	var findings []string
	for _, file := range parseGoSources(t, root, true, "apps/lina-core/pkg/plugin/pluginbridge") {
		if file.relPath != "apps/lina-core/pkg/plugin/pluginbridge/pluginbridge.go" {
			continue
		}
		for _, importPath := range file.importPaths {
			if importPath == "lina-core/pkg/plugin/pluginbridge/contract" ||
				strings.HasPrefix(importPath, "lina-core/pkg/plugin/pluginbridge/internal/") {
				findings = append(findings, fmt.Sprintf("%s imports low-level protocol implementation package %q", file.relPath, importPath))
			}
		}
	}
	return findings
}

// scanPluginbridgeRootFacadeAliases prevents the root pluginbridge package from
// growing back into a second protocol or guest facade.
func scanPluginbridgeRootFacadeAliases(t *testing.T, root string) []string {
	t.Helper()

	targetRelPath := "apps/lina-core/pkg/plugin/pluginbridge/pluginbridge.go"
	for _, file := range parseGoSources(t, root, true, "apps/lina-core/pkg/plugin/pluginbridge") {
		if file.relPath != targetRelPath {
			continue
		}
		var findings []string
		ast.Inspect(file.astFile, func(node ast.Node) bool {
			switch typed := node.(type) {
			case *ast.TypeSpec:
				if selectorReferencesPackage(typed.Type, "protocol", "guest") {
					findings = append(findings, fmt.Sprintf("%s aliases protocol or guest type %s from the pluginbridge root", file.position(typed.Name.Pos()), typed.Name.Name))
				}
			case *ast.ValueSpec:
				for _, value := range typed.Values {
					if !selectorReferencesPackage(value, "protocol", "guest") {
						continue
					}
					for _, name := range typed.Names {
						findings = append(findings, fmt.Sprintf("%s aliases protocol or guest value %s from the pluginbridge root", file.position(name.Pos()), name.Name))
					}
				}
			}
			return true
		})
		return findings
	}
	return []string{targetRelPath + " is missing"}
}

// scanGuestRuntimeProtocolAliases keeps pluginbridge/guest from becoming a
// second protocol owner. Guest helpers may use protocol DTOs in method
// signatures and implementations, but they must not alias protocol symbols back
// into the guest package.
func scanGuestRuntimeProtocolAliases(t *testing.T, root string) []string {
	t.Helper()

	targetRelPath := filepath.ToSlash(filepath.Join(
		root,
		"apps/lina-core/pkg/plugin/pluginbridge/guest/guest_types_aliases.go",
	))
	if fileExists(targetRelPath) {
		return []string{"apps/lina-core/pkg/plugin/pluginbridge/guest/guest_types_aliases.go reintroduces a guest protocol alias file"}
	}

	var findings []string
	for _, file := range parseGoSources(t, root, true, "apps/lina-core/pkg/plugin/pluginbridge/guest") {
		protocolImports := file.importNamesForPath("lina-core/pkg/plugin/pluginbridge/protocol")
		if len(protocolImports) == 0 {
			continue
		}
		ast.Inspect(file.astFile, func(node ast.Node) bool {
			switch typed := node.(type) {
			case *ast.TypeSpec:
				if typed.Assign.IsValid() && selectorReferencesImport(typed.Type, protocolImports) {
					findings = append(findings, fmt.Sprintf("%s aliases protocol type %s from pluginbridge/guest", file.position(typed.Name.Pos()), typed.Name.Name))
				}
			case *ast.ValueSpec:
				for _, value := range typed.Values {
					if !selectorReferencesImport(value, protocolImports) {
						continue
					}
					for _, name := range typed.Names {
						findings = append(findings, fmt.Sprintf("%s aliases protocol value %s from pluginbridge/guest", file.position(name.Pos()), name.Name))
					}
				}
			}
			return true
		})
	}
	return findings
}

// parsedCapabilityGovernanceFile is a parsed Go source file with stable
// repository-relative diagnostics.
type parsedCapabilityGovernanceFile struct {
	absPath     string
	relPath     string
	content     []byte
	fileSet     *token.FileSet
	astFile     *ast.File
	importPaths []string
	importNames map[string][]string
}

// parseGoSources parses Go files under relRoots. includeTests controls whether
// *_test.go files participate in the scan.
func parseGoSources(t *testing.T, root string, includeTests bool, relRoots ...string) []parsedCapabilityGovernanceFile {
	t.Helper()

	paths := collectGoSourcePaths(t, root, includeTests, relRoots...)
	files := make([]parsedCapabilityGovernanceFile, 0, len(paths))
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read Go source %s: %v", path, err)
		}
		fileSet := token.NewFileSet()
		astFile, err := parser.ParseFile(fileSet, path, content, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse Go source %s: %v", path, err)
		}
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatalf("relativize Go source %s: %v", path, err)
		}
		parsed := parsedCapabilityGovernanceFile{
			absPath:     path,
			relPath:     filepath.ToSlash(relPath),
			content:     content,
			fileSet:     fileSet,
			astFile:     astFile,
			importNames: make(map[string][]string),
		}
		for _, importSpec := range astFile.Imports {
			importPath, err := strconv.Unquote(importSpec.Path.Value)
			if err != nil {
				t.Fatalf("parse import path %s in %s: %v", importSpec.Path.Value, parsed.relPath, err)
			}
			parsed.importPaths = append(parsed.importPaths, importPath)
			name := importLocalName(importSpec, importPath)
			if name == "." {
				t.Fatalf("%s uses dot import for %q; plugin boundary scans require explicit imports", parsed.relPath, importPath)
			}
			if name != "" && name != "_" {
				parsed.importNames[importPath] = append(parsed.importNames[importPath], name)
			}
		}
		files = append(files, parsed)
	}
	return files
}

// collectGoSourcePaths returns deterministic Go source paths for the requested
// repository-relative roots.
func collectGoSourcePaths(t *testing.T, root string, includeTests bool, relRoots ...string) []string {
	t.Helper()

	var paths []string
	for _, relRoot := range relRoots {
		absRoot := filepath.Join(root, filepath.FromSlash(relRoot))
		if _, err := os.Stat(absRoot); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			t.Fatalf("stat scan root %s: %v", absRoot, err)
		}
		err := filepath.WalkDir(absRoot, func(path string, dirEntry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if dirEntry.IsDir() {
				switch dirEntry.Name() {
				case ".git", "node_modules", "dist", "build", "coverage", "temp", "vendor":
					return filepath.SkipDir
				default:
					return nil
				}
			}
			if !strings.HasSuffix(dirEntry.Name(), ".go") {
				return nil
			}
			if isCapabilityGovernanceTransientSource(dirEntry.Name()) {
				return nil
			}
			if !includeTests && strings.HasSuffix(dirEntry.Name(), "_test.go") {
				return nil
			}
			paths = append(paths, path)
			return nil
		})
		if err != nil {
			t.Fatalf("walk scan root %s: %v", absRoot, err)
		}
	}
	sort.Strings(paths)
	return paths
}

// isCapabilityGovernanceTransientSource reports whether a Go source path belongs
// to generated build-time artifacts that are not stable repository source.
func isCapabilityGovernanceTransientSource(name string) bool {
	return name == transientWasmDispatcherFileName
}

// findCapabilityGovernanceRepositoryRoot walks upward until it finds the
// repository root markers needed by this governance scan.
func findCapabilityGovernanceRepositoryRoot(t *testing.T) string {
	t.Helper()

	current, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for {
		if fileExists(filepath.Join(current, "openspec")) &&
			fileExists(filepath.Join(current, "apps", "lina-core", "go.mod")) &&
			fileExists(filepath.Join(current, "apps", "lina-plugins")) {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			t.Fatalf("could not locate repository root from %s", current)
		}
		current = parent
	}
}

// importLocalName returns the identifier used to reference an import.
func importLocalName(importSpec *ast.ImportSpec, importPath string) string {
	if importSpec.Name != nil {
		return importSpec.Name.Name
	}
	parts := strings.Split(importPath, "/")
	return parts[len(parts)-1]
}

// importNamesForPath returns all local identifiers used for importPath.
func (file parsedCapabilityGovernanceFile) importNamesForPath(importPath string) map[string]struct{} {
	names := make(map[string]struct{})
	for _, name := range file.importNames[importPath] {
		names[name] = struct{}{}
	}
	return names
}

// position formats a repository-relative source position.
func (file parsedCapabilityGovernanceFile) position(pos token.Pos) string {
	position := file.fileSet.Position(pos)
	return fmt.Sprintf("%s:%d", file.relPath, position.Line)
}

// selectorUsesImport reports whether selector.X names one of importNames.
func selectorUsesImport(selector *ast.SelectorExpr, importNames map[string]struct{}) bool {
	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}
	_, ok = importNames[ident.Name]
	return ok
}

// selectorReferencesPackage reports whether an expression references one of the
// provided package identifiers through a selector.
func selectorReferencesPackage(expr ast.Expr, packageNames ...string) bool {
	if expr == nil {
		return false
	}
	names := make(map[string]struct{}, len(packageNames))
	for _, name := range packageNames {
		names[name] = struct{}{}
	}
	found := false
	ast.Inspect(expr, func(node ast.Node) bool {
		if found {
			return false
		}
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := selector.X.(*ast.Ident)
		if !ok {
			return true
		}
		if _, matched := names[ident.Name]; matched {
			found = true
			return false
		}
		return true
	})
	return found
}

// selectorReferencesImport reports whether an expression references one of the
// provided imported package identifiers through a selector.
func selectorReferencesImport(expr ast.Expr, importNames map[string]struct{}) bool {
	if expr == nil || len(importNames) == 0 {
		return false
	}
	found := false
	ast.Inspect(expr, func(node ast.Node) bool {
		if found {
			return false
		}
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if selectorUsesImport(selector, importNames) {
			found = true
			return false
		}
		return true
	})
	return found
}

// removedPluginCapabilityImport reports whether importPath points at a removed
// pre-boundary plugin package.
func removedPluginCapabilityImport(importPath string) bool {
	removedRoots := []string{
		"lina-core/pkg/orgcap",
		"lina-core/pkg/plugindb",
		"lina-core/pkg/plugin/capability/ai",
		"lina-core/pkg/plugin/capability/bizctx",
		"lina-core/pkg/plugin/capability/config",
		"lina-core/pkg/plugin/capability/contract",
		"lina-core/pkg/plugin/capability/guest",
		"lina-core/pkg/plugin/capability/hostconfig",
		"lina-core/pkg/plugin/capability/manifest",
		"lina-core/pkg/plugin/capability/pluginlifecycle",
		"lina-core/pkg/plugin/capability/pluginstate",
		"lina-core/pkg/plugin/capability/tenantfilter",
		"lina-core/pkg/plugin/capability/tenantfiltercap",
		"lina-core/pkg/pluginbridge",
		"lina-core/pkg/pluginhost",
		"lina-core/pkg/pluginservice",
		"lina-core/pkg/sourceupgrade",
		"lina-core/pkg/tenantcap",
	}
	for _, root := range removedRoots {
		if importPath == root || strings.HasPrefix(importPath, root+"/") {
			return true
		}
	}
	return false
}

// lineForOffset returns a one-based line number for offset.
func lineForOffset(content []byte, offset int) int {
	return bytes.Count(content[:offset], []byte("\n")) + 1
}

// fileExists reports whether path exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
