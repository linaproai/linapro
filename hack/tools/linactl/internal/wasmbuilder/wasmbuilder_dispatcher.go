// This file generates zero-reflection guest dispatchers for dynamic Wasm
// plugin runtime builds. The generated source is written into the plugin
// backend package only for the duration of `go build` and is removed
// afterward, keeping author-owned plugin source free of generated artifacts.

package wasmbuilder

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"lina-core/pkg/plugin/pluginbridge/protocol"
)

const (
	// generatedDispatcherFileName is the temporary backend package file
	// injected during dynamic guest runtime builds.
	generatedDispatcherFileName = "zz_generated_wasm_dispatcher.go"
)

// prepareGeneratedWasmDispatcher writes the temporary generated dispatcher and
// returns a cleanup function. It returns a no-op cleanup when the plugin does
// not expose enough typed controller metadata to generate a dispatcher.
func prepareGeneratedWasmDispatcher(
	pluginDir string,
	pluginID string,
	routes []*routeContractSource,
	lifecycleSpecs []*protocol.LifecycleContract,
) (func() error, error) {
	spec, err := buildWasmDispatcherSpec(pluginDir, pluginID, routes, lifecycleSpecs)
	if err != nil {
		return nil, err
	}
	if spec == nil {
		return func() error { return nil }, nil
	}
	content, err := renderWasmDispatcher(spec)
	if err != nil {
		return nil, err
	}

	generatedPath := filepath.Join(pluginDir, "backend", generatedDispatcherFileName)
	if err = os.WriteFile(generatedPath, content, 0o644); err != nil {
		return nil, fmt.Errorf("failed to write generated wasm dispatcher: %w", err)
	}
	return func() error {
		if err := os.Remove(generatedPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}, nil
}

// buildWasmDispatcherSpec derives one generated dispatcher plan from API DTO
// route contracts, backend API interfaces, and lifecycle contracts.
func buildWasmDispatcherSpec(
	pluginDir string,
	pluginID string,
	routeSources []*routeContractSource,
	lifecycleSpecs []*protocol.LifecycleContract,
) (*wasmDispatcherSpec, error) {
	modulePath, err := readGoModulePath(pluginDir)
	if err != nil {
		return nil, err
	}
	apiControllers, routeHandlers, err := buildWasmAPIDispatcherSpecs(pluginDir, modulePath, routeSources)
	if err != nil {
		return nil, err
	}
	dtoPackages, err := collectWasmDTOPackages(pluginDir)
	if err != nil {
		return nil, err
	}
	lifecycleHandlers := buildWasmLifecycleDispatcherSpecs(lifecycleSpecs, apiControllers, dtoPackages)
	cronHandlers, err := discoverWasmCallbackDispatcherSpecs(
		pluginDir,
		apiControllers,
		routeHandlers,
		lifecycleHandlers,
		dtoPackages,
	)
	if err != nil {
		return nil, err
	}
	if len(apiControllers) == 0 {
		return nil, nil
	}
	if len(apiControllers) == 0 && len(routeHandlers) == 0 && len(lifecycleHandlers) == 0 && len(cronHandlers) == 0 {
		return nil, nil
	}
	return &wasmDispatcherSpec{
		PluginID:        strings.TrimSpace(pluginID),
		APIControllers:  apiControllers,
		Routes:          routeHandlers,
		LifecycleRoutes: lifecycleHandlers,
		CronRoutes:      cronHandlers,
	}, nil
}

// buildWasmAPIDispatcherSpecs maps DTO request types to typed controller
// methods by reading backend/api interface declarations.
func buildWasmAPIDispatcherSpecs(
	pluginDir string,
	modulePath string,
	routeSources []*routeContractSource,
) ([]*wasmAPIControllerSpec, []*wasmRouteHandlerSpec, error) {
	if len(routeSources) == 0 {
		return nil, nil, nil
	}

	interfaces, err := collectWasmAPIInterfaces(pluginDir)
	if err != nil {
		return nil, nil, err
	}
	dtoFields, err := collectWasmDTOFields(pluginDir)
	if err != nil {
		return nil, nil, err
	}
	controllerByPackage := make(map[string]*wasmAPIControllerSpec)
	handlers := make([]*wasmRouteHandlerSpec, 0)
	for _, source := range routeSources {
		if source == nil || len(source.contracts) == 0 {
			continue
		}
		apiPackage := normalizeAPIPackagePath(source.apiPackage)
		iface, ok := resolveWasmAPIInterface(interfaces, apiPackage)
		if !ok {
			continue
		}
		controller := controllerByPackage[apiPackage]
		if controller == nil {
			controller = buildWasmAPIControllerSpec(modulePath, iface.apiPackage, len(controllerByPackage)+1, iface.name)
			controllerByPackage[apiPackage] = controller
		}
		for _, contract := range source.contracts {
			if contract == nil {
				continue
			}
			requestType := strings.TrimSpace(contract.RequestType)
			methodName := strings.TrimSpace(iface.methods[requestType])
			if requestType == "" || methodName == "" {
				continue
			}
			handlers = append(handlers, &wasmRouteHandlerSpec{
				RequestType:     requestType,
				Method:          strings.ToUpper(strings.TrimSpace(contract.Method)),
				Path:            normalizeRouteDeclarationPath(contract.Path),
				APIPackage:      apiPackage,
				ControllerAlias: controller.ImportAlias,
				ControllerType:  controller.InterfaceName,
				MethodName:      methodName,
				DTOImportAlias:  buildWasmDTOImportAlias(apiPackage),
				RequestTypeExpr: buildWasmDTOTypeExpr(apiPackage, requestType),
				Fields:          dtoFields[buildWasmDTOFieldKey(apiPackage, requestType)],
			})
		}
	}

	controllers := make([]*wasmAPIControllerSpec, 0, len(controllerByPackage))
	for _, controller := range controllerByPackage {
		controllers = append(controllers, controller)
	}
	sort.Slice(controllers, func(left int, right int) bool {
		return controllers[left].PackagePath < controllers[right].PackagePath
	})
	sort.Slice(handlers, func(left int, right int) bool {
		if handlers[left].RequestType == handlers[right].RequestType {
			return handlers[left].Path < handlers[right].Path
		}
		return handlers[left].RequestType < handlers[right].RequestType
	})
	return controllers, handlers, nil
}

// collectWasmAPIInterfaces extracts request DTO to method-name mappings from
// backend/api interface declarations.
func collectWasmAPIInterfaces(pluginDir string) (map[string]*wasmAPIInterfaceSpec, error) {
	apiDir := filepath.Join(pluginDir, "backend", "api")
	result := make(map[string]*wasmAPIInterfaceSpec)
	err := filepath.WalkDir(apiDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fileNode, parseErr := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
		if parseErr != nil {
			return fmt.Errorf("failed to parse backend api interface file %s: %w", path, parseErr)
		}
		apiPackage, relErr := backendAPIPackageForDir(apiDir, filepath.Dir(path))
		if relErr != nil {
			return relErr
		}
		for _, decl := range fileNode.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || typeSpec.Name == nil {
					continue
				}
				interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
				if !ok || interfaceType.Methods == nil {
					continue
				}
				item := &wasmAPIInterfaceSpec{
					name:       typeSpec.Name.Name,
					apiPackage: apiPackage,
					methods:    make(map[string]string),
				}
				for _, field := range interfaceType.Methods.List {
					methodName, requestType := extractWasmAPIInterfaceMethod(field)
					if methodName == "" || requestType == "" {
						continue
					}
					item.methods[requestType] = methodName
				}
				if len(item.methods) > 0 {
					result[apiPackage] = item
				}
			}
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, err
	}
	return result, nil
}

// resolveWasmAPIInterface finds the closest interface package at or above the
// DTO package, matching the local backend/api/<group>/<version> convention.
func resolveWasmAPIInterface(
	interfaces map[string]*wasmAPIInterfaceSpec,
	apiPackage string,
) (*wasmAPIInterfaceSpec, bool) {
	for current := normalizeAPIPackagePath(apiPackage); ; current = filepath.ToSlash(filepath.Dir(current)) {
		if item, ok := interfaces[current]; ok {
			return item, true
		}
		if current == "." || current == "" {
			break
		}
		next := filepath.ToSlash(filepath.Dir(current))
		if next == current {
			break
		}
	}
	return nil, false
}

// wasmAPIInterfaceSpec stores the parsed request-to-method mapping for one
// backend/api interface package.
type wasmAPIInterfaceSpec struct {
	name       string
	apiPackage string
	methods    map[string]string
}

// collectWasmDTOFields extracts exported JSON-tagged request fields so the
// generated dispatcher can decide whether a body is required and hydrate
// supported path/query parameters without reflection.
func collectWasmDTOFields(pluginDir string) (map[string][]*wasmDTOFieldSpec, error) {
	apiDir := filepath.Join(pluginDir, "backend", "api")
	result := make(map[string][]*wasmDTOFieldSpec)
	err := filepath.WalkDir(apiDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fileNode, parseErr := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
		if parseErr != nil {
			return fmt.Errorf("failed to parse backend api dto file %s: %w", path, parseErr)
		}
		apiPackage, relErr := backendAPIPackageForDir(apiDir, filepath.Dir(path))
		if relErr != nil {
			return relErr
		}
		for _, decl := range fileNode.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || typeSpec.Name == nil {
					continue
				}
				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok || structType.Fields == nil {
					continue
				}
				fields := extractWasmDTOBindableFields(structType)
				if len(fields) > 0 {
					result[buildWasmDTOFieldKey(apiPackage, typeSpec.Name.Name)] = fields
				}
			}
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, err
	}
	return result, nil
}

// collectWasmDTOPackages maps request DTO type names to backend/api-relative
// packages so non-route callbacks can still use typed request contracts.
func collectWasmDTOPackages(pluginDir string) (map[string]string, error) {
	apiDir := filepath.Join(pluginDir, "backend", "api")
	result := make(map[string]string)
	err := filepath.WalkDir(apiDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fileNode, parseErr := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
		if parseErr != nil {
			return fmt.Errorf("failed to parse backend api dto file %s: %w", path, parseErr)
		}
		apiPackage, relErr := backendAPIPackageForDir(apiDir, filepath.Dir(path))
		if relErr != nil {
			return relErr
		}
		for _, decl := range fileNode.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || typeSpec.Name == nil {
					continue
				}
				requestType := strings.TrimSpace(typeSpec.Name.Name)
				if strings.HasSuffix(requestType, "Req") {
					result[requestType] = apiPackage
				}
			}
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, err
	}
	return result, nil
}

// extractWasmDTOBindableFields returns JSON-tagged fields that may require a
// request body or can be populated from path/query values by generated code.
func extractWasmDTOBindableFields(structType *ast.StructType) []*wasmDTOFieldSpec {
	fields := make([]*wasmDTOFieldSpec, 0)
	for _, field := range structType.Fields.List {
		if field == nil || field.Tag == nil || len(field.Names) == 0 {
			continue
		}
		jsonName := wasmDTOJSONFieldName(field)
		if jsonName == "" {
			continue
		}
		required := wasmDTOFieldRequired(field)
		for _, name := range field.Names {
			if name == nil || !name.IsExported() {
				continue
			}
			fields = append(fields, &wasmDTOFieldSpec{
				GoName:   name.Name,
				JSONName: jsonName,
				GoType:   wasmDTOFieldGoType(field.Type),
				Required: required,
			})
		}
	}
	return fields
}

// wasmDTOJSONFieldName extracts the JSON name used by route path/query values.
func wasmDTOJSONFieldName(field *ast.Field) string {
	tagValue := strings.Trim(field.Tag.Value, "`")
	jsonTag := strings.TrimSpace(parseStructTagValues(tagValue)["json"])
	if jsonTag == "" || jsonTag == "-" {
		return ""
	}
	if index := strings.Index(jsonTag, ","); index >= 0 {
		jsonTag = jsonTag[:index]
	}
	return strings.TrimSpace(jsonTag)
}

// wasmDTOFieldRequired reports whether one DTO field requires a request body
// when no route path or query value supplies it.
func wasmDTOFieldRequired(field *ast.Field) bool {
	if field == nil || field.Tag == nil {
		return false
	}
	tagValue := strings.Trim(field.Tag.Value, "`")
	validationTag := strings.TrimSpace(parseStructTagValues(tagValue)["v"])
	for _, rule := range strings.Split(validationTag, "|") {
		ruleName := strings.TrimSpace(rule)
		if index := strings.Index(ruleName, ":"); index >= 0 {
			ruleName = ruleName[:index]
		}
		if ruleName == "required" {
			return true
		}
	}
	return false
}

// wasmDTOFieldGoType converts supported AST field types to generator binding
// type names used for path/query value assignment.
func wasmDTOFieldGoType(expr ast.Expr) string {
	switch value := expr.(type) {
	case *ast.Ident:
		switch value.Name {
		case "string", "bool",
			"int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64":
			return value.Name
		default:
			return ""
		}
	default:
		return ""
	}
}

// buildWasmDTOFieldKey builds the lookup key for one request DTO.
func buildWasmDTOFieldKey(apiPackage string, requestType string) string {
	return normalizeAPIPackagePath(apiPackage) + ":" + strings.TrimSpace(requestType)
}

// extractWasmAPIInterfaceMethod returns the method name and request DTO type
// for a GoFrame-style typed controller signature.
func extractWasmAPIInterfaceMethod(field *ast.Field) (string, string) {
	if field == nil || len(field.Names) == 0 {
		return "", ""
	}
	methodName := strings.TrimSpace(field.Names[0].Name)
	if methodName == "" {
		return "", ""
	}
	funcType, ok := field.Type.(*ast.FuncType)
	if !ok {
		return "", ""
	}
	params := flattenFieldTypes(funcType.Params)
	results := flattenFieldTypes(funcType.Results)
	if len(params) != 2 || len(results) != 2 || !isContextType(params[0]) || !isErrorType(results[1]) {
		return "", ""
	}
	if !isPointerToNamedType(params[1]) || !isPointerToNamedType(results[0]) {
		return "", ""
	}
	return methodName, pointerTypeName(params[1])
}

// buildWasmAPIControllerSpec creates one generated controller binding for an
// API package.
func buildWasmAPIControllerSpec(
	modulePath string,
	apiPackage string,
	index int,
	interfaceName string,
) *wasmAPIControllerSpec {
	importAlias := fmt.Sprintf("controller%d", index)
	controllerPackage := strings.TrimSpace(modulePath) + "/backend"
	if normalizedPackage := normalizeAPIPackagePath(apiPackage); normalizedPackage != "." {
		controllerPackage += "/internal/controller/" + normalizedPackage
	}
	interfaceAlias := fmt.Sprintf("api%d", index)
	interfacePackage := strings.TrimSpace(modulePath) + "/backend/api"
	if normalizedPackage := normalizeAPIPackagePath(apiPackage); normalizedPackage != "." {
		interfacePackage += "/" + normalizedPackage
	}
	return &wasmAPIControllerSpec{
		APIPackage:        normalizeAPIPackagePath(apiPackage),
		ImportAlias:       importAlias,
		PackagePath:       controllerPackage,
		InterfaceAlias:    interfaceAlias,
		InterfacePath:     interfacePackage,
		Constructor:       importAlias + ".New()",
		ConcreteType:      "*" + importAlias + ".Controller",
		InterfaceName:     interfaceName,
		InterfaceTypeExpr: interfaceAlias + "." + interfaceName,
	}
}

// buildWasmDTOTypeExpr returns the generated Go expression for one request DTO.
func buildWasmDTOTypeExpr(apiPackage string, requestType string) string {
	return buildWasmDTOImportAlias(apiPackage) + "." + strings.TrimSpace(requestType)
}

// buildWasmDTOImportAlias returns the import alias for one DTO package.
func buildWasmDTOImportAlias(apiPackage string) string {
	var (
		normalizedPackage = normalizeAPIPackagePath(apiPackage)
		segments          = strings.Split(normalizedPackage, "/")
		alias             = "dto"
	)
	if len(segments) > 0 {
		alias = segments[len(segments)-1]
	}
	if alias == "" || alias == "." {
		alias = "dto"
	}
	return "dto" + sanitizeIdentifierTitle(alias)
}

// sanitizeIdentifierTitle normalizes one package path segment into an exported
// identifier suffix for generated import aliases.
func sanitizeIdentifierTitle(value string) string {
	var builder strings.Builder
	upperNext := true
	for _, r := range strings.TrimSpace(value) {
		switch {
		case r >= 'a' && r <= 'z':
			if upperNext {
				r -= 'a' - 'A'
			}
			builder.WriteRune(r)
			upperNext = false
		case r >= 'A' && r <= 'Z':
			builder.WriteRune(r)
			upperNext = false
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
			upperNext = false
		default:
			upperNext = true
		}
	}
	if builder.Len() == 0 {
		return "DTO"
	}
	return builder.String()
}

// buildWasmLifecycleDispatcherSpecs converts lifecycle contracts to generated
// typed dispatch entries.
func buildWasmLifecycleDispatcherSpecs(
	lifecycleSpecs []*protocol.LifecycleContract,
	apiControllers []*wasmAPIControllerSpec,
	dtoPackages map[string]string,
) []*wasmLifecycleHandlerSpec {
	items := make([]*wasmLifecycleHandlerSpec, 0, len(lifecycleSpecs))
	for _, spec := range lifecycleSpecs {
		if spec == nil {
			continue
		}
		methodName := strings.TrimSpace(spec.Operation.String())
		requestType := strings.TrimSpace(spec.RequestType)
		if methodName == "" || requestType == "" {
			continue
		}
		controller, apiPackage, ok := resolveWasmControllerForRequestType(apiControllers, dtoPackages, requestType)
		if !ok {
			continue
		}
		items = append(items, &wasmLifecycleHandlerSpec{
			RequestType:     requestType,
			MethodName:      methodName,
			APIPackage:      apiPackage,
			ControllerAlias: controller.ImportAlias,
			DTOImportAlias:  buildWasmDTOImportAlias(apiPackage),
			RequestTypeExpr: buildWasmDTOTypeExpr(apiPackage, requestType),
		})
	}
	sort.Slice(items, func(left int, right int) bool {
		return items[left].RequestType < items[right].RequestType
	})
	return items
}

// discoverWasmCallbackDispatcherSpecs discovers typed callback handlers that
// are not represented by public API routes or lifecycle contracts.
func discoverWasmCallbackDispatcherSpecs(
	pluginDir string,
	apiControllers []*wasmAPIControllerSpec,
	routeHandlers []*wasmRouteHandlerSpec,
	lifecycleHandlers []*wasmLifecycleHandlerSpec,
	dtoPackages map[string]string,
) ([]*wasmCallbackHandlerSpec, error) {
	var (
		backendDir        = filepath.Join(pluginDir, "backend")
		controllerDir     = filepath.Join(backendDir, "internal", "controller")
		knownRequestTypes = make(map[string]struct{})
	)
	for _, route := range routeHandlers {
		if route != nil {
			knownRequestTypes[route.RequestType] = struct{}{}
		}
	}
	for _, route := range lifecycleHandlers {
		if route != nil {
			knownRequestTypes[route.RequestType] = struct{}{}
		}
	}
	items := make([]*wasmCallbackHandlerSpec, 0)
	err := filepath.WalkDir(controllerDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fileNode, parseErr := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
		if parseErr != nil {
			return fmt.Errorf("failed to parse backend controller file %s: %w", path, parseErr)
		}
		for _, decl := range fileNode.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl == nil || funcDecl.Recv == nil || funcDecl.Name == nil {
				continue
			}
			methodName, requestType, ok := extractTypedControllerMethod(funcDecl)
			if !ok {
				continue
			}
			if _, exists := knownRequestTypes[requestType]; exists {
				continue
			}
			controller, apiPackage, ok := resolveWasmControllerForRequestType(apiControllers, dtoPackages, requestType)
			if !ok {
				continue
			}
			if _, exists := knownRequestTypes[requestType]; exists {
				return fmt.Errorf("dynamic callback request type is duplicated: %s", requestType)
			}
			knownRequestTypes[requestType] = struct{}{}
			item := &wasmCallbackHandlerSpec{
				RequestType:     requestType,
				MethodName:      methodName,
				APIPackage:      apiPackage,
				ControllerAlias: controller.ImportAlias,
				DTOImportAlias:  buildWasmDTOImportAlias(apiPackage),
				RequestTypeExpr: buildWasmDTOTypeExpr(apiPackage, requestType),
			}
			items = append(items, item)
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return items, nil
		}
		return nil, err
	}
	sort.Slice(items, func(left int, right int) bool {
		return items[left].RequestType < items[right].RequestType
	})
	return items, nil
}

// extractTypedControllerMethod reports one exported typed controller method and
// its request DTO.
func extractTypedControllerMethod(decl *ast.FuncDecl) (string, string, bool) {
	if decl == nil || decl.Name == nil || !decl.Name.IsExported() {
		return "", "", false
	}
	params := flattenFieldTypes(decl.Type.Params)
	results := flattenFieldTypes(decl.Type.Results)
	if len(params) != 2 || len(results) != 2 || !isContextType(params[0]) || !isErrorType(results[1]) {
		return "", "", false
	}
	if !isPointerToNamedType(params[1]) || !isPointerToNamedType(results[0]) {
		return "", "", false
	}
	return strings.TrimSpace(decl.Name.Name), pointerTypeName(params[1]), true
}

func resolveWasmControllerForRequestType(
	apiControllers []*wasmAPIControllerSpec,
	dtoPackages map[string]string,
	requestType string,
) (*wasmAPIControllerSpec, string, bool) {
	apiPackage, ok := dtoPackages[strings.TrimSpace(requestType)]
	if !ok {
		return nil, "", false
	}
	normalizedPackage := normalizeAPIPackagePath(apiPackage)
	for _, controller := range apiControllers {
		if controller == nil {
			continue
		}
		if normalizeAPIPackagePath(controller.APIPackage) == normalizedPackage {
			return controller, normalizedPackage, true
		}
	}
	for _, controller := range apiControllers {
		if controller != nil {
			return controller, normalizedPackage, true
		}
	}
	return nil, "", false
}

// readGoModulePath returns the plugin module path used by generated imports.
func readGoModulePath(pluginDir string) (string, error) {
	content, err := os.ReadFile(filepath.Join(pluginDir, "go.mod"))
	if err != nil {
		if os.IsNotExist(err) {
			return "lina-plugin-runtime-guest", nil
		}
		return "", err
	}
	for _, line := range strings.Split(string(content), "\n") {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "module ") {
			modulePath := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "module "))
			if modulePath != "" {
				return modulePath, nil
			}
		}
	}
	return "", fmt.Errorf("go.mod in %s does not declare a module path", pluginDir)
}

// renderWasmDispatcher renders and formats one generated dispatcher source file.
func renderWasmDispatcher(spec *wasmDispatcherSpec) ([]byte, error) {
	var builder strings.Builder
	builder.WriteString("//go:build wasip1\n\n")
	builder.WriteString("// Code generated by linactl wasm; DO NOT EDIT.\n")
	builder.WriteString("// This temporary file is removed after the runtime Wasm build completes.\n\n")
	builder.WriteString("package backend\n\n")
	writeWasmDispatcherImports(&builder, spec)
	writeWasmDispatcherControllerVars(&builder, spec)
	writeWasmHandleRequest(&builder, spec)
	writeWasmDispatchByRequestType(&builder, spec)
	writeWasmHandlerFunctions(&builder, spec)
	writeWasmRouteHelpers(&builder, spec)

	formatted, err := format.Source([]byte(builder.String()))
	if err != nil {
		return nil, fmt.Errorf("format generated wasm dispatcher: %w\n%s", err, builder.String())
	}
	return formatted, nil
}

// writeWasmDispatcherImports renders imports required by the generated dispatcher.
func writeWasmDispatcherImports(builder *strings.Builder, spec *wasmDispatcherSpec) {
	imports := map[string]string{
		"protocol":     "lina-core/pkg/plugin/pluginbridge/protocol",
		"bridgeplugin": "lina-core/pkg/plugin/pluginbridge",
		"strconv":      "strconv",
		"strings":      "strings",
		"sync":         "sync",
	}
	for _, controller := range spec.APIControllers {
		imports[controller.ImportAlias] = controller.PackagePath
	}
	dtoImports := buildWasmDTOImports(spec)
	for alias, path := range dtoImports {
		imports[alias] = path
	}
	aliases := make([]string, 0, len(imports))
	for alias := range imports {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)

	builder.WriteString("import (\n")
	for _, alias := range aliases {
		path := imports[alias]
		if alias == filepath.Base(path) {
			builder.WriteString("\t")
		} else {
			builder.WriteString("\t" + alias + " ")
		}
		builder.WriteString(strconv.Quote(path) + "\n")
	}
	builder.WriteString(")\n\n")
}

// buildWasmDTOImports returns the DTO package imports referenced by typed route handlers.
func buildWasmDTOImports(spec *wasmDispatcherSpec) map[string]string {
	imports := make(map[string]string)
	modulePath := ""
	if len(spec.APIControllers) > 0 {
		apiPath := strings.TrimSpace(spec.APIControllers[0].InterfacePath)
		if index := strings.Index(apiPath, "/backend/api"); index >= 0 {
			modulePath = apiPath[:index]
		}
	}
	for _, route := range spec.Routes {
		if route.DTOImportAlias == "" {
			continue
		}
		if modulePath != "" {
			imports[route.DTOImportAlias] = modulePath + "/backend/api/" + normalizeAPIPackagePath(route.APIPackage)
		}
	}
	for _, route := range spec.LifecycleRoutes {
		if modulePath != "" && route.DTOImportAlias != "" {
			imports[route.DTOImportAlias] = modulePath + "/backend/api/" + normalizeAPIPackagePath(route.APIPackage)
		}
	}
	for _, route := range spec.CronRoutes {
		if modulePath != "" && route.DTOImportAlias != "" {
			imports[route.DTOImportAlias] = modulePath + "/backend/api/" + normalizeAPIPackagePath(route.APIPackage)
		}
	}
	return imports
}

// writeWasmDispatcherControllerVars renders package-level controller bindings.
func writeWasmDispatcherControllerVars(builder *strings.Builder, spec *wasmDispatcherSpec) {
	if len(spec.APIControllers) == 0 {
		return
	}
	builder.WriteString("var (\n")
	for _, controller := range spec.APIControllers {
		builder.WriteString(fmt.Sprintf(
			"\tgenerated%sOnce  sync.Once\n",
			upperFirst(controller.ImportAlias),
		))
		builder.WriteString(fmt.Sprintf(
			"\tgenerated%sValue %s\n",
			upperFirst(controller.ImportAlias),
			controller.ConcreteType,
		))
	}
	builder.WriteString(")\n\n")
	for _, controller := range spec.APIControllers {
		builder.WriteString(fmt.Sprintf(
			"func generated%s() %s {\n",
			upperFirst(controller.ImportAlias),
			controller.ConcreteType,
		))
		builder.WriteString(fmt.Sprintf("\tgenerated%sOnce.Do(func() {\n", upperFirst(controller.ImportAlias)))
		builder.WriteString(fmt.Sprintf("\t\tgenerated%sValue = %s\n", upperFirst(controller.ImportAlias), controller.Constructor))
		builder.WriteString("\t})\n")
		builder.WriteString(fmt.Sprintf("\treturn generated%sValue\n", upperFirst(controller.ImportAlias)))
		builder.WriteString("}\n\n")
	}
}

// writeWasmHandleRequest renders the public generated request entrypoint.
func writeWasmHandleRequest(builder *strings.Builder, spec *wasmDispatcherSpec) {
	builder.WriteString("func HandleRequest(request *protocol.BridgeRequestEnvelopeV1) (*protocol.BridgeResponseEnvelopeV1, error) {\n")
	builder.WriteString("\treturn handleGeneratedWasmBridgeRequest(request)\n")
	builder.WriteString("}\n\n")
	builder.WriteString("func handleGeneratedWasmBridgeRequest(request *protocol.BridgeRequestEnvelopeV1) (*protocol.BridgeResponseEnvelopeV1, error) {\n")
	builder.WriteString("\tif request == nil || request.Route == nil {\n")
	builder.WriteString("\t\treturn protocol.NewBadRequestResponse(\"Dynamic bridge request is missing route metadata\"), nil\n")
	builder.WriteString("\t}\n")
	builder.WriteString("\trequestType := strings.TrimSpace(request.Route.RequestType)\n")
	builder.WriteString("\tif response, err, ok := dispatchGeneratedWasmRequestType(requestType, request); ok {\n")
	builder.WriteString("\t\treturn response, err\n")
	builder.WriteString("\t}\n")
	builder.WriteString("\tif requestType == \"\" {\n")
	builder.WriteString("\t\treturn protocol.NewBadRequestResponse(\"Dynamic bridge request is missing route request type\"), nil\n")
	builder.WriteString("\t}\n")
	builder.WriteString("\treturn protocol.NewNotFoundResponse(\"Dynamic bridge route not found\"), nil\n")
	builder.WriteString("}\n\n")
}

// writeWasmDispatchByRequestType renders the primary requestType switch.
func writeWasmDispatchByRequestType(builder *strings.Builder, spec *wasmDispatcherSpec) {
	builder.WriteString("func dispatchGeneratedWasmRequestType(requestType string, request *protocol.BridgeRequestEnvelopeV1) (*protocol.BridgeResponseEnvelopeV1, error, bool) {\n")
	builder.WriteString("\tswitch strings.TrimSpace(requestType) {\n")
	for _, route := range spec.Routes {
		builder.WriteString(fmt.Sprintf("\tcase %s:\n", strconv.Quote(route.RequestType)))
		builder.WriteString(fmt.Sprintf("\t\tresponse, err := handleGenerated%s(request)\n", route.RequestType))
		builder.WriteString("\t\treturn response, err, true\n")
	}
	for _, route := range spec.LifecycleRoutes {
		builder.WriteString(fmt.Sprintf("\tcase %s:\n", strconv.Quote(route.RequestType)))
		builder.WriteString(fmt.Sprintf("\t\tresponse, err := handleGenerated%s(request)\n", route.RequestType))
		builder.WriteString("\t\treturn response, err, true\n")
	}
	for _, route := range spec.CronRoutes {
		builder.WriteString(fmt.Sprintf("\tcase %s:\n", strconv.Quote(route.RequestType)))
		builder.WriteString(fmt.Sprintf("\t\tresponse, err := handleGenerated%s(request)\n", route.RequestType))
		builder.WriteString("\t\treturn response, err, true\n")
	}
	builder.WriteString("\tdefault:\n")
	builder.WriteString(fmt.Sprintf("\t\treturn %s, %s, false\n", "nil", "nil"))
	builder.WriteString("\t}\n")
	builder.WriteString("}\n\n")
}

// generatedWasmRouteBodyFieldList renders the JSON field names that may still
// need a body when no matching path or query value is present on the envelope.
func generatedWasmRouteBodyFieldList(route *wasmRouteHandlerSpec) string {
	if route == nil || len(route.Fields) == 0 {
		return "nil"
	}
	names := make([]string, 0, len(route.Fields))
	seen := make(map[string]struct{}, len(route.Fields))
	for _, field := range route.Fields {
		if field == nil {
			continue
		}
		if !field.Required {
			continue
		}
		jsonName := strings.TrimSpace(field.JSONName)
		if jsonName == "" {
			continue
		}
		if _, ok := seen[jsonName]; ok {
			continue
		}
		seen[jsonName] = struct{}{}
		names = append(names, jsonName)
	}
	if len(names) == 0 {
		return "nil"
	}
	sort.Strings(names)
	quoted := make([]string, 0, len(names))
	for _, name := range names {
		quoted = append(quoted, strconv.Quote(name))
	}
	return "[]string{" + strings.Join(quoted, ", ") + "}"
}

// writeWasmHandlerFunctions renders typed route handlers.
func writeWasmHandlerFunctions(builder *strings.Builder, spec *wasmDispatcherSpec) {
	for _, route := range spec.Routes {
		builder.WriteString(fmt.Sprintf("func handleGenerated%s(request *protocol.BridgeRequestEnvelopeV1) (*protocol.BridgeResponseEnvelopeV1, error) {\n", route.RequestType))
		builder.WriteString("\tctx := bridgeplugin.NewGuestControllerContext(request)\n")
		builder.WriteString(fmt.Sprintf("\treq := &%s{}\n", route.RequestTypeExpr))
		builder.WriteString(fmt.Sprintf("\tif response := bindGeneratedWasmRequest(request, req, %s); response != nil {\n", generatedWasmRouteBodyFieldList(route)))
		builder.WriteString("\t\treturn response, nil\n")
		builder.WriteString("\t}\n")
		builder.WriteString(fmt.Sprintf("\tres, err := generated%s().%s(ctx, req)\n", upperFirst(route.ControllerAlias), route.MethodName))
		builder.WriteString("\tif response := bridgeplugin.ResponseFromError(err); response != nil {\n")
		builder.WriteString("\t\treturn response, nil\n")
		builder.WriteString("\t}\n")
		builder.WriteString("\tif err != nil {\n")
		builder.WriteString("\t\treturn nil, err\n")
		builder.WriteString("\t}\n")
		builder.WriteString("\treturn bridgeplugin.BuildGuestControllerResponse(ctx, res)\n")
		builder.WriteString("}\n\n")
	}
	for _, route := range spec.LifecycleRoutes {
		writeWasmCallbackHandlerFunction(builder, route.RequestType, route.RequestTypeExpr, route.ControllerAlias, route.MethodName)
	}
	for _, route := range spec.CronRoutes {
		writeWasmCallbackHandlerFunction(builder, route.RequestType, route.RequestTypeExpr, route.ControllerAlias, route.MethodName)
	}
	builder.WriteString("func bindGeneratedWasmRequest[T any](request *protocol.BridgeRequestEnvelopeV1, target *T, bodyFields []string) *protocol.BridgeResponseEnvelopeV1 {\n")
	builder.WriteString("\tif target == nil {\n")
	builder.WriteString("\t\treturn protocol.NewBadRequestResponse(\"Dynamic bridge request target is nil\")\n")
	builder.WriteString("\t}\n")
	builder.WriteString("\tif shouldBindGeneratedWasmJSONBody(request, bodyFields) {\n")
	builder.WriteString("\t\tbound, err := bridgeplugin.BindJSON[T](request)\n")
	builder.WriteString("\t\tif err != nil {\n")
	builder.WriteString("\t\t\tif response := bridgeplugin.ClassifyBindJSONError(err); response != nil {\n")
	builder.WriteString("\t\t\t\treturn response\n")
	builder.WriteString("\t\t\t}\n")
	builder.WriteString("\t\t\treturn protocol.NewBadRequestResponse(err.Error())\n")
	builder.WriteString("\t\t}\n")
	builder.WriteString("\t\t*target = *bound\n")
	builder.WriteString("\t}\n")
	builder.WriteString("\tapplyGeneratedWasmRouteValues(request, target)\n")
	builder.WriteString("\treturn nil\n")
	builder.WriteString("}\n\n")
}

// writeWasmCallbackHandlerFunction renders one typed non-route callback
// handler such as lifecycle or cron callbacks.
func writeWasmCallbackHandlerFunction(
	builder *strings.Builder,
	requestType string,
	requestTypeExpr string,
	controllerAlias string,
	methodName string,
) {
	builder.WriteString(fmt.Sprintf("func handleGenerated%s(request *protocol.BridgeRequestEnvelopeV1) (*protocol.BridgeResponseEnvelopeV1, error) {\n", requestType))
	builder.WriteString("\tctx := bridgeplugin.NewGuestControllerContext(request)\n")
	builder.WriteString(fmt.Sprintf("\treq := &%s{}\n", requestTypeExpr))
	builder.WriteString("\tif response := bindGeneratedWasmRequest(request, req, nil); response != nil {\n")
	builder.WriteString("\t\treturn response, nil\n")
	builder.WriteString("\t}\n")
	builder.WriteString(fmt.Sprintf("\tres, err := generated%s().%s(ctx, req)\n", upperFirst(controllerAlias), methodName))
	builder.WriteString("\tif response := bridgeplugin.ResponseFromError(err); response != nil {\n")
	builder.WriteString("\t\treturn response, nil\n")
	builder.WriteString("\t}\n")
	builder.WriteString("\tif err != nil {\n")
	builder.WriteString("\t\treturn nil, err\n")
	builder.WriteString("\t}\n")
	builder.WriteString("\treturn bridgeplugin.BuildGuestControllerResponse(ctx, res)\n")
	builder.WriteString("}\n\n")
}

// writeWasmRouteHelpers renders small non-reflective path, method, and binding
// helpers used by the generated dispatcher.
func writeWasmRouteHelpers(builder *strings.Builder, spec *wasmDispatcherSpec) {
	builder.WriteString(`func shouldBindGeneratedWasmJSONBody(request *protocol.BridgeRequestEnvelopeV1, bodyFields []string) bool {
	if request == nil {
		return false
	}
	if request.Request != nil && len(request.Request.Body) > 0 {
		return true
	}
	switch generatedWasmRequestMethod(request) {
	case "POST", "PUT", "PATCH":
	default:
		return false
	}
	if len(bodyFields) == 0 {
		return false
	}
	pathParams := generatedWasmPathParams(request)
	queryValues := generatedWasmQueryValues(request)
	for _, key := range bodyFields {
		if _, ok := pathParams[key]; ok {
			continue
		}
		if _, ok := queryValues[key]; ok {
			continue
		}
		return true
	}
	return false
}

func generatedWasmParseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true":
		return true
	case "0", "false", "":
		return false
	}
	return false
}

func applyGeneratedWasmRouteValues(targetRequest *protocol.BridgeRequestEnvelopeV1, target any) {
	switch req := target.(type) {
`)
	writeWasmRouteValueCases(builder, spec)
	builder.WriteString(`	}
}

func normalizeGeneratedWasmRoutePath(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	return strings.TrimRight(trimmed, "/")
}

func generatedWasmRequestMethod(request *protocol.BridgeRequestEnvelopeV1) string {
	if request == nil {
		return ""
	}
	if request.Route != nil {
		if method := strings.ToUpper(strings.TrimSpace(request.Route.Method)); method != "" {
			return method
		}
	}
	if request.Request != nil {
		return strings.ToUpper(strings.TrimSpace(request.Request.Method))
	}
	return ""
}

func generatedWasmRouteValue(request *protocol.BridgeRequestEnvelopeV1, key string) string {
	if value := bridgeplugin.PathParam(request, key); value != "" {
		return value
	}
	return bridgeplugin.QueryValue(request, key)
}

func generatedWasmPathParams(request *protocol.BridgeRequestEnvelopeV1) map[string]string {
	if request == nil || request.Route == nil {
		return nil
	}
	return request.Route.PathParams
}

func generatedWasmQueryValues(request *protocol.BridgeRequestEnvelopeV1) map[string][]string {
	if request == nil || request.Route == nil {
		return nil
	}
	return request.Route.QueryValues
}

func matchGeneratedWasmRoute(routePath string, actualPath string) bool {
	var (
		normalizedRoute = normalizeGeneratedWasmRoutePath(routePath)
		normalizedActual = normalizeGeneratedWasmRoutePath(actualPath)
		routeSegments = strings.Split(strings.TrimPrefix(normalizedRoute, "/"), "/")
		actualSegments = strings.Split(strings.TrimPrefix(normalizedActual, "/"), "/")
	)
	if normalizedRoute == "/" || normalizedRoute == "" {
		routeSegments = []string{}
	}
	if normalizedActual == "/" || normalizedActual == "" {
		actualSegments = []string{}
	}
	if len(routeSegments) != len(actualSegments) {
		return false
	}
	for index, routeSegment := range routeSegments {
		actualSegment := actualSegments[index]
		if strings.HasPrefix(routeSegment, "{") && strings.HasSuffix(routeSegment, "}") {
			if strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(routeSegment, "{"), "}")) == "" {
				return false
			}
			continue
		}
		if routeSegment != actualSegment {
			return false
		}
	}
	return true
}
`)
}

// writeWasmRouteValueCases renders request DTO path/query binding cases.
func writeWasmRouteValueCases(builder *strings.Builder, spec *wasmDispatcherSpec) {
	seen := make(map[string]struct{})
	wroteAny := false
	for _, route := range spec.Routes {
		if route == nil || len(route.Fields) == 0 {
			continue
		}
		if _, exists := seen[route.RequestTypeExpr]; exists {
			continue
		}
		seen[route.RequestTypeExpr] = struct{}{}
		builder.WriteString(fmt.Sprintf("\tcase *%s:\n", route.RequestTypeExpr))
		for _, field := range route.Fields {
			if field == nil || field.GoType == "" {
				continue
			}
			writeWasmRouteValueAssignment(builder, field)
		}
		wroteAny = true
	}
	if !wroteAny {
		builder.WriteString("\tdefault:\n")
		builder.WriteString("\t\treturn\n")
	}
}

// writeWasmRouteValueAssignment renders one field assignment from path or query values.
func writeWasmRouteValueAssignment(builder *strings.Builder, field *wasmDTOFieldSpec) {
	goName := strings.TrimSpace(field.GoName)
	jsonName := strings.TrimSpace(field.JSONName)
	if goName == "" || jsonName == "" {
		return
	}
	switch field.GoType {
	case "string":
		builder.WriteString(fmt.Sprintf("\t\tif value := bridgeplugin.PathParam(targetRequest, %s); value != \"\" {\n", strconv.Quote(jsonName)))
		builder.WriteString(fmt.Sprintf("\t\t\treq.%s = value\n", goName))
		builder.WriteString(fmt.Sprintf("\t\t} else if value := bridgeplugin.QueryValue(targetRequest, %s); value != \"\" {\n", strconv.Quote(jsonName)))
		builder.WriteString(fmt.Sprintf("\t\t\treq.%s = value\n", goName))
		builder.WriteString("\t\t}\n")
	case "bool":
		builder.WriteString(fmt.Sprintf("\t\tif value, ok := generatedWasmPathParams(targetRequest)[%s]; ok {\n", strconv.Quote(jsonName)))
		builder.WriteString(fmt.Sprintf("\t\t\treq.%s = generatedWasmParseBool(value)\n", goName))
		builder.WriteString(fmt.Sprintf("\t\t} else if values, ok := generatedWasmQueryValues(targetRequest)[%s]; ok && len(values) > 0 {\n", strconv.Quote(jsonName)))
		builder.WriteString(fmt.Sprintf("\t\t\treq.%s = generatedWasmParseBool(values[0])\n", goName))
		builder.WriteString("\t\t}\n")
	case "int", "int8", "int16", "int32", "int64":
		builder.WriteString(fmt.Sprintf("\t\tif value := generatedWasmRouteValue(targetRequest, %s); value != \"\" {\n", strconv.Quote(jsonName)))
		builder.WriteString("\t\t\tif parsed, err := strconv.ParseInt(value, 10, 64); err == nil {\n")
		builder.WriteString(fmt.Sprintf("\t\t\t\treq.%s = %s(parsed)\n", goName, field.GoType))
		builder.WriteString("\t\t\t}\n")
		builder.WriteString("\t\t}\n")
	case "uint", "uint8", "uint16", "uint32", "uint64":
		builder.WriteString(fmt.Sprintf("\t\tif value := generatedWasmRouteValue(targetRequest, %s); value != \"\" {\n", strconv.Quote(jsonName)))
		builder.WriteString("\t\t\tif parsed, err := strconv.ParseUint(value, 10, 64); err == nil {\n")
		builder.WriteString(fmt.Sprintf("\t\t\t\treq.%s = %s(parsed)\n", goName, field.GoType))
		builder.WriteString("\t\t\t}\n")
		builder.WriteString("\t\t}\n")
	}
}

// upperFirst returns value with the first ASCII letter uppercased.
func upperFirst(value string) string {
	if value == "" {
		return ""
	}
	return strings.ToUpper(value[:1]) + value[1:]
}
