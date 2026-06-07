// Tests for the dynamic-plugin side pluginbridge guest capability directory.

package guest

import (
	"context"
	"errors"
	"testing"

	"lina-core/pkg/plugin/capability/aicap/aitext"
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
	assertSameClient(t, services.Plugins().Config(), pluginConfig(), "plugin config")
	assertSameClient(t, services.Notify(), Notify(), "notify")
	assertSameClient(t, services.Cron(), Cron(), "cron")
	assertSameClient(t, services.HostConfig(), HostConfig(), "host config")
	assertSameClient(t, services.Manifest(), Manifest(), "manifest")
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
	assertGuestInterfaceType(t, (*RuntimeHostService)(nil), "RuntimeHostService")
	assertGuestInterfaceType(t, (*StorageHostService)(nil), "StorageHostService")
	assertGuestInterfaceType(t, (*NetworkHostService)(nil), "NetworkHostService")
	assertGuestInterfaceType(t, (*CacheHostService)(nil), "CacheHostService")
	assertGuestInterfaceType(t, (*LockHostService)(nil), "LockHostService")
	assertGuestInterfaceType(t, (*ConfigHostService)(nil), "ConfigHostService")
	assertGuestInterfaceType(t, (*NotifyHostService)(nil), "NotifyHostService")
	assertGuestInterfaceType(t, (*CronHostService)(nil), "CronHostService")
	assertGuestInterfaceType(t, (*HostConfigHostService)(nil), "HostConfigHostService")
	assertGuestInterfaceType(t, (*ManifestHostService)(nil), "ManifestHostService")
}

// assertSameClient verifies directory methods return the package default clients.
func assertSameClient(t *testing.T, got any, want any, name string) {
	t.Helper()

	if got != want {
		t.Fatalf("expected %s client to come from pluginbridge guest package", name)
	}
}
