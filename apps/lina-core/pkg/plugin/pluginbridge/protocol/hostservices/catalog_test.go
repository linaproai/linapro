// This file verifies the public host-service catalog owns payload kind
// classification for ordinary JSON services and dedicated codec services.

package hostservices

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCatalogPayloadKinds(t *testing.T) {
	dedicatedServices := map[string]struct{}{
		"runtime":       {},
		"storage":       {},
		"network":       {},
		"data":          {},
		"cache":         {},
		"lock":          {},
		"hostconfig":    {},
		"manifest":      {},
		"ai":            {},
		"jobs":          {},
		"notifications": {},
		"plugins":       {},
	}

	for _, descriptor := range Catalog() {
		for _, method := range descriptor.Methods {
			key := method.Service + "." + method.Method
			if method.PayloadKind == "" {
				t.Fatalf("catalog method %s is missing payload kind", key)
			}
			if method.PayloadKind == PayloadKindReserved {
				continue
			}
			if method.PayloadKind == PayloadKindDedicated {
				if _, ok := dedicatedServices[method.Service]; !ok {
					t.Fatalf("ordinary host service %s uses dedicated codec without whitelist", key)
				}
				continue
			}
		}
	}
}

func TestOrdinaryJSONServicesHaveNoDedicatedCapabilityCodecs(t *testing.T) {
	root := repoRootForCatalogTest(t)
	protocolDir := filepath.Dir(root)
	protocolTypes := declaredTypeNames(t, protocolDir)
	protocolFuncs := declaredFuncNames(t, protocolDir)
	for _, name := range []string{
		"HostServiceUsersBatchGetRequest",
		"HostServiceUsersSearchRequest",
		"HostServiceUsersEnsureVisibleRequest",
		"HostServiceCapabilityUserRequest",
		"HostServiceCapabilityUsersRequest",
		"HostServiceCapabilityTenantRequest",
		"HostServiceCapabilityUserTenantRequest",
		"HostServiceCapabilityUserTenantSwitchRequest",
	} {
		if _, ok := protocolTypes[name]; ok {
			t.Fatalf("ordinary JSON host service must not keep dedicated request type %s", name)
		}
	}
	for _, name := range []string{
		"MarshalHostServiceUsersBatchGetRequest",
		"UnmarshalHostServiceUsersBatchGetRequest",
		"MarshalHostServiceUsersSearchRequest",
		"UnmarshalHostServiceUsersSearchRequest",
		"MarshalHostServiceUsersEnsureVisibleRequest",
		"UnmarshalHostServiceUsersEnsureVisibleRequest",
		"MarshalHostServiceCapabilityUserRequest",
		"UnmarshalHostServiceCapabilityUserRequest",
		"MarshalHostServiceCapabilityUsersRequest",
		"UnmarshalHostServiceCapabilityUsersRequest",
		"MarshalHostServiceCapabilityTenantRequest",
		"UnmarshalHostServiceCapabilityTenantRequest",
		"MarshalHostServiceCapabilityUserTenantRequest",
		"UnmarshalHostServiceCapabilityUserTenantRequest",
		"MarshalHostServiceCapabilityUserTenantSwitchRequest",
		"UnmarshalHostServiceCapabilityUserTenantSwitchRequest",
	} {
		if _, ok := protocolFuncs[name]; ok {
			t.Fatalf("ordinary JSON host service must not keep dedicated codec function %s", name)
		}
	}
}

func repoRootForCatalogTest(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("resolve test working directory failed: %v", err)
	}
	for {
		if filepath.Base(dir) == "hostservices" {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate hostservices test directory")
		}
		dir = parent
	}
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
				if ok && strings.HasPrefix(typeSpec.Name.Name, "HostService") {
					result[typeSpec.Name.Name] = struct{}{}
				}
			}
		}
	}
	return result
}

func declaredFuncNames(t *testing.T, dir string) map[string]struct{} {
	t.Helper()
	result := make(map[string]struct{})
	for _, file := range parseGoFiles(t, dir) {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if ok && strings.Contains(fn.Name.Name, "HostService") {
				result[fn.Name.Name] = struct{}{}
			}
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
