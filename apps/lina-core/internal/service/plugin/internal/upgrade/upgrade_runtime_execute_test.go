// This file covers unified runtime-upgrade execution helpers that do not need
// the root plugin facade.

package upgrade

import (
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// TestBuildAuthorizedHostServicesForPluginFiltersResourceScopes verifies
// dynamic upgrade lifecycle execution does not expose resource-scoped host
// services when no target release authorization snapshot exists yet.
func TestBuildAuthorizedHostServicesForPluginFiltersResourceScopes(t *testing.T) {
	manifest := &catalog.Manifest{
		ID:      "plugin-dev-dynamic-lifecycle-missing-release-auth",
		Version: "v0.9.1",
		HostServices: []*protocol.HostServiceSpec{
			{
				Service: protocol.HostServiceRuntime,
				Methods: []string{
					protocol.HostServiceMethodRuntimeLogWrite,
				},
			},
			{
				Service: protocol.HostServiceStorage,
				Methods: []string{
					protocol.HostServiceMethodStorageGet,
				},
				Paths: []string{"private-files/"},
			},
		},
	}

	filtered, err := cloneManifestWithAuthorizedHostServices(manifest, nil)
	if err != nil {
		t.Fatalf("expected missing release authorization to filter cleanly, got error: %v", err)
	}
	if filtered == nil {
		t.Fatal("expected filtered manifest")
	}
	if len(filtered.HostServices) != 1 || filtered.HostServices[0].Service != protocol.HostServiceRuntime {
		t.Fatalf("expected missing release to keep only capability host services, got %#v", filtered.HostServices)
	}
	if _, ok := filtered.HostCapabilities[protocol.CapabilityStorage]; ok {
		t.Fatalf("expected missing release to remove storage capability, got %#v", filtered.HostCapabilities)
	}
}
