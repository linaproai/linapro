// This file loads backend hook and resource declarations, extracts dynamic
// route contracts from API DTOs, and validates the collected contracts.

package wasmbuilder

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	bridgeplugin "lina-core/pkg/plugin/pluginbridge"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

const (
	// pluginRegisterFunctionName is the dynamic backend callback inspected by
	// the builder to mirror source-plugin startup declarations.
	pluginRegisterFunctionName = "RegisterPlugin"
	// routeRegisterGroupMethodName is the registrar method used to bind a route
	// group prefix to one backend/api-relative package.
	routeRegisterGroupMethodName = "Group"
	// routeFacadeMethodName is the Declarations method returning route
	// declarations.
	routeFacadeMethodName = "Routes"
)

func collectHookSpecs(pluginDir string, pluginID string) ([]*hookSpec, error) {
	hookDir := filepath.Join(pluginDir, "backend", "hooks")
	entries, err := os.ReadDir(hookDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*hookSpec{}, nil
		}
		return nil, err
	}

	fileNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		fileNames = append(fileNames, entry.Name())
	}
	sortStrings(fileNames)

	items := make([]*hookSpec, 0, len(fileNames))
	for _, name := range fileNames {
		filePath := filepath.Join(hookDir, name)
		spec := &hookSpec{}
		if err = loadYAMLFile(filePath, spec); err != nil {
			return nil, err
		}
		if err = validateHookSpec(pluginID, spec, filePath); err != nil {
			return nil, err
		}
		items = append(items, spec)
	}
	return items, nil
}

func collectLifecycleSpecs(pluginDir string, pluginID string) ([]*protocol.LifecycleContract, error) {
	discovered, err := discoverLifecycleSpecs(pluginDir, pluginID)
	if err != nil {
		return nil, err
	}
	overrides, err := collectLifecycleOverrides(pluginDir, pluginID)
	if err != nil {
		return nil, err
	}
	items, err := mergeLifecycleSpecs(pluginID, discovered, overrides)
	if err != nil {
		return nil, err
	}
	if err = protocol.ValidateLifecycleContracts(pluginID, items); err != nil {
		return nil, fmt.Errorf("plugin lifecycle declaration is invalid: %w", err)
	}
	return items, nil
}

func collectLifecycleOverrides(pluginDir string, pluginID string) ([]*protocol.LifecycleContract, error) {
	lifecycleDir := filepath.Join(pluginDir, "backend", "lifecycle")
	entries, err := os.ReadDir(lifecycleDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*protocol.LifecycleContract{}, nil
		}
		return nil, err
	}

	fileNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		fileNames = append(fileNames, entry.Name())
	}
	sortStrings(fileNames)

	items := make([]*protocol.LifecycleContract, 0, len(fileNames))
	for _, name := range fileNames {
		filePath := filepath.Join(lifecycleDir, name)
		spec := &protocol.LifecycleContract{}
		if err = loadYAMLFile(filePath, spec); err != nil {
			return nil, err
		}
		protocol.NormalizeLifecycleContract(spec)
		if spec.Operation == "" {
			return nil, fmt.Errorf("plugin lifecycle override operation is unsupported for plugin %s: %s", pluginID, filePath)
		}
		if spec.TimeoutMs < 0 {
			return nil, fmt.Errorf("plugin lifecycle override timeoutMs cannot be negative for plugin %s operation %s: %s", pluginID, spec.Operation, filePath)
		}
		items = append(items, spec)
	}
	return items, nil
}

func discoverLifecycleSpecs(pluginDir string, pluginID string) ([]*protocol.LifecycleContract, error) {
	backendDir := filepath.Join(pluginDir, "backend")
	info, err := os.Stat(backendDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*protocol.LifecycleContract{}, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("runtime backend path is not a directory: %s", backendDir)
	}

	items := make([]*protocol.LifecycleContract, 0)
	seen := make(map[protocol.LifecycleOperation]string)
	err = filepath.WalkDir(backendDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if shouldSkipRuntimeBackendDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		if !isLifecycleControllerSourceFile(backendDir, path) {
			return nil
		}

		fileSet := token.NewFileSet()
		fileNode, parseErr := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
		if parseErr != nil {
			return fmt.Errorf("failed to parse backend file %s: %w", path, parseErr)
		}
		lifecycleTimeouts, timeoutErr := collectLifecycleTimeoutsFromFile(pluginID, fileNode)
		if timeoutErr != nil {
			return fmt.Errorf("failed to inspect lifecycle timeout in %s: %w", path, timeoutErr)
		}
		for _, decl := range fileNode.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl == nil || funcDecl.Recv == nil || funcDecl.Name == nil {
				continue
			}
			spec, matched, metadataErr := extractLifecycleSpecFromFunc(pluginID, funcDecl)
			if metadataErr != nil {
				return fmt.Errorf("failed to inspect lifecycle handler in %s: %w", path, metadataErr)
			}
			if !matched {
				continue
			}
			if previousPath, exists := seen[spec.Operation]; exists {
				return fmt.Errorf("plugin lifecycle handler operation is duplicated for plugin %s: %s in %s and %s", pluginID, spec.Operation, previousPath, path)
			}
			if timeoutMs, exists := lifecycleTimeouts[spec.Operation]; exists {
				spec.TimeoutMs = timeoutMs
			}
			seen[spec.Operation] = path
			items = append(items, spec)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sortLifecycleSpecs(items)
	return items, nil
}

func collectLifecycleTimeoutsFromFile(
	pluginID string,
	fileNode *ast.File,
) (map[protocol.LifecycleOperation]int, error) {
	timeouts := make(map[protocol.LifecycleOperation]int)
	if fileNode == nil {
		return timeouts, nil
	}
	for _, decl := range fileNode.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl == nil || genDecl.Tok != token.CONST {
			continue
		}
		for _, item := range genDecl.Specs {
			valueSpec, ok := item.(*ast.ValueSpec)
			if !ok || valueSpec == nil {
				continue
			}
			if err := collectLifecycleTimeoutsFromValueSpec(pluginID, valueSpec, timeouts); err != nil {
				return nil, err
			}
		}
	}
	return timeouts, nil
}

func collectLifecycleTimeoutsFromValueSpec(
	pluginID string,
	valueSpec *ast.ValueSpec,
	timeouts map[protocol.LifecycleOperation]int,
) error {
	for index, name := range valueSpec.Names {
		if name == nil {
			continue
		}
		operation, ok := lifecycleOperationFromTimeoutConst(name.Name)
		if !ok {
			continue
		}
		if _, exists := timeouts[operation]; exists {
			return fmt.Errorf("plugin lifecycle timeout constant is duplicated for plugin %s operation %s", pluginID, operation)
		}
		if index >= len(valueSpec.Values) {
			return fmt.Errorf("plugin lifecycle timeout constant requires an explicit integer value for plugin %s operation %s", pluginID, operation)
		}
		timeoutMs, err := lifecycleTimeoutConstValue(pluginID, operation, valueSpec.Values[index])
		if err != nil {
			return err
		}
		timeouts[operation] = timeoutMs
	}
	return nil
}

func lifecycleOperationFromTimeoutConst(name string) (protocol.LifecycleOperation, bool) {
	trimmed := strings.TrimSpace(name)
	if !strings.HasSuffix(trimmed, "TimeoutMs") {
		return "", false
	}
	operationName := strings.TrimSuffix(trimmed, "TimeoutMs")
	if !protocol.IsSupportedLifecycleOperation(operationName) {
		return "", false
	}
	return protocol.LifecycleOperation(operationName), true
}

func lifecycleTimeoutConstValue(
	pluginID string,
	operation protocol.LifecycleOperation,
	expr ast.Expr,
) (int, error) {
	basicLit, ok := expr.(*ast.BasicLit)
	if !ok || basicLit.Kind != token.INT {
		return 0, fmt.Errorf("plugin lifecycle timeout constant must be an integer literal for plugin %s operation %s", pluginID, operation)
	}
	timeoutMs, err := strconv.Atoi(strings.TrimSpace(basicLit.Value))
	if err != nil {
		return 0, fmt.Errorf("plugin lifecycle timeout constant is invalid for plugin %s operation %s: %w", pluginID, operation, err)
	}
	if timeoutMs < 0 {
		return 0, fmt.Errorf("plugin lifecycle timeout constant cannot be negative for plugin %s operation %s", pluginID, operation)
	}
	return timeoutMs, nil
}

func isLifecycleControllerSourceFile(backendDir string, filePath string) bool {
	relativePath, err := filepath.Rel(backendDir, filePath)
	if err != nil {
		return false
	}
	normalizedPath := filepath.ToSlash(filepath.Clean(relativePath))
	if strings.HasPrefix(normalizedPath, "internal/controller/") {
		return true
	}
	if strings.Contains(normalizedPath, "/") {
		return false
	}
	return strings.HasPrefix(filepath.Base(normalizedPath), "controller")
}

func shouldSkipRuntimeBackendDir(name string) bool {
	trimmed := strings.TrimSpace(name)
	return trimmed == "" || strings.HasPrefix(trimmed, ".") || strings.HasPrefix(trimmed, "_")
}

func extractLifecycleSpecFromFunc(pluginID string, decl *ast.FuncDecl) (*protocol.LifecycleContract, bool, error) {
	methodName := strings.TrimSpace(decl.Name.Name)
	if methodName == "" {
		return nil, false, nil
	}
	if !protocol.IsSupportedLifecycleOperation(methodName) {
		return nil, false, nil
	}

	requestType, ok, err := inferGuestHandlerRequestType(decl)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}
	spec := &protocol.LifecycleContract{
		Operation:    protocol.LifecycleOperation(methodName),
		RequestType:  requestType,
		InternalPath: buildLifecycleInternalPath(methodName),
	}
	protocol.NormalizeLifecycleContract(spec)
	return spec, true, nil
}

func inferGuestHandlerRequestType(decl *ast.FuncDecl) (string, bool, error) {
	params := flattenFieldTypes(decl.Type.Params)
	results := flattenFieldTypes(decl.Type.Results)
	if len(results) != 2 || !isErrorType(results[1]) {
		return "", false, nil
	}

	if len(params) == 2 && isContextType(params[0]) && isPointerToNamedType(params[1]) && isPointerToNamedType(results[0]) {
		requestType := pointerTypeName(params[1])
		if requestType == "" {
			return "", false, fmt.Errorf("typed lifecycle handler request DTO name is empty: %s", decl.Name.Name)
		}
		return requestType, true, nil
	}
	return "", false, nil
}

func flattenFieldTypes(fields *ast.FieldList) []ast.Expr {
	if fields == nil {
		return nil
	}
	items := make([]ast.Expr, 0, len(fields.List))
	for _, field := range fields.List {
		if field == nil || field.Type == nil {
			continue
		}
		nameCount := len(field.Names)
		if nameCount == 0 {
			items = append(items, field.Type)
			continue
		}
		for index := 0; index < nameCount; index++ {
			items = append(items, field.Type)
		}
	}
	return items
}

func isPointerToNamedType(expr ast.Expr) bool {
	return pointerTypeName(expr) != ""
}

func pointerTypeName(expr ast.Expr) string {
	starExpr, ok := expr.(*ast.StarExpr)
	if !ok || starExpr.X == nil {
		return ""
	}
	return astTypeName(starExpr.X)
}

func astTypeName(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return strings.TrimSpace(typed.Name)
	case *ast.SelectorExpr:
		if typed.Sel == nil {
			return ""
		}
		return strings.TrimSpace(typed.Sel.Name)
	default:
		return ""
	}
}

func isContextType(expr ast.Expr) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	return ok && selector.Sel != nil && selector.Sel.Name == "Context"
}

func isErrorType(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "error"
}

func buildLifecycleInternalPath(operation string) string {
	return "/__lifecycle" + bridgeplugin.BuildGuestControllerInternalPath(operation)
}

func mergeLifecycleSpecs(
	pluginID string,
	discovered []*protocol.LifecycleContract,
	overrides []*protocol.LifecycleContract,
) ([]*protocol.LifecycleContract, error) {
	byOperation := make(map[protocol.LifecycleOperation]*protocol.LifecycleContract, len(discovered))
	for _, item := range discovered {
		if item == nil {
			continue
		}
		protocol.NormalizeLifecycleContract(item)
		byOperation[item.Operation] = item
	}
	seenOverrides := make(map[protocol.LifecycleOperation]struct{}, len(overrides))
	for _, override := range overrides {
		if override == nil {
			return nil, fmt.Errorf("plugin lifecycle override cannot be nil for plugin %s", pluginID)
		}
		if _, exists := seenOverrides[override.Operation]; exists {
			return nil, fmt.Errorf("plugin lifecycle override operation is duplicated for plugin %s: %s", pluginID, override.Operation)
		}
		seenOverrides[override.Operation] = struct{}{}
		base, exists := byOperation[override.Operation]
		if !exists {
			return nil, fmt.Errorf("plugin lifecycle override has no matching handler for plugin %s operation %s", pluginID, override.Operation)
		}
		discoveredRequestType := strings.TrimSpace(base.RequestType)
		if strings.TrimSpace(override.RequestType) != "" {
			base.RequestType = strings.TrimSpace(override.RequestType)
		}
		if strings.TrimSpace(override.InternalPath) != "" {
			base.InternalPath = strings.TrimSpace(override.InternalPath)
		}
		if override.TimeoutMs > 0 {
			base.TimeoutMs = override.TimeoutMs
		}
		protocol.NormalizeLifecycleContract(base)
		if base.RequestType != discoveredRequestType {
			return nil, fmt.Errorf(
				"plugin lifecycle override requestType is not reachable by guest dispatcher for plugin %s operation %s",
				pluginID,
				override.Operation,
			)
		}
	}

	items := make([]*protocol.LifecycleContract, 0, len(byOperation))
	for _, item := range byOperation {
		items = append(items, item)
	}
	sortLifecycleSpecs(items)
	return items, nil
}

func sortLifecycleSpecs(items []*protocol.LifecycleContract) {
	sort.Slice(items, func(left int, right int) bool {
		return lifecycleOperationOrder(items[left].Operation) < lifecycleOperationOrder(items[right].Operation)
	})
}

func lifecycleOperationOrder(operation protocol.LifecycleOperation) int {
	switch operation {
	case protocol.LifecycleOperationBeforeInstall:
		return 10
	case protocol.LifecycleOperationAfterInstall:
		return 20
	case protocol.LifecycleOperationBeforeUpgrade:
		return 30
	case protocol.LifecycleOperationUpgrade:
		return 35
	case protocol.LifecycleOperationAfterUpgrade:
		return 40
	case protocol.LifecycleOperationBeforeDisable:
		return 50
	case protocol.LifecycleOperationAfterDisable:
		return 60
	case protocol.LifecycleOperationBeforeUninstall:
		return 70
	case protocol.LifecycleOperationUninstall:
		return 75
	case protocol.LifecycleOperationAfterUninstall:
		return 80
	case protocol.LifecycleOperationBeforeTenantDisable:
		return 90
	case protocol.LifecycleOperationAfterTenantDisable:
		return 100
	case protocol.LifecycleOperationBeforeTenantDelete:
		return 110
	case protocol.LifecycleOperationAfterTenantDelete:
		return 120
	case protocol.LifecycleOperationBeforeInstallModeChange:
		return 130
	case protocol.LifecycleOperationAfterInstallModeChange:
		return 140
	default:
		return 1000
	}
}

func collectResourceSpecs(pluginDir string, pluginID string) ([]*resourceSpec, error) {
	resourceDir := filepath.Join(pluginDir, "backend", "resources")
	entries, err := os.ReadDir(resourceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*resourceSpec{}, nil
		}
		return nil, err
	}

	fileNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		fileNames = append(fileNames, entry.Name())
	}
	sortStrings(fileNames)

	items := make([]*resourceSpec, 0, len(fileNames))
	for _, name := range fileNames {
		filePath := filepath.Join(resourceDir, name)
		spec := &resourceSpec{}
		if err = loadYAMLFile(filePath, spec); err != nil {
			return nil, err
		}
		if err = validateResourceSpec(pluginID, spec, filePath); err != nil {
			return nil, err
		}
		items = append(items, spec)
	}
	return items, nil
}

func collectRouteContracts(pluginDir string, pluginID string) ([]*routeContractSource, []*protocol.RouteContract, error) {
	apiDir := filepath.Join(pluginDir, "backend", "api")
	info, err := os.Stat(apiDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, []*protocol.RouteContract{}, nil
		}
		return nil, nil, err
	}
	if !info.IsDir() {
		return nil, nil, fmt.Errorf("runtime backend api path is not a directory: %s", apiDir)
	}

	prefixes, err := collectRouteGroupBindings(pluginDir, apiDir)
	if err != nil {
		return nil, nil, err
	}
	fset := token.NewFileSet()
	sources := make([]*routeContractSource, 0)
	err = filepath.WalkDir(apiDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		fileNode, parseErr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if parseErr != nil {
			return fmt.Errorf("failed to parse api file %s: %w", path, parseErr)
		}
		dir := filepath.Dir(path)
		apiPackage, relErr := backendAPIPackageForDir(apiDir, dir)
		if relErr != nil {
			return relErr
		}
		items, extractErr := extractRouteContractsFromFile(fileNode)
		if extractErr != nil {
			return fmt.Errorf("failed to extract route contract from %s: %w", path, extractErr)
		}
		sources = append(sources, &routeContractSource{
			dir:        dir,
			contracts:  items,
			apiPackage: apiPackage,
		})
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	contracts := make([]*protocol.RouteContract, 0)
	for _, source := range sources {
		routeGroupPrefix := routeGroupPrefixForDir(apiDir, source.dir, prefixes)
		for _, contract := range source.contracts {
			applyRouteGroupPrefix(routeGroupPrefix, contract)
			contracts = append(contracts, contract)
		}
	}
	if err = protocol.ValidateRouteContracts(pluginID, contracts); err != nil {
		return nil, nil, err
	}
	return sources, contracts, nil
}

// routeContractSource records DTO-derived route contracts before their
// registered route group prefix has been applied.
type routeContractSource struct {
	// dir is the API package directory containing the DTO declarations.
	dir string
	// contracts are DTO-derived route contracts before group-prefix composition.
	contracts []*protocol.RouteContract
	// apiPackage is the backend/api-relative package path containing the DTO declarations.
	apiPackage string
}

// backendAPIPackageForDir returns one backend/api-relative package path for a
// DTO source directory.
func backendAPIPackageForDir(apiDir string, sourceDir string) (string, error) {
	relativePath, err := filepath.Rel(apiDir, sourceDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve backend api package path for %s: %w", sourceDir, err)
	}
	normalizedPath := normalizeAPIPackagePath(filepath.ToSlash(relativePath))
	if normalizedPath == "" {
		return ".", nil
	}
	return normalizedPath, nil
}

// collectRouteGroupBindings reads dynamic backend route registration code and
// maps backend/api-relative packages to plugin-owned route group prefixes.
func collectRouteGroupBindings(pluginDir string, apiDir string) (map[string]string, error) {
	backendFile := filepath.Join(pluginDir, "backend", "plugin.go")
	fileNode, err := parser.ParseFile(token.NewFileSet(), backendFile, nil, parser.ParseComments)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("failed to parse dynamic backend plugin file %s: %w", backendFile, err)
	}

	prefixes := make(map[string]string)
	for _, decl := range fileNode.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl == nil || funcDecl.Name == nil || funcDecl.Name.Name != pluginRegisterFunctionName {
			continue
		}
		items, extractErr := extractRouteGroupBindingsFromFunc(funcDecl, collectStringConstsFromFile(fileNode))
		if extractErr != nil {
			return nil, fmt.Errorf("failed to extract dynamic route groups from %s: %w", backendFile, extractErr)
		}
		for _, item := range items {
			dir, dirErr := routeGroupBindingDir(apiDir, item.apiPackage)
			if dirErr != nil {
				return nil, fmt.Errorf("invalid dynamic route group api package %q in %s: %w", item.apiPackage, backendFile, dirErr)
			}
			if previousPrefix, ok := prefixes[dir]; ok && previousPrefix != item.prefix {
				return nil, fmt.Errorf("dynamic route group package %s is bound to conflicting prefixes: %s != %s", item.apiPackage, previousPrefix, item.prefix)
			}
			prefixes[dir] = item.prefix
		}
	}
	return prefixes, nil
}

// routeGroupBinding records one plugin.Routes().Group(prefix, apiPackage) call.
type routeGroupBinding struct {
	// prefix is the plugin-owned route group prefix.
	prefix string
	// apiPackage is the backend/api-relative package path.
	apiPackage string
}

// extractRouteGroupBindingsFromFunc extracts plugin.Routes().Group calls from
// one dynamic RegisterPlugin function.
func extractRouteGroupBindingsFromFunc(
	funcDecl *ast.FuncDecl,
	stringConsts map[string]string,
) ([]routeGroupBinding, error) {
	if funcDecl.Body == nil {
		return nil, nil
	}
	pluginNames := declarationParamNames(funcDecl)
	registrarNames := routeDeclarationParamNames(funcDecl)
	if len(pluginNames) == 0 && len(registrarNames) == 0 {
		return nil, nil
	}
	items := make([]routeGroupBinding, 0)
	var firstErr error
	ast.Inspect(funcDecl.Body, func(node ast.Node) bool {
		if firstErr != nil {
			return false
		}
		callExpr, ok := node.(*ast.CallExpr)
		if !ok || callExpr == nil {
			return true
		}
		selector, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok || selector == nil || selector.Sel == nil || selector.Sel.Name != routeRegisterGroupMethodName {
			return true
		}
		if !isRouteGroupReceiver(selector.X, pluginNames, registrarNames) {
			return true
		}
		if len(callExpr.Args) != 2 {
			firstErr = fmt.Errorf("registrar Group calls must pass prefix and api package strings")
			return false
		}
		prefix, err := stringArg(callExpr.Args[0], stringConsts)
		if err != nil {
			firstErr = fmt.Errorf("route group prefix must be a string literal or string const: %w", err)
			return false
		}
		apiPackage, err := stringArg(callExpr.Args[1], stringConsts)
		if err != nil {
			firstErr = fmt.Errorf("route group api package must be a string literal or string const: %w", err)
			return false
		}
		items = append(items, routeGroupBinding{
			prefix:     normalizeRouteDeclarationPath(prefix),
			apiPackage: normalizeAPIPackagePath(apiPackage),
		})
		return true
	})
	if firstErr != nil {
		return nil, firstErr
	}
	return items, nil
}

// declarationParamNames returns RegisterPlugin parameters with the dynamic
// plugin declaration contract type name.
func declarationParamNames(funcDecl *ast.FuncDecl) map[string]struct{} {
	return parameterNamesByTypeName(funcDecl, "Declarations")
}

// routeDeclarationParamNames returns direct route facade parameters that
// implement the route declaration contract by type name.
func routeDeclarationParamNames(funcDecl *ast.FuncDecl) map[string]struct{} {
	return parameterNamesByTypeName(funcDecl, "RouteDeclarations")
}

// parameterNamesByTypeName returns parameters matching one interface type name.
func parameterNamesByTypeName(funcDecl *ast.FuncDecl, typeName string) map[string]struct{} {
	names := make(map[string]struct{})
	if funcDecl == nil || funcDecl.Type == nil || funcDecl.Type.Params == nil {
		return names
	}
	for _, field := range funcDecl.Type.Params.List {
		if field == nil || astTypeName(field.Type) != typeName {
			continue
		}
		for _, name := range field.Names {
			if name == nil || strings.TrimSpace(name.Name) == "" {
				continue
			}
			names[name.Name] = struct{}{}
		}
	}
	return names
}

// isRouteGroupReceiver reports whether a Group call receiver is either a
// RouteDeclarations parameter or plugin.Routes() from a Declarations
// parameter.
func isRouteGroupReceiver(
	expr ast.Expr,
	pluginNames map[string]struct{},
	registrarNames map[string]struct{},
) bool {
	receiver, ok := expr.(*ast.Ident)
	if ok && receiver != nil {
		_, allowed := registrarNames[receiver.Name]
		return allowed
	}

	callExpr, ok := expr.(*ast.CallExpr)
	if !ok || callExpr == nil || len(callExpr.Args) != 0 {
		return false
	}
	selector, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok || selector == nil || selector.Sel == nil || selector.Sel.Name != routeFacadeMethodName {
		return false
	}
	pluginIdent, ok := selector.X.(*ast.Ident)
	if !ok || pluginIdent == nil {
		return false
	}
	_, allowed := pluginNames[pluginIdent.Name]
	return allowed
}

// collectStringConstsFromFile collects file-local string constants that route
// registration declarations may use for group prefixes.
func collectStringConstsFromFile(fileNode *ast.File) map[string]string {
	values := make(map[string]string)
	if fileNode == nil {
		return values
	}
	for _, decl := range fileNode.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}
		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for index, name := range valueSpec.Names {
				if name == nil || index >= len(valueSpec.Values) {
					continue
				}
				value, err := stringArg(valueSpec.Values[index], values)
				if err != nil {
					continue
				}
				values[name.Name] = value
			}
		}
	}
	return values
}

// stringArg returns the compile-time string value for one supported AST expression.
func stringArg(expr ast.Expr, stringConsts map[string]string) (string, error) {
	switch value := expr.(type) {
	case *ast.BasicLit:
		if value.Kind != token.STRING {
			return "", fmt.Errorf("expected string literal")
		}
		unquoted, err := strconv.Unquote(value.Value)
		if err != nil {
			return "", fmt.Errorf("invalid string literal: %w", err)
		}
		return unquoted, nil
	case *ast.Ident:
		if resolved, ok := stringConsts[value.Name]; ok {
			return resolved, nil
		}
		return "", fmt.Errorf("unknown string const %s", value.Name)
	case *ast.BinaryExpr:
		if value.Op != token.ADD {
			return "", fmt.Errorf("only string concatenation is supported")
		}
		left, err := stringArg(value.X, stringConsts)
		if err != nil {
			return "", err
		}
		right, err := stringArg(value.Y, stringConsts)
		if err != nil {
			return "", err
		}
		return left + right, nil
	default:
		return "", fmt.Errorf("expected string literal or const")
	}
}

// routeGroupBindingDir converts one backend/api-relative package path into an
// absolute API package directory.
func routeGroupBindingDir(apiDir string, apiPackage string) (string, error) {
	if hasParentPathSegment(apiPackage) {
		return "", fmt.Errorf("api package must stay under backend/api")
	}
	normalizedPackage := normalizeAPIPackagePath(apiPackage)
	if normalizedPackage == "" || normalizedPackage == "." {
		return filepath.Clean(apiDir), nil
	}
	if strings.HasPrefix(normalizedPackage, "../") || strings.Contains(normalizedPackage, "/../") {
		return "", fmt.Errorf("api package must stay under backend/api")
	}
	return filepath.Clean(filepath.Join(apiDir, filepath.FromSlash(normalizedPackage))), nil
}

// hasParentPathSegment reports whether a package path attempts parent traversal.
func hasParentPathSegment(value string) bool {
	for _, segment := range strings.Split(filepath.ToSlash(strings.TrimSpace(value)), "/") {
		if segment == ".." {
			return true
		}
	}
	return false
}

// normalizeAPIPackagePath canonicalizes a backend/api-relative package path.
func normalizeAPIPackagePath(value string) string {
	normalized := filepath.ToSlash(strings.TrimSpace(value))
	normalized = strings.TrimPrefix(normalized, "./")
	normalized = strings.Trim(normalized, "/")
	if normalized == "" {
		return "."
	}
	return filepath.ToSlash(filepath.Clean(normalized))
}

// routeContractTagKeys lists GoFrame route tags consumed by the dynamic route contract.
var routeContractTagKeys = map[string]struct{}{
	"path":        {},
	"method":      {},
	"operationId": {},
	"tags":        {},
	"summary":     {},
	"dc":          {},
	"description": {},
	"access":      {},
	"permission":  {},
}

func extractRouteContractsFromFile(fileNode *ast.File) ([]*protocol.RouteContract, error) {
	items := make([]*protocol.RouteContract, 0)
	for _, decl := range fileNode.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok || structType.Fields == nil {
				continue
			}
			for _, field := range structType.Fields.List {
				if field == nil || field.Tag == nil {
					continue
				}
				if len(field.Names) != 0 {
					continue
				}
				tagValue := strings.Trim(field.Tag.Value, "`")
				if strings.TrimSpace(tagValue) == "" {
					continue
				}
				metaValues := parseStructTagValues(tagValue)
				if metaValues["path"] == "" || metaValues["method"] == "" {
					continue
				}
				contract := &protocol.RouteContract{
					Path:        metaValues["path"],
					Method:      metaValues["method"],
					Tags:        splitTagList(metaValues["tags"]),
					Summary:     metaValues["summary"],
					Description: metaValues["dc"],
					Access:      metaValues["access"],
					Permission:  metaValues["permission"],
					Meta:        buildRouteContractMeta(metaValues),
					RequestType: strings.TrimSpace(typeSpec.Name.Name),
				}
				items = append(items, contract)
			}
		}
	}
	return items, nil
}

// routeGroupPrefixForDir returns the closest registered route group prefix
// declared at or above the DTO package directory.
func routeGroupPrefixForDir(apiDir string, sourceDir string, prefixes map[string]string) string {
	for current := filepath.Clean(sourceDir); ; current = filepath.Dir(current) {
		if prefix, ok := prefixes[current]; ok {
			return prefix
		}
		if current == filepath.Clean(apiDir) {
			break
		}
		next := filepath.Dir(current)
		if next == current {
			break
		}
	}
	return ""
}

// applyRouteGroupPrefix composes the registered group prefix with one DTO route
// path, mirroring source-plugin `Group(prefix).Bind(controller)` routing.
func applyRouteGroupPrefix(prefix string, contract *protocol.RouteContract) {
	if contract == nil {
		return
	}
	contract.Path = joinRouteDeclarationPaths(prefix, contract.Path)
}

// joinRouteDeclarationPaths combines normalized route declaration paths.
func joinRouteDeclarationPaths(prefix string, routePath string) string {
	normalizedPrefix := normalizeRouteDeclarationPath(prefix)
	normalizedRoutePath := normalizeRouteDeclarationPath(routePath)
	if normalizedPrefix == "/" {
		return normalizedRoutePath
	}
	if normalizedRoutePath == "/" {
		return normalizedPrefix
	}
	return normalizedPrefix + normalizedRoutePath
}

// normalizeRouteDeclarationPath makes one route declaration path absolute and
// stable without interpreting plugin-owned path segments.
func normalizeRouteDeclarationPath(value string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" || normalized == "/" {
		return "/"
	}
	if !strings.HasPrefix(normalized, "/") {
		normalized = "/" + normalized
	}
	return strings.TrimRight(normalized, "/")
}

// buildRouteContractMeta preserves plugin-defined route metadata without host interpretation.
func buildRouteContractMeta(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	meta := make(map[string]string)
	for key, value := range values {
		normalizedKey := strings.TrimSpace(key)
		normalizedValue := strings.TrimSpace(value)
		if normalizedKey == "" || normalizedValue == "" {
			continue
		}
		if _, reserved := routeContractTagKeys[normalizedKey]; reserved {
			continue
		}
		meta[normalizedKey] = normalizedValue
	}
	if len(meta) == 0 {
		return nil
	}
	return meta
}

func parseStructTagValues(tagValue string) map[string]string {
	values := make(map[string]string)
	cursor := 0
	for cursor < len(tagValue) {
		for cursor < len(tagValue) && tagValue[cursor] == ' ' {
			cursor++
		}
		if cursor >= len(tagValue) {
			break
		}
		keyStart := cursor
		for cursor < len(tagValue) && tagValue[cursor] != ':' {
			cursor++
		}
		if cursor >= len(tagValue) || tagValue[cursor] != ':' {
			break
		}
		key := strings.TrimSpace(tagValue[keyStart:cursor])
		cursor++
		if cursor >= len(tagValue) || tagValue[cursor] != '"' {
			break
		}
		cursor++
		valueStart := cursor
		for cursor < len(tagValue) {
			if tagValue[cursor] == '"' && tagValue[cursor-1] != '\\' {
				break
			}
			cursor++
		}
		if cursor >= len(tagValue) {
			break
		}
		values[key] = tagValue[valueStart:cursor]
		cursor++
	}
	return values
}

func splitTagList(value string) []string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return nil
	}
	items := strings.Split(normalized, ",")
	result := make([]string, 0, len(items))
	for _, item := range items {
		tag := strings.TrimSpace(item)
		if tag == "" {
			continue
		}
		result = append(result, tag)
	}
	return result
}

func validateHookSpec(pluginID string, spec *hookSpec, filePath string) error {
	if spec == nil {
		return fmt.Errorf("plugin hook cannot be nil: %s", filePath)
	}
	if strings.TrimSpace(string(spec.Event)) == "" {
		return fmt.Errorf("plugin hook missing event: %s", filePath)
	}
	if !isHookExtensionPoint(spec.Event) {
		return fmt.Errorf("plugin hook event is not published by host: %s", filePath)
	}
	if spec.Action == "" {
		spec.Action = hookActionInsert
	}
	if !isSupportedHookAction(spec.Action) {
		return fmt.Errorf("plugin hook action is not supported: %s", filePath)
	}
	if spec.Mode == "" {
		spec.Mode = defaultCallbackExecutionMode(spec.Event)
	}
	if !isExtensionPointExecutionModeSupported(spec.Event, spec.Mode) {
		return fmt.Errorf("plugin hook execution mode is not supported: %s", filePath)
	}
	if spec.TimeoutMs < 0 {
		return fmt.Errorf("plugin hook timeoutMs cannot be negative: %s", filePath)
	}

	switch spec.Action {
	case hookActionInsert:
		if err := validateIdentifier(spec.Table); err != nil {
			return fmt.Errorf("plugin %s hook table is invalid: %s: %w", pluginID, filePath, err)
		}
		if len(spec.Fields) == 0 {
			return fmt.Errorf("plugin hook missing fields: %s", filePath)
		}
		for column := range spec.Fields {
			if err := validateIdentifier(column); err != nil {
				return fmt.Errorf("plugin %s hook field is invalid: %s: %w", pluginID, filePath, err)
			}
		}
	case hookActionSleep:
		if spec.SleepMs <= 0 {
			return fmt.Errorf("plugin hook sleep action requires sleepMs > 0: %s", filePath)
		}
	case hookActionError:
		if strings.TrimSpace(spec.ErrorMessage) == "" {
			return fmt.Errorf("plugin hook error action requires non-empty errorMessage: %s", filePath)
		}
	}

	return nil
}

func validateResourceSpec(pluginID string, spec *resourceSpec, filePath string) error {
	if spec == nil {
		return fmt.Errorf("plugin resource cannot be nil: %s", filePath)
	}
	if strings.TrimSpace(spec.Key) == "" {
		return fmt.Errorf("plugin resource missing key: %s", filePath)
	}
	if spec.Type == "" {
		spec.Type = string(resourceSpecTypeTableList)
	}
	if normalizeResourceSpecType(spec.Type) != resourceSpecTypeTableList {
		return fmt.Errorf("plugin resource type only supports table-list: %s", filePath)
	}
	if err := validateIdentifier(spec.Table); err != nil {
		return fmt.Errorf("plugin %s resource table is invalid: %s: %w", pluginID, filePath, err)
	}
	if len(spec.Fields) == 0 {
		return fmt.Errorf("plugin resource missing fields: %s", filePath)
	}
	for _, field := range spec.Fields {
		if field == nil {
			return fmt.Errorf("plugin resource field cannot be nil: %s", filePath)
		}
		if err := validateIdentifier(field.Name); err != nil {
			return fmt.Errorf("plugin %s resource field name is invalid: %s: %w", pluginID, filePath, err)
		}
		if err := validateIdentifier(field.Column); err != nil {
			return fmt.Errorf("plugin %s resource column is invalid: %s: %w", pluginID, filePath, err)
		}
	}
	for _, filter := range spec.Filters {
		if filter == nil {
			return fmt.Errorf("plugin resource filter cannot be nil: %s", filePath)
		}
		if strings.TrimSpace(filter.Param) == "" {
			return fmt.Errorf("plugin resource filter missing param: %s", filePath)
		}
		if err := validateIdentifier(filter.Column); err != nil {
			return fmt.Errorf("plugin %s resource filter column is invalid: %s: %w", pluginID, filePath, err)
		}
		if normalizeResourceFilterOperator(filter.Operator) == "" {
			return fmt.Errorf("plugin resource filter operator is not supported: %s", filePath)
		}
	}
	if err := validateIdentifier(spec.OrderBy.Column); err != nil {
		return fmt.Errorf("plugin %s resource orderBy column is invalid: %s: %w", pluginID, filePath, err)
	}
	if spec.OrderBy.Direction == "" {
		spec.OrderBy.Direction = string(resourceOrderDirectionASC)
	}
	if normalizeResourceOrderDirection(spec.OrderBy.Direction) == "" {
		return fmt.Errorf("plugin resource order direction only supports asc/desc: %s", filePath)
	}
	if spec.DataScope != nil {
		if spec.DataScope.UserColumn != "" {
			if err := validateIdentifier(spec.DataScope.UserColumn); err != nil {
				return fmt.Errorf("plugin %s resource dataScope userColumn is invalid: %s: %w", pluginID, filePath, err)
			}
		}
		if spec.DataScope.DeptColumn != "" {
			if err := validateIdentifier(spec.DataScope.DeptColumn); err != nil {
				return fmt.Errorf("plugin %s resource dataScope deptColumn is invalid: %s: %w", pluginID, filePath, err)
			}
		}
		if spec.DataScope.UserColumn == "" && spec.DataScope.DeptColumn == "" {
			return fmt.Errorf("plugin resource dataScope requires userColumn or deptColumn: %s", filePath)
		}
	}
	if len(spec.Operations) == 0 {
		spec.Operations = []string{string(resourceOperationQuery)}
	}
	operationSeen := make(map[string]struct{}, len(spec.Operations))
	for _, operation := range spec.Operations {
		normalizedOperation := normalizeResourceOperation(operation)
		if normalizedOperation == "" {
			return fmt.Errorf("plugin resource operation is not supported: %s", filePath)
		}
		operationSeen[string(normalizedOperation)] = struct{}{}
	}
	spec.Operations = normalizeResourceEnumStringSlice(spec.Operations)

	if spec.KeyField != "" {
		if err := validateIdentifier(spec.KeyField); err != nil {
			return fmt.Errorf("plugin %s resource keyField is invalid: %s: %w", pluginID, filePath, err)
		}
		if !resourceSpecHasField(spec, spec.KeyField) {
			return fmt.Errorf("plugin resource keyField is not declared in fields: %s", filePath)
		}
	}
	if _, ok := operationSeen[string(resourceOperationGet)]; ok && strings.TrimSpace(spec.KeyField) == "" {
		return fmt.Errorf("plugin resource get operation requires keyField: %s", filePath)
	}
	if _, ok := operationSeen[string(resourceOperationUpdate)]; ok && strings.TrimSpace(spec.KeyField) == "" {
		return fmt.Errorf("plugin resource update operation requires keyField: %s", filePath)
	}
	if _, ok := operationSeen[string(resourceOperationDelete)]; ok && strings.TrimSpace(spec.KeyField) == "" {
		return fmt.Errorf("plugin resource delete operation requires keyField: %s", filePath)
	}

	if len(spec.WritableFields) > 0 {
		spec.WritableFields = normalizeResourceFieldNameSlice(spec.WritableFields)
		for _, writableField := range spec.WritableFields {
			if err := validateIdentifier(writableField); err != nil {
				return fmt.Errorf("plugin %s resource writableField is invalid: %s: %w", pluginID, filePath, err)
			}
			if !resourceSpecHasField(spec, writableField) {
				return fmt.Errorf("plugin resource writableField is not declared in fields: %s", filePath)
			}
		}
	}
	if _, ok := operationSeen[string(resourceOperationCreate)]; ok && len(spec.WritableFields) == 0 {
		return fmt.Errorf("plugin resource create operation requires writableFields: %s", filePath)
	}
	if _, ok := operationSeen[string(resourceOperationUpdate)]; ok && len(spec.WritableFields) == 0 {
		return fmt.Errorf("plugin resource update operation requires writableFields: %s", filePath)
	}

	if spec.Access == "" {
		spec.Access = string(resourceAccessModeRequest)
	}
	if normalizeResourceAccessMode(spec.Access) == "" {
		return fmt.Errorf("plugin resource access is not supported: %s", filePath)
	}
	spec.Access = strings.ToLower(strings.TrimSpace(spec.Access))
	return nil
}

func validateIdentifier(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("identifier cannot be empty")
	}
	if !safeIdentifierPattern.MatchString(value) {
		return fmt.Errorf("identifier is invalid: %s", value)
	}
	return nil
}

func defaultCallbackExecutionMode(point hookExtensionPoint) callbackExecutionMode {
	return publishedHookPoints[point]
}

func isHookExtensionPoint(point hookExtensionPoint) bool {
	_, ok := publishedHookPoints[point]
	return ok
}

func isSupportedHookAction(action hookAction) bool {
	switch action {
	case hookActionInsert, hookActionSleep, hookActionError:
		return true
	default:
		return false
	}
}

func isExtensionPointExecutionModeSupported(point hookExtensionPoint, mode callbackExecutionMode) bool {
	modes, ok := supportedHookModes[point]
	if !ok {
		return false
	}
	_, ok = modes[mode]
	return ok
}

func normalizeResourceSpecType(value string) resourceSpecType {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(resourceSpecTypeTableList):
		return resourceSpecTypeTableList
	default:
		return ""
	}
}

func normalizeResourceFilterOperator(value string) resourceFilterOperator {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(resourceFilterOperatorEQ):
		return resourceFilterOperatorEQ
	case string(resourceFilterOperatorLike):
		return resourceFilterOperatorLike
	case string(resourceFilterOperatorGTEDate):
		return resourceFilterOperatorGTEDate
	case string(resourceFilterOperatorLTEDate):
		return resourceFilterOperatorLTEDate
	default:
		return ""
	}
}

func normalizeResourceOrderDirection(value string) resourceOrderDirection {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(resourceOrderDirectionASC):
		return resourceOrderDirectionASC
	case string(resourceOrderDirectionDESC):
		return resourceOrderDirectionDESC
	default:
		return ""
	}
}

func normalizeResourceOperation(value string) resourceOperation {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(resourceOperationQuery):
		return resourceOperationQuery
	case string(resourceOperationGet):
		return resourceOperationGet
	case string(resourceOperationCreate):
		return resourceOperationCreate
	case string(resourceOperationUpdate):
		return resourceOperationUpdate
	case string(resourceOperationDelete):
		return resourceOperationDelete
	case string(resourceOperationTransaction):
		return resourceOperationTransaction
	default:
		return ""
	}
}

func normalizeResourceAccessMode(value string) resourceAccessMode {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", string(resourceAccessModeRequest):
		return resourceAccessModeRequest
	case string(resourceAccessModeSystem):
		return resourceAccessModeSystem
	case string(resourceAccessModeBoth):
		return resourceAccessModeBoth
	default:
		return ""
	}
}

func resourceSpecHasField(spec *resourceSpec, fieldName string) bool {
	if spec == nil {
		return false
	}
	targetFieldName := strings.TrimSpace(fieldName)
	if targetFieldName == "" {
		return false
	}
	for _, field := range spec.Fields {
		if field != nil && field.Name == targetFieldName {
			return true
		}
	}
	return false
}

func normalizeResourceEnumStringSlice(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		normalized := strings.ToLower(strings.TrimSpace(item))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	sort.Strings(result)
	return result
}

func normalizeResourceFieldNameSlice(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		lookupKey := strings.ToLower(trimmed)
		if _, ok := seen[lookupKey]; ok {
			continue
		}
		seen[lookupKey] = struct{}{}
		result = append(result, trimmed)
	}
	sort.Strings(result)
	return result
}

func sortStrings(items []string) {
	if len(items) <= 1 {
		return
	}
	sort.Strings(items)
}
