// This file tests plugin enablement snapshots, tenant visibility, and shared
// source route-binding state used by integration enablement guards.

package integration

import (
	"context"
	pluginv1 "lina-core/api/plugin/v1"
	"testing"

	"lina-core/internal/service/datascope"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/plugin/pluginhost"
	"lina-core/pkg/statusflag"
)

type testClusterTopology struct {
	enabled bool
	primary bool
	nodeID  string
}

func (t testClusterTopology) Start(context.Context) {}

func (t testClusterTopology) Stop(context.Context) {}

func (t testClusterTopology) IsEnabled() bool {
	return t.enabled
}

func (t testClusterTopology) IsPrimary() bool {
	return t.primary
}

func (t testClusterTopology) NodeID() string {
	return t.nodeID
}

// TestSharedStateCrossInstance verifies route guards and route-binding
// projections stay aligned even when different callers hold different
// integration service instances.
func TestSharedStateCrossInstance(t *testing.T) {
	var (
		shared = newSharedState()
		first  = &serviceImpl{sharedState: shared}
		second = &serviceImpl{sharedState: shared}
	)

	first.SetPluginEnabledState("plugin-demo", true)

	if !first.canExposePluginBusinessEntries(context.Background(), "plugin-demo") {
		t.Fatal("expected first instance to read enabled snapshot")
	}
	if !second.canExposePluginBusinessEntries(context.Background(), "plugin-demo") {
		t.Fatal("expected second instance to share enabled snapshot updates")
	}
	if !first.CanExposeBusinessEntries(context.Background(), "plugin-demo") {
		t.Fatal("expected platform CanExposeBusinessEntries to reuse loaded shared snapshot")
	}
	if first.CanExposeBusinessEntries(datascope.WithTenantScope(context.Background(), 42), "plugin-demo") {
		t.Fatal("expected tenant CanExposeBusinessEntries to bypass platform snapshot")
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
	if first.canExposePluginBusinessEntries(context.Background(), "plugin-demo") {
		t.Fatal("expected deleting shared snapshot entry to affect all instances")
	}
}

// TestSourceJobRegistrarReportsPrimaryNodeFromTopology verifies jobs registrars
// expose primary-node state from the integration topology service directly.
func TestSourceJobRegistrarReportsPrimaryNodeFromTopology(t *testing.T) {
	registrar := newSourceJobRegistrar("plugin-demo", &serviceImpl{
		topology: testClusterTopology{enabled: true, primary: false, nodeID: "node-a"},
	})
	if registrar.IsPrimaryNode() {
		t.Fatal("expected registrar to report non-primary topology")
	}

	registrar = newSourceJobRegistrar("plugin-demo", &serviceImpl{
		topology: testClusterTopology{enabled: true, primary: true, nodeID: "node-b"},
	})
	if !registrar.IsPrimaryNode() {
		t.Fatal("expected registrar to report primary topology")
	}
}

// TestStoreLoadedEnabledSnapshotBackfillsSharedState verifies a registry read
// warms the shared enablement snapshot for later filter passes.
func TestStoreLoadedEnabledSnapshotBackfillsSharedState(t *testing.T) {
	shared := newSharedState()
	svc := &serviceImpl{sharedState: shared}

	svc.storeLoadedEnabledSnapshot(context.Background(), map[string]bool{
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

// TestTenantSnapshotDoesNotOverwritePlatformSnapshot verifies tenant-scoped
// visibility checks cannot poison the shared platform menu-filter snapshot.
func TestTenantSnapshotDoesNotOverwritePlatformSnapshot(t *testing.T) {
	shared := newSharedState()
	svc := &serviceImpl{sharedState: shared}

	svc.storeLoadedEnabledSnapshot(context.Background(), map[string]bool{
		"linapro-monitor-loginlog": true,
	})
	svc.storeLoadedEnabledSnapshot(datascope.WithTenantScope(context.Background(), 42), map[string]bool{
		"linapro-monitor-loginlog": false,
	})

	enabledByID := map[string]bool{"linapro-monitor-loginlog": false}
	if !svc.applyLoadedEnabledSnapshot(enabledByID) {
		t.Fatal("expected platform snapshot to remain available")
	}
	if !enabledByID["linapro-monitor-loginlog"] {
		t.Fatal("expected tenant snapshot not to hide platform admin monitor menu")
	}
}

// TestPlatformOnlyGlobalPluginRemainsEnabledInTenantContext verifies
// platform-only governs tenant plugin-list visibility, not runtime
// availability for global plugin APIs and permission menus.
func TestPlatformOnlyGlobalPluginRemainsEnabledInTenantContext(t *testing.T) {
	svc := &serviceImpl{}
	ctx := datascope.WithTenantScope(context.Background(), 42)

	enabled, err := svc.registryEnabledForTenant(ctx, &store.PluginRecord{
		PluginId:    "linapro-tenant-core",
		Installed:   statusflag.Installed.Int(),
		Status:      statusflag.EnabledValue.Int(),
		ScopeNature: pluginv1.ScopeNaturePlatformOnly.String(),
		InstallMode: pluginv1.InstallModeGlobal.String(),
	})
	if err != nil {
		t.Fatalf("expected platform-only global enablement check to succeed, got error: %v", err)
	}
	if !enabled {
		t.Fatal("expected platform-only global plugin to stay enabled in tenant context")
	}
}
