// This file verifies the parent wasm package wires the host-service dispatch
// registry explicitly and covers every catalog-published dispatcher method.

package wasm

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"lina-core/internal/service/plugin/internal/wasm/hostservicedispatch"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-core/pkg/plugin/pluginbridge/protocol/hostservices"
)

func TestHostServiceDispatchRegistryCoversCatalog(t *testing.T) {
	registry, err := defaultHostServiceDispatchRegistry()
	if err != nil {
		t.Fatalf("build host service dispatch registry failed: %v", err)
	}
	expected := make(map[string]struct{})
	for _, descriptor := range hostservices.Methods() {
		if !descriptor.Published || !descriptor.Dispatcher {
			continue
		}
		expected[descriptor.Service+"\x00"+descriptor.Method] = struct{}{}
		if _, ok := registry.Lookup(descriptor.Service, descriptor.Method); !ok {
			t.Fatalf("host service dispatch registry is missing %s.%s", descriptor.Service, descriptor.Method)
		}
	}
	for _, registered := range registry.Methods() {
		key := registered.Service + "\x00" + registered.Method
		if _, ok := expected[key]; !ok {
			t.Fatalf("host service dispatch registry contains orphan method %s.%s", registered.Service, registered.Method)
		}
		delete(expected, key)
	}
	for key := range expected {
		t.Fatalf("host service dispatch registry is missing catalog method key %q", key)
	}
}

func TestHostServiceDispatchRegistryRejectsMissingContext(t *testing.T) {
	registry, err := newHostServiceDispatchRegistry()
	if err != nil {
		t.Fatalf("build host service dispatch registry failed: %v", err)
	}
	response := registry.Dispatch(context.Background(), hostservicedispatch.Context{
		Service: hostservices.Methods()[0].Service,
		Method:  hostservices.Methods()[0].Method,
	})
	if response == nil || response.Status != bridgehostcall.HostCallStatusInternalError {
		t.Fatalf("missing host call context should fail before invoking handler: %#v", response)
	}
}

func TestHostServicePublicHelpersStayOutOfDomainFiles(t *testing.T) {
	root := wasmPackageDir(t)
	usersFile := readWasmSourceFile(t, root, "wasm_host_service_users.go")
	if strings.Contains(usersFile, "func capabilityContextForHostCall") {
		t.Fatal("capabilityContextForHostCall must stay in a common host-service file, not in the users domain dispatcher")
	}

	commonFile := readWasmSourceFile(t, root, "wasm_host_service.go")
	if !strings.Contains(commonFile, "func capabilityContextForHostCall") {
		t.Fatal("wasm_host_service.go must own capabilityContextForHostCall as shared host-service helper")
	}
}

func TestHostServiceEntrypointUsesRegistryDispatch(t *testing.T) {
	root := wasmPackageDir(t)
	entrypoint := readWasmSourceFile(t, root, "wasm_host_service.go")
	if !strings.Contains(entrypoint, "dispatchRegisteredHostService(ctx, hcc, request)") {
		t.Fatal("host-service entrypoint must dispatch through the explicit registry")
	}
	if strings.Contains(entrypoint, "switch request.Service") || strings.Contains(entrypoint, "switch strings.TrimSpace(request.Service)") {
		t.Fatal("host-service entrypoint must not reintroduce a service-level switch")
	}
}

func wasmPackageDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve wasm package directory failed")
	}
	return filepath.Dir(file)
}

func readWasmSourceFile(t *testing.T, root string, name string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(root, name))
	if err != nil {
		t.Fatalf("read %s failed: %v", name, err)
	}
	return string(content)
}
