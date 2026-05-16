// This file tests that multiple integration service instances share the same
// in-memory source-plugin enablement and route-binding state inside one host
// process.

package integration

import (
	"context"
	"testing"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/pluginhost"
)

// TestSharedStateCrossInstance verifies route guards and route-binding
// projections stay aligned even when different callers hold different
// integration service instances.
func TestSharedStateCrossInstance(t *testing.T) {
	shared := &sharedState{
		sourceRouteBindings: make(map[string][]pluginhost.SourceRouteBinding),
		enabledSnapshot:     make(map[string]bool),
	}
	first := &serviceImpl{sharedState: shared}
	second := &serviceImpl{sharedState: shared}

	first.SetPluginEnabledState("plugin-demo", true)

	firstChecker := first.buildBackgroundEnabledChecker()
	secondChecker := second.buildBackgroundEnabledChecker()
	if !firstChecker(context.Background(), "plugin-demo") {
		t.Fatal("expected first instance to read enabled snapshot")
	}
	if !secondChecker(context.Background(), "plugin-demo") {
		t.Fatal("expected second instance to share enabled snapshot updates")
	}

	first.setSourceRouteBindings("plugin-demo", []pluginhost.SourceRouteBinding{
		{
			PluginID: "plugin-demo",
			Method:   "GET",
			Path:     "/api/v1/plugins/plugin-demo/summary",
		},
	})
	bindings := second.ListSourceRouteBindings()
	if len(bindings) != 1 {
		t.Fatalf("expected second instance to observe 1 shared route binding, got %d", len(bindings))
	}
	if bindings[0].PluginID != "plugin-demo" || bindings[0].Path != "/api/v1/plugins/plugin-demo/summary" {
		t.Fatalf("unexpected shared route binding: %#v", bindings[0])
	}

	second.DeletePluginEnabledState("plugin-demo")
	if firstChecker(context.Background(), "plugin-demo") {
		t.Fatal("expected deleting shared snapshot entry to affect all instances")
	}
}

// TestStoreLoadedEnabledSnapshotBackfillsSharedState verifies a registry read
// warms the shared enablement snapshot for later filter passes.
func TestStoreLoadedEnabledSnapshotBackfillsSharedState(t *testing.T) {
	shared := &sharedState{
		sourceRouteBindings: make(map[string][]pluginhost.SourceRouteBinding),
		enabledSnapshot:     make(map[string]bool),
	}
	svc := &serviceImpl{sharedState: shared}

	svc.storeLoadedEnabledSnapshot(map[string]bool{
		"plugin-enabled":  true,
		"plugin-disabled": false,
	})

	enabledByID := map[string]bool{
		"plugin-enabled":  false,
		"plugin-disabled": true,
		"plugin-missing":  true,
	}
	if !svc.applyLoadedEnabledSnapshot(enabledByID) {
		t.Fatal("expected loaded snapshot to be applied")
	}
	if !enabledByID["plugin-enabled"] {
		t.Fatal("expected enabled plugin to remain enabled")
	}
	if enabledByID["plugin-disabled"] {
		t.Fatal("expected disabled plugin to be disabled")
	}
	if enabledByID["plugin-missing"] {
		t.Fatal("expected missing plugin to default to disabled")
	}
}

// TestPlatformOnlyGlobalPluginRemainsEnabledInTenantContext verifies
// platform-only governs tenant plugin-list visibility, not runtime
// availability for global plugin APIs and permission menus.
func TestPlatformOnlyGlobalPluginRemainsEnabledInTenantContext(t *testing.T) {
	svc := &serviceImpl{}
	ctx := datascope.WithTenantForTest(context.Background(), 42)

	enabled, err := svc.registryEnabledForTenant(ctx, &entity.SysPlugin{
		PluginId:    "multi-tenant",
		Installed:   catalog.InstalledYes,
		Status:      catalog.StatusEnabled,
		ScopeNature: catalog.ScopeNaturePlatformOnly.String(),
		InstallMode: catalog.InstallModeGlobal.String(),
	})
	if err != nil {
		t.Fatalf("expected platform-only global enablement check to succeed, got error: %v", err)
	}
	if !enabled {
		t.Fatal("expected platform-only global plugin to stay enabled in tenant context")
	}
}
