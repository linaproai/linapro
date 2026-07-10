// This file verifies the public host-service catalog owns payload kind
// classification for ordinary JSON services and freezes dedicated codec methods.

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

// frozenDedicatedMethods is the closed allowlist of historical dedicated binary
// codec methods. New core-owned host-service methods must use PayloadKindJSON
// or PayloadKindNone and must not expand this set.
var frozenDedicatedMethods = map[string]struct{}{
	"cache.delete":                {},
	"cache.expire":                {},
	"cache.get":                   {},
	"cache.incr":                  {},
	"cache.set":                   {},
	"data.batch_get":              {},
	"data.create":                 {},
	"data.delete":                 {},
	"data.get":                    {},
	"data.list":                   {},
	"data.transaction":            {},
	"data.update":                 {},
	"hostconfig.get":              {},
	"jobs.jobs.register":          {},
	"lock.acquire":                {},
	"lock.release":                {},
	"lock.renew":                  {},
	"manifest.get":                {},
	"network.request":             {},
	"notifications.messages.send": {},
	"plugins.config.get":          {},
	"runtime.info.node":           {},
	"runtime.info.now":            {},
	"runtime.info.uuid":           {},
	"runtime.log.write":           {},
	"runtime.state.delete":        {},
	"runtime.state.get":           {},
	"runtime.state.set":           {},
	"storage.delete":              {},
	"storage.get":                 {},
	"storage.list":                {},
	"storage.put":                 {},
	"storage.put.abort":           {},
	"storage.put.chunk":           {},
	"storage.put.commit":          {},
	"storage.put.init":            {},
	"storage.stat":                {},
}

func TestCatalogPayloadKinds(t *testing.T) {
	seenDedicated := map[string]struct{}{}
	for _, descriptor := range Catalog() {
		for _, method := range descriptor.Methods {
			key := method.Service + "." + method.Method
			if method.PayloadKind == "" {
				t.Fatalf("catalog method %s is missing payload kind", key)
			}
			switch method.PayloadKind {
			case PayloadKindDedicated:
				if _, ok := frozenDedicatedMethods[key]; !ok {
					t.Fatalf("dedicated codec method %s is not in the frozen allowlist; new methods must use JSON envelope", key)
				}
				seenDedicated[key] = struct{}{}
			case PayloadKindJSON, PayloadKindNone:
				// Allowed for all methods, including new ones.
			default:
				t.Fatalf("catalog method %s has unknown payload kind %q", key, method.PayloadKind)
			}
		}
	}
	for key := range frozenDedicatedMethods {
		if _, ok := seenDedicated[key]; !ok {
			t.Fatalf("frozen dedicated method %s is missing from catalog", key)
		}
	}
}

func TestCatalogUsesWireConstants(t *testing.T) {
	// Catalog entries must reference wire_constants.go values so method/service
	// wire strings are maintained in one place only.
	for _, descriptor := range Catalog() {
		serviceConst := hostServiceConstNamesForTest[descriptor.Service]
		if serviceConst == "" {
			t.Fatalf("catalog service %q missing const name map", descriptor.Service)
		}
		// Resolve service const value via package-level const table from AST of wire_constants.go
		// is unnecessary when Service field already stores the const value; check method linkage.
		for _, method := range descriptor.Methods {
			if method.MethodConst == "" {
				t.Fatalf("catalog method %s.%s missing MethodConst", descriptor.Service, method.Method)
			}
			// MethodConst identifies the Go constant; its value must equal Method wire string.
			// Evaluating const identifiers happens by matching known package constants below.
		}
	}
	constants := loadHostservicesWireConstants(t)
	for _, descriptor := range Catalog() {
		serviceConst := hostServiceConstNamesForTest[descriptor.Service]
		if got := constants[serviceConst]; got != descriptor.Service {
			t.Fatalf("wire constant %s = %q, catalog service = %q (catalog must use the constant)", serviceConst, got, descriptor.Service)
		}
		for _, method := range descriptor.Methods {
			if got := constants[method.MethodConst]; got != method.Method {
				t.Fatalf("wire constant %s = %q, catalog method = %q (catalog must use the constant)", method.MethodConst, got, method.Method)
			}
		}
	}
}

func loadHostservicesWireConstants(t *testing.T) map[string]string {
	t.Helper()
	filePath := filepath.Join(".", "wire_constants.go")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read wire_constants.go: %v", err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), filePath, content, 0)
	if err != nil {
		t.Fatalf("parse wire_constants.go: %v", err)
	}
	constants := map[string]string{}
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
			for i, ident := range valueSpec.Names {
				if i >= len(valueSpec.Values) {
					continue
				}
				basic, ok := valueSpec.Values[i].(*ast.BasicLit)
				if !ok || basic.Kind != token.STRING {
					continue
				}
				constants[ident.Name] = strings.Trim(basic.Value, `"`)
			}
		}
	}
	if len(constants) == 0 {
		t.Fatal("expected HostService* wire constants in wire_constants.go")
	}
	return constants
}

// hostServiceConstNamesForTest maps catalog service wire values to HostService* names.
var hostServiceConstNamesForTest = map[string]string{
	"runtime":       "HostServiceRuntime",
	"storage":       "HostServiceStorage",
	"network":       "HostServiceNetwork",
	"data":          "HostServiceData",
	"cache":         "HostServiceCache",
	"lock":          "HostServiceLock",
	"hostconfig":    "HostServiceHostConfig",
	"manifest":      "HostServiceManifest",
	"apidoc":        "HostServiceAPIDoc",
	"auth":          "HostServiceAuth",
	"users":         "HostServiceUsers",
	"bizctx":        "HostServiceBizCtx",
	"dict":          "HostServiceDict",
	"files":         "HostServiceFiles",
	"jobs":          "HostServiceJobs",
	"notifications": "HostServiceNotifications",
	"plugins":       "HostServicePlugins",
	"route":         "HostServiceRoute",
	"sessions":      "HostServiceSessions",
	"org":           "HostServiceOrg",
	"tenant":        "HostServiceTenant",
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
		t.Fatalf("CatalogWithDescriptors returned error: %v", err)
	}
	var found bool
	for _, service := range catalog {
		if service.Owner != "linapro-ai-core" || service.Service != "ai" || service.Version != "v1" {
			continue
		}
		found = true
		if len(service.Methods) != 1 || service.Methods[0].Method != "text.generate" {
			t.Fatalf("unexpected owner methods: %#v", service.Methods)
		}
		if service.Methods[0].PayloadKind != PayloadKindJSON || !service.Methods[0].Published || !service.Methods[0].GuestClient || !service.Methods[0].Dispatcher {
			t.Fatalf("unexpected owner method projection: %#v", service.Methods[0])
		}
	}
	if !found {
		t.Fatal("expected owner-aware AI service projection in catalog")
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
