// This file verifies host-service descriptor governance across protocol,
// dynamic bridge SDK, non-WASI stubs, and host dispatcher synchronization
// points.

package hostservice

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"lina-core/pkg/plugin/pluginbridge/protocol/hostservices"
)

// TestHostServiceDescriptorsCoverProtocolGuestAndDispatcher verifies the
// descriptor table remains synchronized with public aliases, bridge SDK clients,
// and host-side dispatchers for all published host-service methods.
func TestHostServiceDescriptorsCoverProtocolGuestAndDispatcher(t *testing.T) {
	root := repoRootForDescriptorTest(t)
	protocolDir := filepath.Join(root, "pkg/plugin/pluginbridge/protocol")
	guestDirs := []string{
		filepath.Join(root, "pkg/plugin/pluginbridge"),
		filepath.Join(root, "pkg/plugin/pluginbridge/internal/domainhostcall"),
		filepath.Join(root, "pkg/plugin/pluginbridge/recordstore"),
	}
	wasmDir := filepath.Join(root, "internal/service/plugin/internal/wasm")

	var (
		protocolConsts = declaredConstNames(t, protocolDir)
		protocolTypes  = declaredTypeNames(t, protocolDir)
		protocolValues = declaredValueNames(t, protocolDir)
	)
	for name := range declaredFuncNames(t, protocolDir) {
		protocolValues[name] = struct{}{}
	}
	guestSelectors := selectorNames(t, guestDirs...)
	dispatcherSelectors := selectorNames(t, wasmDir)
	expectedGuestSelectors := descriptorMethodConstSet(func(descriptor hostServiceMethodDescriptor) bool {
		return descriptor.Published && descriptor.GuestClient
	})
	expectedDispatcherSelectors := descriptorMethodConstSet(func(descriptor hostServiceMethodDescriptor) bool {
		return descriptor.Published && descriptor.Dispatcher
	})

	for _, descriptor := range hostServiceMethodDescriptors() {
		if !descriptor.Published {
			continue
		}
		key := descriptor.Service + "." + descriptor.Method
		if descriptor.MethodConst == "" {
			t.Fatalf("published host service method %s is missing MethodConst", key)
		}
		if _, ok := protocolConsts[descriptor.MethodConst]; !ok {
			t.Fatalf("published host service method %s is missing public protocol const alias %s", key, descriptor.MethodConst)
		}
		assertPayloadAliases(t, protocolTypes, protocolValues, key, descriptor.RequestPayload)
		assertPayloadAliases(t, protocolTypes, protocolValues, key, descriptor.ResponsePayload)
		if descriptor.GuestClient {
			if _, ok := guestSelectors[descriptor.MethodConst]; !ok {
				t.Fatalf("published host service method %s is missing guest client usage of %s", key, descriptor.MethodConst)
			}
		}
		if descriptor.Dispatcher {
			if _, ok := dispatcherSelectors[descriptor.MethodConst]; !ok {
				t.Fatalf("published host service method %s is missing wasm dispatcher usage of %s", key, descriptor.MethodConst)
			}
		}
	}
	assertNoUnexpectedSelectors(t, "guest client", guestSelectors, expectedGuestSelectors)
	assertNoUnexpectedSelectors(t, "wasm dispatcher", dispatcherSelectors, expectedDispatcherSelectors)
	assertHostServiceEntryHasNoServiceSwitch(t, wasmDir)
	assertDispatcherFunctionsMatchDescriptors(t, wasmDir)
}

// TestProtocolHostServiceCodecsOwnPayloadImplementation verifies public
// protocol codec files contain the payload implementation instead of re-aliasing
// internal hostservice codec owners.
func TestProtocolHostServiceCodecsOwnPayloadImplementation(t *testing.T) {
	root := repoRootForDescriptorTest(t)
	protocolDir := filepath.Join(root, "pkg/plugin/pluginbridge/protocol")
	for _, filePath := range productionGoFilesInDir(t, protocolDir) {
		if !strings.HasPrefix(filepath.Base(filePath), "protocol_hostservice_") ||
			!strings.Contains(filepath.Base(filePath), "_codec.go") {
			continue
		}
		content := string(readFileForDescriptorTest(t, filePath))
		if strings.Contains(content, "pluginbridge/internal/hostservice") {
			t.Fatalf("protocol host service codec must own payload implementation, but %s imports internal hostservice", filePath)
		}
		if strings.Contains(content, "= hostservice.MarshalHostService") ||
			strings.Contains(content, "= hostservice.UnmarshalHostService") {
			t.Fatalf("protocol host service codec must not alias internal codec functions: %s", filePath)
		}
	}
}

// TestHostServiceDescriptorsHaveNoPerDomainGuestStubs verifies guest clients
// use the shared injected transport instead of per-domain WASI singletons or
// non-WASI mirror stubs. RecordStore keeps its separate executor files because
// they implement query-plan execution rather than a mirrored host-service client.
func TestHostServiceDescriptorsHaveNoPerDomainGuestStubs(t *testing.T) {
	var (
		root          = repoRootForDescriptorTest(t)
		dataStubFuncs = declaredFuncNamesForBuildTag(t, filepath.Join(root, "pkg/plugin/pluginbridge/recordstore"), "!wasip1")
		guestDir      = filepath.Join(root, "pkg/plugin/pluginbridge")
	)
	for _, filePath := range productionGoFilesInDir(t, guestDir) {
		name := filepath.Base(filePath)
		if strings.HasPrefix(name, "pluginbridge_hostcall_") &&
			strings.HasSuffix(name, "_wasip1.go") &&
			name != "pluginbridge_hostcall_wasip1.go" {
			t.Fatalf("pluginbridge root contains per-domain WASI host-service client residual: %s", filePath)
		}
		if strings.HasSuffix(name, "_adapter.go") {
			t.Fatalf("pluginbridge root contains adapter residual: %s", filePath)
		}
		content := string(readFileForDescriptorTest(t, filePath))
		for _, fragment := range []string{
			"unsupportedRuntimeHostService",
			"unsupportedStorageHostService",
			"unsupportedNetworkHostService",
			"unsupportedCacheHostService",
			"unsupportedLockHostService",
			"unsupportedHostConfigHostService",
			"unsupportedManifestHostService",
			"defaultRuntimeHostService",
			"defaultStorageHostService",
			"defaultNetworkHostService",
			"defaultCacheHostService",
			"defaultLockHostService",
			"defaultHostConfigHostService",
			"defaultManifestHostService",
		} {
			if strings.Contains(content, fragment) {
				t.Fatalf("pluginbridge root contains non-WASI mirror stub residual %q in %s", fragment, filePath)
			}
		}
		if name == "pluginbridge_hostcall_stub.go" {
			stubFuncs := declaredFuncNamesFromFiles(t, parseGoFilesForBuildTag(t, guestDir, "!wasip1"))
			if _, ok := stubFuncs["InvokeHostService"]; !ok {
				t.Fatalf("pluginbridge non-WASI transport stub is missing InvokeHostService")
			}
		}
	}
	for _, fn := range []string{"One", "All", "Count", "Insert", "Update", "Delete", "Transaction"} {
		if _, ok := dataStubFuncs[fn]; !ok {
			t.Fatalf("data host service is missing non-WASI data stub method %s", fn)
		}
	}
}

// TestHostServiceDescriptorCapabilityMetadata verifies capability and resource
// metadata are derived from the descriptor table without duplicate service
// method entries.
func TestHostServiceDescriptorCapabilityMetadata(t *testing.T) {
	seen := make(map[string]struct{})
	for _, descriptor := range hostServiceMethodDescriptors() {
		key := descriptor.Service + "." + descriptor.Method
		if descriptor.Service == "" || descriptor.Method == "" || descriptor.Capability == "" {
			t.Fatalf("host service descriptor has incomplete identity: %#v", descriptor)
		}
		if _, ok := seen[key]; ok {
			t.Fatalf("host service descriptor is duplicated: %s", key)
		}
		seen[key] = struct{}{}
		if !descriptor.Published {
			if got := RequiredCapabilityForHostServiceMethod(descriptor.Service, descriptor.Method); got != "" {
				t.Fatalf("reserved descriptor %s must not be declared at runtime, got capability %s", key, got)
			}
			continue
		}
		if got := RequiredCapabilityForHostServiceMethod(descriptor.Service, descriptor.Method); got != descriptor.Capability {
			t.Fatalf("descriptor %s expects capability %s, got %s", key, descriptor.Capability, got)
		}
	}
}

// TestHostServiceDescriptorsUsePublicCatalog verifies internal descriptor
// governance is derived from the public protocol host-service catalog.
func TestHostServiceDescriptorsUsePublicCatalog(t *testing.T) {
	if !reflect.DeepEqual(hostServiceDescriptors(), hostservices.Catalog()) {
		t.Fatal("internal host service descriptors must be derived from protocol/hostservices catalog")
	}
	if !reflect.DeepEqual(hostServiceMethodDescriptors(), hostservices.Methods()) {
		t.Fatal("internal host service method descriptors must be derived from protocol/hostservices catalog")
	}
}

// TestHostServiceReadmeGeneratedBlocksMatchCatalog verifies the public README
// host-service tables are kept in sync with the catalog-owned descriptor table.
func TestHostServiceReadmeGeneratedBlocksMatchCatalog(t *testing.T) {
	root := repoRootForDescriptorTest(t)
	readmeDir := filepath.Join(root, "pkg/plugin")
	for _, tc := range []struct {
		file          string
		noneLabel     string
		header        string
		notifications string
	}{
		{
			file:          "README.md",
			noneLabel:     "None",
			header:        "| Service | Resource declaration | Derived capability | Methods |",
			notifications: "None except `messages.send`, which uses `resources[].ref`",
		},
		{
			file:          "README.zh-CN.md",
			noneLabel:     "无",
			header:        "| Service | 资源声明 | 派生能力 | Methods |",
			notifications: "除`messages.send`使用`resources[].ref`外无需资源声明",
		},
	} {
		t.Run(tc.file, func(t *testing.T) {
			var (
				path     = filepath.Join(readmeDir, tc.file)
				actual   = extractGeneratedHostServicesBlock(t, path)
				expected = renderReadmeHostServicesBlock(tc.noneLabel, tc.header, tc.notifications)
			)
			if actual != expected {
				t.Fatalf("%s generated host-services block is stale\nexpected:\n%s\nactual:\n%s", path, expected, actual)
			}
		})
	}
}

// TestWASMHostServiceDoesNotImportInternalHostservice verifies the host runtime
// does not cross the pluginbridge internal package boundary for catalog data.
func TestWASMHostServiceDoesNotImportInternalHostservice(t *testing.T) {
	root := repoRootForDescriptorTest(t)
	wasmDir := filepath.Join(root, "internal/service/plugin/internal/wasm")
	for _, filePath := range productionGoFilesRecursive(t, wasmDir) {
		content := string(readFileForDescriptorTest(t, filePath))
		if strings.Contains(content, "pluginbridge/internal/hostservice") {
			t.Fatalf("wasm host service must use public protocol/hostservices catalog, but %s imports internal hostservice", filePath)
		}
	}
}

// TestDomainCapabilityBoundaryGovernance verifies ordinary domain capabilities
// keep one contract owner, one Wasm configuration entry, and one guest-domain
// proxy location.
func TestDomainCapabilityBoundaryGovernance(t *testing.T) {
	root := repoRootForDescriptorTest(t)
	productionDirs := []string{
		filepath.Join(root, "internal/cmd"),
		filepath.Join(root, "internal/service/plugin"),
		filepath.Join(root, "pkg/plugin/pluginbridge"),
	}
	forbiddenFragments := []string{
		"ConfigureAITextHostService",
		"ConfigureUserHostService",
		"ConfigureOrgHostService",
		"ConfigureTenantHostService",
		"aiTextHostServices",
		"userHostServices",
		"orgHostServices",
		"tenantHostServices",
		"internal/service/plugin/internal/hostservices",
	}
	for _, filePath := range productionGoFilesRecursive(t, productionDirs...) {
		content := string(readFileForDescriptorTest(t, filePath))
		for _, fragment := range forbiddenFragments {
			if strings.Contains(content, fragment) {
				t.Fatalf("domain capability boundary regression: %s contains %q", filePath, fragment)
			}
		}
	}

	cmdDir := filepath.Join(root, "internal/cmd")
	for _, filePath := range productionGoFilesRecursive(t, cmdDir) {
		content := string(readFileForDescriptorTest(t, filePath))
		if strings.Contains(content, "internal/service/plugin/internal/capabilityhost") {
			t.Fatalf("startup layer must use plugin facade, but %s imports capabilityhost directly", filePath)
		}
	}

	guestDir := filepath.Join(root, "pkg/plugin/pluginbridge")
	for _, filePath := range productionGoFilesInDir(t, guestDir) {
		content := string(readFileForDescriptorTest(t, filePath))
		for _, typeName := range []string{
			"type AIService interface",
			"type AITextService interface",
			"type AIImageService interface",
			"type AIEmbeddingService interface",
			"type AIAudioService interface",
			"type AIVisionService interface",
			"type AIDocumentService interface",
			"type AISafetyService interface",
			"type AIVideoService interface",
			"type PluginService interface",
		} {
			if strings.Contains(content, typeName) {
				t.Fatalf("pluginbridge public package must return aicap contracts instead of parallel AI interfaces: %s contains %q", filePath, typeName)
			}
		}
	}

	directoryPath := filepath.Join(guestDir, "pluginbridge_directory.go")
	directoryContent := string(readFileForDescriptorTest(t, directoryPath))
	for _, fragment := range []string{
		"protocol.HostService",
		"MarshalHostService",
		"UnmarshalHostService",
		"callJSONRequest",
		"callHostService",
	} {
		if strings.Contains(directoryContent, fragment) {
			t.Fatalf("pluginbridge_directory.go must only inject invokers and select typed clients, but contains %q", fragment)
		}
	}

	for _, descriptor := range hostServiceDescriptors() {
		capDir := capabilityContractDirForHostService(root, descriptor.Service)
		if capDir == "" {
			continue
		}
		if _, err := os.Stat(capDir); err != nil {
			t.Fatalf("ordinary host service %s must keep capability contract owner under %s: %v", descriptor.Service, capDir, err)
		}
	}
}

func capabilityContractDirForHostService(root string, service string) string {
	capabilityRoot := filepath.Join(root, "pkg/plugin/capability")
	switch service {
	case hostservices.HostServiceRuntime,
		hostservices.HostServiceStorage,
		hostservices.HostServiceNetwork,
		hostservices.HostServiceData,
		hostservices.HostServiceCache,
		hostservices.HostServiceLock,
		hostservices.HostServiceHostConfig,
		hostservices.HostServiceManifest:
		return ""
	case hostservices.HostServiceAPIDoc:
		return filepath.Join(capabilityRoot, "apidoccap")
	case hostservices.HostServiceAuth:
		return filepath.Join(capabilityRoot, "authcap")
	case hostservices.HostServiceUsers:
		return filepath.Join(capabilityRoot, "usercap")
	case hostservices.HostServiceBizCtx:
		return filepath.Join(capabilityRoot, "bizctxcap")
	case hostservices.HostServiceFiles:
		return filepath.Join(capabilityRoot, "filecap")
	case hostservices.HostServiceJobs:
		return filepath.Join(capabilityRoot, "jobcap")
	case hostservices.HostServiceNotifications:
		return filepath.Join(capabilityRoot, "notifycap")
	case hostservices.HostServicePlugins:
		return filepath.Join(capabilityRoot, "plugincap")
	case hostservices.HostServiceSessions:
		return filepath.Join(capabilityRoot, "sessioncap")
	default:
		return filepath.Join(capabilityRoot, service+"cap")
	}
}

func assertPayloadAliases(
	t *testing.T,
	protocolTypes map[string]struct{},
	protocolValues map[string]struct{},
	methodKey string,
	payload string,
) {
	t.Helper()
	if payload == "" {
		return
	}
	if _, ok := protocolTypes[payload]; !ok {
		t.Fatalf("host service method %s is missing public protocol payload type %s", methodKey, payload)
	}
	for _, fn := range []string{"Marshal" + payload, "Unmarshal" + payload} {
		if _, ok := protocolValues[fn]; !ok {
			t.Fatalf("host service method %s is missing public protocol codec %s", methodKey, fn)
		}
	}
}

func descriptorMethodConstSet(match func(hostServiceMethodDescriptor) bool) map[string]struct{} {
	result := make(map[string]struct{})
	for _, descriptor := range hostServiceMethodDescriptors() {
		if descriptor.MethodConst == "" || !match(descriptor) {
			continue
		}
		result[descriptor.MethodConst] = struct{}{}
	}
	return result
}

func descriptorDispatcherServices() map[string]struct{} {
	result := make(map[string]struct{})
	for _, descriptor := range hostServiceMethodDescriptors() {
		if descriptor.Dispatcher {
			result[descriptor.Service] = struct{}{}
		}
	}
	return result
}

func assertNoUnexpectedSelectors(
	t *testing.T,
	label string,
	actual map[string]struct{},
	expected map[string]struct{},
) {
	t.Helper()
	for selector := range actual {
		if _, ok := expected[selector]; !ok {
			t.Fatalf("%s contains host service method selector %s that is not declared by descriptor", label, selector)
		}
	}
}

func assertHostServiceEntryHasNoServiceSwitch(t *testing.T, wasmDir string) {
	t.Helper()
	actual := hostServiceSwitchSelectors(t, filepath.Join(wasmDir, "wasm_host_service.go"))
	for constName := range actual {
		t.Fatalf("wasm_host_service.go must use registry dispatch instead of service-level switch case %s", constName)
	}
}

func assertDispatcherFunctionsMatchDescriptors(t *testing.T, wasmDir string) {
	t.Helper()
	actual := dispatchFunctionNames(t, wasmDir)
	expected := make(map[string]string)
	for service := range descriptorDispatcherServices() {
		expected[dispatcherFunctionNameForService(t, service)] = service
	}
	for name, service := range expected {
		if _, ok := actual[name]; !ok {
			t.Fatalf("wasm dispatcher service %s is missing dispatcher function %s", service, name)
		}
	}
	for name := range actual {
		if _, ok := expected[name]; !ok {
			t.Fatalf("wasm dispatcher function %s is not declared by descriptor service set", name)
		}
	}
}

func repoRootForDescriptorTest(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("resolve test working directory failed: %v", err)
	}
	for {
		if _, err = os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate apps/lina-core go.mod from descriptor test")
		}
		dir = parent
	}
}

func productionGoFilesRecursive(t *testing.T, dirs ...string) []string {
	t.Helper()
	files := make([]string, 0)
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read Go source directory %s failed: %v", dir, err)
		}
		for _, entry := range entries {
			path := filepath.Join(dir, entry.Name())
			if entry.IsDir() {
				files = append(files, productionGoFilesRecursive(t, path)...)
				continue
			}
			if isProductionGoFile(entry.Name()) {
				files = append(files, path)
			}
		}
	}
	return files
}

func productionGoFilesInDir(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read Go source directory %s failed: %v", dir, err)
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if isProductionGoFile(entry.Name()) {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	return files
}

func isProductionGoFile(name string) bool {
	return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
}

func readFileForDescriptorTest(t *testing.T, filePath string) []byte {
	t.Helper()
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read Go source %s failed: %v", filePath, err)
	}
	return content
}

func declaredConstNames(t *testing.T, dir string) map[string]struct{} {
	t.Helper()
	result := make(map[string]struct{})
	for _, file := range parseGoFiles(t, dir) {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.CONST {
				continue
			}
			for _, spec := range gen.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range valueSpec.Names {
					result[name.Name] = struct{}{}
				}
			}
		}
	}
	return result
}

func declaredTypeNames(t *testing.T, dir string) map[string]struct{} {
	t.Helper()
	result := make(map[string]struct{})
	for _, file := range parseGoFiles(t, dir) {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if ok {
					result[typeSpec.Name.Name] = struct{}{}
				}
			}
		}
	}
	return result
}

func declaredValueNames(t *testing.T, dir string) map[string]struct{} {
	t.Helper()
	result := make(map[string]struct{})
	for _, file := range parseGoFiles(t, dir) {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}
			for _, spec := range gen.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range valueSpec.Names {
					result[name.Name] = struct{}{}
				}
			}
		}
	}
	return result
}

func declaredFuncNames(t *testing.T, dir string) map[string]struct{} {
	t.Helper()
	return declaredFuncNamesFromFiles(t, parseGoFiles(t, dir))
}

func declaredFuncNamesForBuildTag(t *testing.T, dir string, buildTag string) map[string]struct{} {
	t.Helper()
	return declaredFuncNamesFromFiles(t, parseGoFilesForBuildTag(t, dir, buildTag))
}

func declaredFuncNamesFromFiles(t *testing.T, files []*ast.File) map[string]struct{} {
	t.Helper()
	result := make(map[string]struct{})
	for _, file := range files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if ok {
				result[fn.Name.Name] = struct{}{}
			}
		}
	}
	return result
}

func selectorNames(t *testing.T, dirs ...string) map[string]struct{} {
	t.Helper()
	result := make(map[string]struct{})
	for _, dir := range dirs {
		for _, file := range parseGoFiles(t, dir) {
			ast.Inspect(file, func(node ast.Node) bool {
				selector, ok := node.(*ast.SelectorExpr)
				if ok && strings.HasPrefix(selector.Sel.Name, "HostServiceMethod") {
					result[selector.Sel.Name] = struct{}{}
				}
				return true
			})
		}
	}
	return result
}

func hostServiceSwitchSelectors(t *testing.T, filePath string) map[string]struct{} {
	t.Helper()
	content := readFileForDescriptorTest(t, filePath)
	file, err := parser.ParseFile(token.NewFileSet(), filePath, content, 0)
	if err != nil {
		t.Fatalf("parse Go source %s failed: %v", filePath, err)
	}
	result := make(map[string]struct{})
	ast.Inspect(file, func(node ast.Node) bool {
		caseClause, ok := node.(*ast.CaseClause)
		if !ok {
			return true
		}
		for _, expr := range caseClause.List {
			selector, ok := expr.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			if strings.HasPrefix(selector.Sel.Name, "HostService") &&
				!strings.HasPrefix(selector.Sel.Name, "HostServiceMethod") {
				result[selector.Sel.Name] = struct{}{}
			}
		}
		return true
	})
	return result
}

func dispatchFunctionNames(t *testing.T, dir string) map[string]struct{} {
	t.Helper()
	result := make(map[string]struct{})
	for _, file := range parseGoFiles(t, dir) {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			name := fn.Name.Name
			if name == "dispatchRegisteredHostService" || name == "dispatchOwnerHostService" {
				continue
			}
			if strings.HasPrefix(name, "dispatch") &&
				(strings.HasSuffix(name, "HostService") || name == "dispatchHostConfigService") {
				result[name] = struct{}{}
			}
		}
	}
	return result
}

func hostServiceConstNameForService(t *testing.T, service string) string {
	t.Helper()
	switch service {
	case hostservices.HostServiceAPIDoc:
		return "hostservices.HostServiceAPIDoc"
	case hostservices.HostServiceBizCtx:
		return "hostservices.HostServiceBizCtx"
	case hostservices.HostServiceHostConfig:
		return "hostservices.HostServiceHostConfig"
	default:
		parts := strings.Split(service, "_")
		for i, part := range parts {
			if part == "" {
				continue
			}
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
		return "HostService" + strings.Join(parts, "")
	}
}

func dispatcherFunctionNameForService(t *testing.T, service string) string {
	t.Helper()
	switch service {
	case hostservices.HostServiceAPIDoc:
		return "dispatchAPIDocHostService"
	case hostservices.HostServiceBizCtx:
		return "dispatchBizCtxHostService"
	case hostservices.HostServiceHostConfig:
		return "dispatchHostConfigService"
	default:
		constName := hostServiceConstNameForService(t, service)
		return "dispatch" + strings.TrimPrefix(constName, "HostService") + "HostService"
	}
}

func extractGeneratedHostServicesBlock(t *testing.T, path string) string {
	t.Helper()
	var (
		content     = string(readFileForDescriptorTest(t, path))
		startMarker = "<!-- BEGIN generated:host-services -->"
		endMarker   = "<!-- END generated:host-services -->"
		start       = strings.Index(content, startMarker)
	)
	if start < 0 {
		t.Fatalf("%s is missing %s", path, startMarker)
	}
	end := strings.Index(content[start:], endMarker)
	if end < 0 {
		t.Fatalf("%s is missing %s", path, endMarker)
	}
	return strings.TrimSpace(content[start : start+end+len(endMarker)])
}

func renderReadmeHostServicesBlock(noneLabel string, header string, notificationsResource string) string {
	var builder strings.Builder
	builder.WriteString("<!-- BEGIN generated:host-services -->\n")
	builder.WriteString(header)
	builder.WriteString("\n")
	builder.WriteString("| --- | --- | --- | --- |\n")
	for _, descriptor := range hostServiceDescriptors() {
		var (
			methods          = make([]string, 0, len(descriptor.Methods))
			capabilities     = make([]string, 0, len(descriptor.Methods))
			seenCapabilities = make(map[string]struct{})
		)
		for _, method := range descriptor.Methods {
			methodName := "`" + method.Method + "`"
			if !method.Published {
				methodName += " reserved"
			}
			methods = append(methods, methodName)
			if _, ok := seenCapabilities[method.Capability]; ok {
				continue
			}
			seenCapabilities[method.Capability] = struct{}{}
			capabilities = append(capabilities, "`"+method.Capability+"`")
		}
		resource := readmeResourceDeclaration(descriptor, noneLabel, notificationsResource)
		builder.WriteString(fmt.Sprintf(
			"| `%s` | %s | %s | %s |\n",
			descriptor.Service,
			resource,
			strings.Join(capabilities, "<br/>"),
			strings.Join(methods, "<br/>"),
		))
	}
	builder.WriteString("<!-- END generated:host-services -->")
	return strings.TrimSpace(builder.String())
}

func readmeResourceDeclaration(
	descriptor hostServiceDescriptor,
	noneLabel string,
	notificationsResource string,
) string {
	if descriptor.Service == hostservices.HostServiceNotifications {
		return notificationsResource
	}
	switch descriptor.ResourceKind {
	case hostServiceResourceNone:
		return noneLabel
	case hostServiceResourcePath:
		return "`resources.paths`"
	case hostServiceResourceTable:
		return "`resources.tables`"
	case hostServiceResourceKey:
		return "`resources.keys`"
	case hostServiceResourceRef:
		if descriptor.Service == hostservices.HostServiceNetwork {
			return "`resources[].url`"
		}
		return "`resources[].ref`"
	default:
		return "`" + string(descriptor.ResourceKind) + "`"
	}
}

func parseGoFiles(t *testing.T, dir string) []*ast.File {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read Go source directory %s failed: %v", dir, err)
	}
	files := make([]*ast.File, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		filePath := filepath.Join(dir, name)
		file, err := parser.ParseFile(token.NewFileSet(), filePath, nil, 0)
		if err != nil {
			t.Fatalf("parse Go source %s failed: %v", filePath, err)
		}
		files = append(files, file)
	}
	return files
}

func parseGoFilesForBuildTag(t *testing.T, dir string, buildTag string) []*ast.File {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read Go source directory %s failed: %v", dir, err)
	}
	files := make([]*ast.File, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		filePath := filepath.Join(dir, name)
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("read Go source %s failed: %v", filePath, err)
		}
		if !strings.Contains(string(content), "//go:build "+buildTag) {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), filePath, content, 0)
		if err != nil {
			t.Fatalf("parse Go source %s failed: %v", filePath, err)
		}
		files = append(files, file)
	}
	return files
}
