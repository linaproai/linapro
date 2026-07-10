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

	"lina-core/pkg/plugin/capability/capregistry"
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
			if method.PayloadKind == PayloadKindDedicated {
				if _, ok := dedicatedServices[method.Service]; !ok {
					t.Fatalf("ordinary host service %s uses dedicated codec without whitelist", key)
				}
				continue
			}
		}
	}
}

func TestCoreCatalogKeepsStaticServicesCoreOwned(t *testing.T) {
	for _, descriptor := range Catalog() {
		if descriptor.Owner != "" || descriptor.Version != "" {
			t.Fatalf("static core catalog must not carry owner-aware identity, got %#v", descriptor)
		}
		for _, method := range descriptor.Methods {
			if method.Owner != "" || method.Version != "" {
				t.Fatalf("static core method must not carry owner-aware identity, got %#v", method)
			}
		}
	}
}

func TestCatalogWithDescriptorsProjectsOwnerMethods(t *testing.T) {
	descriptor := capregistry.Descriptor{
		OwnerPluginID:   "linapro-ai-core",
		Service:         "ai",
		Version:         "v1",
		SourceContract:  "lina-plugin-linapro-ai-core/backend/cap/aicap",
		DynamicContract: "lina-plugin-linapro-ai-core/backend/cap/aicap/bridge",
		Methods: []capregistry.MethodDescriptor{
			{
				Method:          "text.generate",
				Capability:      "plugin.linapro-ai-core.ai.text.v1",
				Risk:            capregistry.RiskLevelExecute,
				ResourceKind:    capregistry.ResourceKindKey,
				RequestPayload:  "aitext.GenerateRequest",
				ResponsePayload: "aitext.GenerateResponse",
			},
		},
	}

	catalog, err := CatalogWithDescriptors([]capregistry.Descriptor{descriptor})
	if err != nil {
		t.Fatalf("merge owner descriptor catalog: %v", err)
	}
	var ownerService *ServiceDescriptor
	for i := range catalog {
		current := &catalog[i]
		if current.Owner == "linapro-ai-core" && current.Service == "ai" && current.Version == "v1" {
			ownerService = current
			break
		}
	}
	if ownerService == nil {
		t.Fatalf("expected owner-aware ai catalog entry, got %#v", catalog)
	}
	if ownerService.SourceContract != descriptor.SourceContract || ownerService.DynamicContract != descriptor.DynamicContract {
		t.Fatalf("expected owner contracts to be projected, got %#v", ownerService)
	}
	if ownerService.ResourceKind != ResourceKindKey {
		t.Fatalf("expected owner service resource kind key, got %s", ownerService.ResourceKind)
	}
	if len(ownerService.Methods) != 1 {
		t.Fatalf("expected one owner method, got %#v", ownerService.Methods)
	}
	method := ownerService.Methods[0]
	if method.Owner != "linapro-ai-core" || method.Service != "ai" || method.Version != "v1" {
		t.Fatalf("expected owner method identity to be projected, got %#v", method)
	}
	if method.Method != "text.generate" || method.Capability != "plugin.linapro-ai-core.ai.text.v1" {
		t.Fatalf("expected method metadata to be projected, got %#v", method)
	}
	if method.Risk != RiskLevelExecute || method.ResourceKind != ResourceKindKey {
		t.Fatalf("expected method risk/resource metadata, got %#v", method)
	}
	if method.RequestPayload != "aitext.GenerateRequest" || method.ResponsePayload != "aitext.GenerateResponse" {
		t.Fatalf("expected owner payload names, got %#v", method)
	}
	if method.PayloadKind != PayloadKindJSON || !method.Published || !method.GuestClient || !method.Dispatcher {
		t.Fatalf("expected owner method publication metadata, got %#v", method)
	}
}

func TestCatalogFromDescriptorsRejectsInvalidOwnerDescriptor(t *testing.T) {
	_, err := CatalogFromDescriptors([]capregistry.Descriptor{{Service: "ai", Version: "v1"}})
	if err == nil || !strings.Contains(err.Error(), "owner plugin id is required") {
		t.Fatalf("expected invalid descriptor error, got %v", err)
	}
}

func TestOrdinaryJSONServicesHaveNoDedicatedCapabilityCodecs(t *testing.T) {
	var (
		root          = repoRootForCatalogTest(t)
		protocolDir   = filepath.Dir(root)
		protocolTypes = declaredTypeNames(t, protocolDir)
		protocolFuncs = declaredFuncNames(t, protocolDir)
	)
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
