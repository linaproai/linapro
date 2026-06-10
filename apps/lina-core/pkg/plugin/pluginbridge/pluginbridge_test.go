// Tests for the dynamic-plugin side pluginbridge guest public contract,
// capability directory, runtime, and small host-call helpers.

package pluginbridge

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"lina-core/pkg/plugin/capability/aicap/aitext"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// TestDefaultDirectoryReturnsCapabilityClients verifies the guest directory
// owns host-service client semantics instead of exposing pluginbridge guest
// client types.
func TestDefaultDirectoryReturnsCapabilityClients(t *testing.T) {
	services := New()

	assertSameClient(t, services.Runtime(), Runtime(), "runtime")
	assertSameClient(t, services.Storage(), Storage(), "storage")
	assertSameClient(t, services.Network(), Network(), "network")
	if services.RecordStore() == nil {
		t.Fatal("expected record store facade to come from pluginbridge guest directory")
	}
	assertSameClient(t, services.Cache(), Cache(), "cache")
	assertSameClient(t, services.Lock(), Lock(), "lock")
	if services.Plugins().Config() == nil {
		t.Fatal("expected plugin config capability client")
	}
	if services.Plugins().Lifecycle() == nil {
		t.Fatal("expected plugin lifecycle capability client")
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
	_, err := New().Org().GetUserDeptIDs(context.Background(), 1)
	if !errors.Is(err, ErrHostCallsUnavailable) {
		t.Fatalf("expected non-WASI org capability to use host-call stub, got %v", err)
	}
	_, err = New().Tenant().ListUserTenants(context.Background(), 1)
	if !errors.Is(err, ErrHostCallsUnavailable) {
		t.Fatalf("expected non-WASI tenant capability to use host-call stub, got %v", err)
	}
	_, err = New().AI().Text().GenerateText(context.Background(), aitext.GenerateRequest{Purpose: "content.summary"})
	if !errors.Is(err, ErrHostCallsUnavailable) {
		t.Fatalf("expected non-WASI AI capability to use host-call stub, got %v", err)
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
	assertGuestInterfaceType(t, (*HostConfigHostService)(nil), "HostConfigHostService")
	assertGuestInterfaceType(t, (*ManifestHostService)(nil), "ManifestHostService")
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

// TestStorageListEffectiveLimit verifies guest storage list responses expose
// the same bounded limit semantics as storagecap.Service implementations.
func TestStorageListEffectiveLimit(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   int
		want int
	}{
		{name: "default", in: 0, want: storagecap.DefaultListLimit},
		{name: "negative default", in: -1, want: storagecap.DefaultListLimit},
		{name: "bounded", in: 10, want: 10},
		{name: "max", in: storagecap.MaxListLimit + 1, want: storagecap.MaxListLimit},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := storageListEffectiveLimit(c.in); got != c.want {
				t.Fatalf("storageListEffectiveLimit(%d) = %d, want %d", c.in, got, c.want)
			}
		})
	}
}

// assertSameClient verifies directory methods return the package default clients.
func assertSameClient(t *testing.T, got any, want any, name string) {
	t.Helper()

	if got != want {
		t.Fatalf("expected %s client to come from pluginbridge guest package", name)
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
