// Tests for the dynamic-plugin side pluginbridge guest public contract,
// capability directory, runtime, and small host-call helpers.

package pluginbridge

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// TestDefaultDirectoryReturnsCapabilityClients verifies the guest directory
// owns host-service client semantics instead of exposing pluginbridge guest
// client types.
func TestDefaultDirectoryReturnsCapabilityClients(t *testing.T) {
	services := New()

	assertClientAvailable(t, services.Runtime(), "runtime")
	assertClientAvailable(t, services.Storage(), "storage")
	assertClientAvailable(t, services.Network(), "network")
	if services.RecordStore() == nil {
		t.Fatal("expected record store facade to come from pluginbridge guest directory")
	}
	assertClientAvailable(t, services.Cache(), "cache")
	assertClientAvailable(t, services.Lock(), "lock")
	if services.Plugins().Config() == nil {
		t.Fatal("expected plugin config capability client")
	}
	if services.Plugins().State() == nil {
		t.Fatal("expected plugin enablement capability client")
	}
	if services.Jobs() == nil {
		t.Fatal("expected jobs capability client")
	}
	if services.HostConfig() == nil {
		t.Fatal("expected host config capability client")
	}
	if services.Manifest() == nil {
		t.Fatal("expected manifest capability client")
	}
}

// TestSharedCapabilityServicesUseBridgeTransport verifies pluginbridge guest
// clients use independent structured host services and surface unsupported
// stubs in ordinary Go builds.
func TestSharedCapabilityServicesUseBridgeTransport(t *testing.T) {
	_, err := New().Org().Assignment().GetUserDeptIDs(context.Background(), 1)
	if !errors.Is(err, ErrHostCallsUnavailable) {
		t.Fatalf("expected non-WASI org capability to use host-call stub, got %v", err)
	}
	_, err = New().Tenant().Membership().ListByUser(context.Background(), 1)
	if !errors.Is(err, ErrHostCallsUnavailable) {
		t.Fatalf("expected non-WASI tenant capability to use host-call stub, got %v", err)
	}
}

// TestGuestCapabilityContractsUseInterfaces verifies guest-facing capability
// clients are published as interfaces.
func TestGuestCapabilityContractsUseInterfaces(t *testing.T) {
	assertGuestInterfaceType(t, (*Services)(nil), "Services")
	assertGuestInterfaceType(t, (*Declarations)(nil), "Declarations")
	assertGuestInterfaceType(t, (*RouteDeclarations)(nil), "RouteDeclarations")
	assertGuestInterfaceType(t, (*JobDeclarations)(nil), "JobDeclarations")
	assertGuestInterfaceType(t, (*GuestRuntime)(nil), "GuestRuntime")
	assertGuestInterfaceType(t, (*GuestControllerRouteDispatcher)(nil), "GuestControllerRouteDispatcher")
	assertGuestInterfaceType(t, (*RuntimeHostService)(nil), "RuntimeHostService")
	assertGuestInterfaceType(t, (*NetworkHostService)(nil), "NetworkHostService")
}

// TestPluginBridgeDoesNotExposeRootAbilityClientFacades verifies ordinary
// capability clients stay behind the Services directory instead of root package
// functions that bypass the directory boundary.
func TestPluginBridgeDoesNotExposeRootAbilityClientFacades(t *testing.T) {
	t.Parallel()

	forbidden := map[string]struct{}{
		"Runtime":             {},
		"Storage":             {},
		"Network":             {},
		"Cache":               {},
		"Lock":                {},
		"HostConfig":          {},
		"Manifest":            {},
		"Data":                {},
		"RecordStore":         {},
		"Cron":                {},
		"HostLog":             {},
		"HostStateGet":        {},
		"HostStateGetMany":    {},
		"HostStateSet":        {},
		"HostStateSetMany":    {},
		"HostStateDelete":     {},
		"HostStateDeleteMany": {},
		"HostStateGetInt":     {},
		"HostStateSetInt":     {},
	}
	var violations []string
	walkErr := filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != "." {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fileSet := token.NewFileSet()
		parsed, parseErr := parser.ParseFile(fileSet, path, nil, 0)
		if parseErr != nil {
			return parseErr
		}
		for _, decl := range parsed.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Recv != nil {
				continue
			}
			if _, exists := forbidden[funcDecl.Name.Name]; exists {
				violations = append(violations, path+" exposes "+funcDecl.Name.Name)
			}
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("scan pluginbridge root facades: %v", walkErr)
	}
	for _, violation := range violations {
		t.Errorf("pluginbridge root ability facade violation: %s", violation)
	}
}

// TestGuestRuntimeRoundTrip verifies the guest runtime allocator and execute
// path expose one decodable bridge response.
func TestGuestRuntimeRoundTrip(t *testing.T) {
	runtime := NewGuestRuntime(func(request *protocol.BridgeRequestEnvelopeV1) (*protocol.BridgeResponseEnvelopeV1, error) {
		return protocol.NewJSONResponse(200, []byte(`{"ok":true}`)), nil
	})

	requestContent, err := protocol.EncodeRequestEnvelope(&protocol.BridgeRequestEnvelopeV1{
		PluginID: "linapro-demo-dynamic",
	})
	if err != nil {
		t.Fatalf("expected request encode to succeed, got error: %v", err)
	}

	pointer := runtime.Alloc(uint32(len(requestContent)))
	if pointer == 0 {
		t.Fatal("expected guest alloc to return non-zero pointer")
	}
	copy(runtime.RequestBuffer(), requestContent)

	responsePointer, responseLength, err := runtime.Execute(uint32(len(requestContent)))
	if err != nil {
		t.Fatalf("expected guest execute to succeed, got error: %v", err)
	}
	if responsePointer == 0 || responseLength == 0 {
		t.Fatal("expected guest execute to expose one encoded response")
	}

	response, err := protocol.DecodeResponseEnvelope(runtime.ResponseBuffer())
	if err != nil {
		t.Fatalf("expected response decode to succeed, got error: %v", err)
	}
	if response.StatusCode != 200 || string(response.Body) != `{"ok":true}` {
		t.Fatalf("unexpected guest response: %#v", response)
	}
}

// TestPluginBridgeDoesNotImportCapabilitySPI verifies dynamic-plugin guest SDK
// code does not compile provider SPI-only packages into guest closures.
func TestPluginBridgeDoesNotImportCapabilitySPI(t *testing.T) {
	t.Parallel()

	violations := collectPluginbridgeImportViolations(t, ".", func(importPath string) bool {
		return isCapabilitySPIImport(importPath)
	})
	for _, violation := range violations {
		t.Errorf("pluginbridge SPI import violation: %s", violation)
	}
}

// assertClientAvailable verifies directory methods and package helpers return a
// concrete guest client. Client construction is intentionally per-call so all
// host service families share the injected invoker shape.
func assertClientAvailable(t *testing.T, got any, name string) {
	t.Helper()

	if got == nil {
		t.Fatalf("expected %s client", name)
	}
}

// assertGuestInterfaceType verifies the reflected type under test is an
// interface.
func assertGuestInterfaceType(t *testing.T, value interface{}, name string) {
	t.Helper()

	if reflect.TypeOf(value).Elem().Kind() != reflect.Interface {
		t.Fatalf("expected %s to be declared as interface", name)
	}
}

// collectPluginbridgeImportViolations walks production Go files under scanRoot
// and reports imports rejected by match.
func collectPluginbridgeImportViolations(t *testing.T, scanRoot string, match func(importPath string) bool) []string {
	t.Helper()

	var violations []string
	walkErr := filepath.WalkDir(scanRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fileSet := token.NewFileSet()
		parsed, parseErr := parser.ParseFile(fileSet, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			return parseErr
		}
		for _, importSpec := range parsed.Imports {
			importPath, unquoteErr := strconv.Unquote(importSpec.Path.Value)
			if unquoteErr != nil {
				return unquoteErr
			}
			if match(importPath) {
				violations = append(violations, path+" imports "+importPath)
			}
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("scan %s for forbidden imports: %v", scanRoot, walkErr)
	}
	return violations
}

// isCapabilitySPIImport reports whether one import path targets a capability
// provider SPI package whose path segment ends with "spi".
func isCapabilitySPIImport(importPath string) bool {
	const prefix = "lina-core/pkg/plugin/capability/"
	if !strings.HasPrefix(importPath, prefix) {
		return false
	}
	for _, segment := range strings.Split(strings.TrimPrefix(importPath, prefix), "/") {
		if strings.HasSuffix(segment, "spi") {
			return true
		}
	}
	return false
}
