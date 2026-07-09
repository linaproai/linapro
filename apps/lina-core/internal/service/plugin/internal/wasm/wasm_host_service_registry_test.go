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
		if _, ok := registry.lookup(descriptor.Service, descriptor.Method); !ok {
			t.Fatalf("host service dispatch registry is missing %s.%s", descriptor.Service, descriptor.Method)
		}
	}
	for _, registered := range registry.registeredMethods() {
		key := registered.service + "\x00" + registered.method
		if _, ok := expected[key]; !ok {
			t.Fatalf("host service dispatch registry contains orphan method %s.%s", registered.service, registered.method)
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
	response := registry.dispatch(context.Background(), hostServiceDispatchContext{
		service: hostservices.Methods()[0].Service,
		method:  hostservices.Methods()[0].Method,
	})
	if response == nil || response.Status != bridgehostcall.HostCallStatusInternalError {
		t.Fatalf("missing host call context should fail before invoking handler: %#v", response)
	}
}

func TestHostServiceDispatchRegisteredHandler(t *testing.T) {
	registry := newEmptyHostServiceDispatchRegistry()
	err := registry.register("users", "users.batch_get", func(_ context.Context, input hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
		if input.service != "users" || input.method != "users.batch_get" {
			t.Fatalf("unexpected input: %#v", input)
		}
		return bridgehostcall.NewHostCallSuccessResponse([]byte("ok"))
	})
	if err != nil {
		t.Fatalf("register handler failed: %v", err)
	}

	response := registry.dispatch(context.Background(), hostServiceDispatchContext{
		service: "users",
		method:  "users.batch_get",
	})
	if response == nil || response.Status != bridgehostcall.HostCallStatusSuccess || string(response.Payload) != "ok" {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestHostServiceDispatchUnknownMethod(t *testing.T) {
	response := newEmptyHostServiceDispatchRegistry().dispatch(context.Background(), hostServiceDispatchContext{
		service: "users",
		method:  "users.missing",
	})
	if response == nil || response.Status != bridgehostcall.HostCallStatusNotFound {
		t.Fatalf("unknown method should return not-found response: %#v", response)
	}
}

func TestHostServiceDispatchRejectsDuplicateRegistration(t *testing.T) {
	registry := newEmptyHostServiceDispatchRegistry()
	handler := func(context.Context, hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
		return bridgehostcall.NewHostCallEmptySuccessResponse()
	}
	if err := registry.register("users", "users.batch_get", handler); err != nil {
		t.Fatalf("first register failed: %v", err)
	}
	if err := registry.register("users", "users.batch_get", handler); err == nil {
		t.Fatal("duplicate register should fail")
	}
}

func TestHostServiceDispatchRejectsMissingRegistrationInput(t *testing.T) {
	handler := func(context.Context, hostServiceDispatchContext) *bridgehostcall.HostCallResponseEnvelope {
		return bridgehostcall.NewHostCallEmptySuccessResponse()
	}
	for _, tc := range []struct {
		name    string
		service string
		method  string
		handler hostServiceDispatchHandler
	}{
		{name: "empty service", method: "users.batch_get", handler: handler},
		{name: "empty method", service: "users", handler: handler},
		{name: "nil handler", service: "users", method: "users.batch_get"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := newEmptyHostServiceDispatchRegistry().register(tc.service, tc.method, tc.handler); err == nil {
				t.Fatal("invalid registration should fail")
			}
		})
	}
	if err := (*hostServiceDispatchRegistry)(nil).register("users", "users.batch_get", handler); err == nil {
		t.Fatal("nil registry registration should fail")
	}
}

func TestHostServicePublicHelpersStayOutOfDomainFiles(t *testing.T) {
	root := wasmPackageDir(t)
	usersFile := readWasmSourceFile(t, root, "wasm_host_service_users.go")
	if strings.Contains(usersFile, "func contextWithHostCallBizContext") {
		t.Fatal("contextWithHostCallBizContext must stay in a common host-service file, not in the users domain dispatcher")
	}

	commonFile := readWasmSourceFile(t, root, "wasm_host_service.go")
	if !strings.Contains(commonFile, "func contextWithHostCallBizContext") {
		t.Fatal("wasm_host_service.go must own contextWithHostCallBizContext as shared host-service helper")
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
