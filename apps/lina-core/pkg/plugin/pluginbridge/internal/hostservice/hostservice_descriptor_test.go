// This file verifies host-service descriptor governance across protocol,
// guest SDK, non-WASI stubs, and host dispatcher synchronization points.

package hostservice

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHostServiceDescriptorsCoverProtocolGuestAndDispatcher verifies the
// descriptor table remains synchronized with public aliases, guest clients, and
// host-side dispatchers for all published host-service methods.
func TestHostServiceDescriptorsCoverProtocolGuestAndDispatcher(t *testing.T) {
	root := repoRootForDescriptorTest(t)
	protocolDir := filepath.Join(root, "pkg/plugin/pluginbridge/protocol")
	guestDirs := []string{
		filepath.Join(root, "pkg/plugin/pluginbridge/guest"),
		filepath.Join(root, "pkg/plugin/capability/recordstore"),
	}
	wasmDir := filepath.Join(root, "internal/service/plugin/internal/wasm")

	protocolConsts := declaredConstNames(t, protocolDir)
	protocolTypes := declaredTypeNames(t, protocolDir)
	protocolValues := declaredValueNames(t, protocolDir)
	guestSelectors := selectorNames(t, guestDirs...)
	dispatcherSelectors := selectorNames(t, wasmDir)

	for _, descriptor := range HostServiceMethodDescriptors() {
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
				t.Fatalf("published host service method %s is missing wasm dispatcher case for %s", key, descriptor.MethodConst)
			}
		}
	}
}

// TestHostServiceDescriptorsCoverNonWASIStubs verifies every WASI-only guest
// service family keeps a host-build unsupported stub.
func TestHostServiceDescriptorsCoverNonWASIStubs(t *testing.T) {
	root := repoRootForDescriptorTest(t)
	guestStubFuncs := declaredFuncNamesForBuildTag(t, filepath.Join(root, "pkg/plugin/pluginbridge/guest"), "!wasip1")
	dataStubFuncs := declaredFuncNamesForBuildTag(t, filepath.Join(root, "pkg/plugin/capability/recordstore"), "!wasip1")
	expectedGuestFactories := map[string]string{
		HostServiceRuntime:    "Runtime",
		HostServiceStorage:    "Storage",
		HostServiceNetwork:    "Network",
		HostServiceCache:      "Cache",
		HostServiceLock:       "Lock",
		HostServiceConfig:     "pluginConfig",
		HostServiceNotify:     "Notify",
		HostServiceCron:       "Cron",
		HostServiceHostConfig: "HostConfig",
		HostServiceManifest:   "Manifest",
	}
	for _, descriptor := range HostServiceDescriptors() {
		factory, ok := expectedGuestFactories[descriptor.Service]
		if !ok {
			continue
		}
		if _, exists := guestStubFuncs[factory]; !exists {
			t.Fatalf("host service %s is missing non-WASI guest factory stub %s", descriptor.Service, factory)
		}
	}
	for _, fn := range []string{"One", "All", "Count", "Insert", "Update", "Delete", "Transaction"} {
		if _, ok := dataStubFuncs[fn]; !ok {
			t.Fatalf("data host service is missing non-WASI data stub method %s", fn)
		}
	}
}

// TestHostServiceDescriptorCapabilitySource verifies capability and resource
// metadata are derived from the descriptor table without duplicate service
// method entries.
func TestHostServiceDescriptorCapabilitySource(t *testing.T) {
	seen := make(map[string]struct{})
	for _, descriptor := range HostServiceMethodDescriptors() {
		key := descriptor.Service + "." + descriptor.Method
		if descriptor.Service == "" || descriptor.Method == "" || descriptor.Capability == "" {
			t.Fatalf("host service descriptor has incomplete identity: %#v", descriptor)
		}
		if _, ok := seen[key]; ok {
			t.Fatalf("host service descriptor is duplicated: %s", key)
		}
		seen[key] = struct{}{}
		if got := RequiredCapabilityForHostServiceMethod(descriptor.Service, descriptor.Method); got != descriptor.Capability {
			t.Fatalf("descriptor %s expects capability %s, got %s", key, descriptor.Capability, got)
		}
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
			t.Fatalf("host service method %s is missing public protocol codec alias %s", methodKey, fn)
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
