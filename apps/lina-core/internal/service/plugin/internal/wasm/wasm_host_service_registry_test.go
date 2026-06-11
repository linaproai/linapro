// This file verifies the parent wasm package wires the host-service dispatch
// registry explicitly and covers every catalog-published dispatcher method.

package wasm

import (
	"context"
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
