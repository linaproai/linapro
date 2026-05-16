// This file covers dynamic-plugin lifecycle precondition response handling.

package runtime

import (
	"encoding/json"
	"net/http"
	"testing"

	"lina-core/internal/service/plugin/internal/catalog"
	bridgecontract "lina-core/pkg/pluginbridge/contract"
	"lina-core/pkg/pluginhost"
)

// TestApplyDynamicLifecycleResponseRecordsVeto verifies explicit guest veto
// responses are preserved as lifecycle decisions instead of bridge errors.
func TestApplyDynamicLifecycleResponseRecordsVeto(t *testing.T) {
	decision := &DynamicLifecycleDecision{
		PluginID:  "plugin-dynamic-veto",
		Operation: pluginhost.LifecycleHookBeforeInstall,
		OK:        true,
	}

	err := applyDynamicLifecycleResponse(decision, &bridgecontract.BridgeResponseEnvelopeV1{
		StatusCode: http.StatusOK,
		Body:       []byte(`{"ok":false,"reason":"plugin.dynamic.veto"}`),
	})
	if err != nil {
		t.Fatalf("expected veto response to decode without bridge error, got %v", err)
	}
	if decision.OK || decision.Reason != "plugin.dynamic.veto" {
		t.Fatalf("expected explicit lifecycle veto decision, got %#v", decision)
	}
}

// TestBuildDynamicLifecycleRequestPublishesTypedManifestSnapshot verifies
// dynamic Upgrade callbacks receive the shared typed manifest snapshot contract.
func TestBuildDynamicLifecycleRequestPublishesTypedManifestSnapshot(t *testing.T) {
	request, err := buildDynamicLifecycleRequest(
		&catalog.Manifest{
			ID:              "plugin-dynamic-upgrade",
			RuntimeArtifact: &catalog.ArtifactSpec{Path: "/tmp/plugin-dynamic-upgrade.wasm"},
			BridgeSpec:      &bridgecontract.BridgeSpec{RouteExecution: true},
		},
		&bridgecontract.LifecycleContract{
			Operation:    bridgecontract.LifecycleOperationBeforeUpgrade,
			RequestType:  "DynamicBeforeUpgradeReq",
			InternalPath: "/__lifecycle/before-upgrade",
		},
		DynamicLifecycleInput{
			PluginID:  "plugin-dynamic-upgrade",
			Operation: pluginhost.LifecycleHookBeforeUpgrade,
			FromManifest: &catalog.ManifestSnapshot{
				ID:                      "plugin-dynamic-upgrade",
				Name:                    "Dynamic Upgrade",
				Version:                 "v0.1.0",
				Type:                    "dynamic",
				ScopeNature:             "tenant_aware",
				SupportsMultiTenant:     true,
				DefaultInstallMode:      "tenant_scoped",
				Description:             "upgrade test",
				InstallSQLCount:         1,
				UninstallSQLCount:       1,
				MockSQLCount:            2,
				MenuCount:               3,
				BackendHookCount:        4,
				ResourceSpecCount:       5,
				HostServiceAuthRequired: false,
			},
			ToManifest: &catalog.ManifestSnapshot{
				ID:                      "plugin-dynamic-upgrade",
				Name:                    "Dynamic Upgrade",
				Version:                 "v0.2.0",
				Type:                    "dynamic",
				ScopeNature:             "tenant_aware",
				SupportsMultiTenant:     true,
				DefaultInstallMode:      "tenant_scoped",
				Description:             "upgrade test",
				InstallSQLCount:         2,
				UninstallSQLCount:       1,
				MockSQLCount:            3,
				MenuCount:               4,
				BackendHookCount:        5,
				ResourceSpecCount:       6,
				HostServiceAuthRequired: true,
			},
		},
	)
	if err != nil {
		t.Fatalf("expected lifecycle request to build, got %v", err)
	}
	payload := &bridgecontract.LifecycleRequest{}
	if err = json.Unmarshal(request.Request.Body, payload); err != nil {
		t.Fatalf("expected lifecycle request body to decode, got %v", err)
	}

	if payload.FromManifest == nil ||
		payload.FromManifest.Version != "v0.1.0" ||
		payload.FromManifest.HostServiceAuthRequired {
		t.Fatalf("unexpected from manifest payload: %#v", payload.FromManifest)
	}
	if payload.ToManifest == nil ||
		payload.ToManifest.Version != "v0.2.0" ||
		payload.ToManifest.SupportsMultiTenant != true ||
		payload.ToManifest.ResourceSpecCount != 6 ||
		payload.ToManifest.HostServiceAuthRequired != true {
		t.Fatalf("unexpected to manifest payload: %#v", payload.ToManifest)
	}
}

// TestPublishedManifestSnapshotUsesBridgeContract verifies catalog snapshots
// project into the shared bridge lifecycle snapshot contract.
func TestPublishedManifestSnapshotUsesBridgeContract(t *testing.T) {
	snapshot := catalog.PublishedManifestSnapshot(&catalog.ManifestSnapshot{
		ID:                      "plugin-dynamic-upgrade",
		Name:                    "Dynamic Upgrade",
		Version:                 "v0.2.0",
		Type:                    "dynamic",
		ScopeNature:             "tenant_aware",
		SupportsMultiTenant:     true,
		DefaultInstallMode:      "tenant_scoped",
		Description:             "upgrade test",
		InstallSQLCount:         2,
		UninstallSQLCount:       1,
		MockSQLCount:            3,
		MenuCount:               4,
		BackendHookCount:        5,
		ResourceSpecCount:       6,
		HostServiceAuthRequired: true,
	})

	if snapshot.ID != "plugin-dynamic-upgrade" ||
		snapshot.Version != "v0.2.0" ||
		snapshot.SupportsMultiTenant != true ||
		snapshot.ResourceSpecCount != 6 ||
		snapshot.HostServiceAuthRequired != true {
		t.Fatalf("unexpected lifecycle snapshot values: %#v", snapshot)
	}
}
