// This file verifies host-service dispatch registry behavior independent of
// concrete Wasm runtime handlers so registration failures are caught early.
package hostservicedispatch

import (
	"context"
	"testing"

	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
)

func TestRegistryDispatchRegisteredHandler(t *testing.T) {
	registry := NewRegistry()
	err := registry.Register("users", "users.batch_get", func(_ context.Context, input Context) *bridgehostcall.HostCallResponseEnvelope {
		if input.Service != "users" || input.Method != "users.batch_get" {
			t.Fatalf("unexpected input: %#v", input)
		}
		return bridgehostcall.NewHostCallSuccessResponse([]byte("ok"))
	})
	if err != nil {
		t.Fatalf("register handler failed: %v", err)
	}

	response := registry.Dispatch(context.Background(), Context{
		Service: "users",
		Method:  "users.batch_get",
	})
	if response == nil || response.Status != bridgehostcall.HostCallStatusSuccess || string(response.Payload) != "ok" {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestRegistryDispatchUnknownMethod(t *testing.T) {
	response := NewRegistry().Dispatch(context.Background(), Context{
		Service: "users",
		Method:  "users.missing",
	})
	if response == nil || response.Status != bridgehostcall.HostCallStatusNotFound {
		t.Fatalf("unknown method should return not-found response: %#v", response)
	}
}

func TestRegistryRejectsDuplicateRegistration(t *testing.T) {
	registry := NewRegistry()
	handler := func(context.Context, Context) *bridgehostcall.HostCallResponseEnvelope {
		return bridgehostcall.NewHostCallEmptySuccessResponse()
	}
	if err := registry.Register("users", "users.batch_get", handler); err != nil {
		t.Fatalf("first register failed: %v", err)
	}
	if err := registry.Register("users", "users.batch_get", handler); err == nil {
		t.Fatal("duplicate register should fail")
	}
}

func TestRegistryRejectsMissingRegistrationInput(t *testing.T) {
	handler := func(context.Context, Context) *bridgehostcall.HostCallResponseEnvelope {
		return bridgehostcall.NewHostCallEmptySuccessResponse()
	}
	for _, tc := range []struct {
		name    string
		service string
		method  string
		handler Handler
	}{
		{name: "empty service", method: "users.batch_get", handler: handler},
		{name: "empty method", service: "users", handler: handler},
		{name: "nil handler", service: "users", method: "users.batch_get"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := NewRegistry().Register(tc.service, tc.method, tc.handler); err == nil {
				t.Fatal("invalid registration should fail")
			}
		})
	}
	if err := (*Registry)(nil).Register("users", "users.batch_get", handler); err == nil {
		t.Fatal("nil registry registration should fail")
	}
}
