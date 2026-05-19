// This file verifies the host-governed plugin consumer frontend projection
// without requiring plugin lifecycle database fixtures or concrete plugins.

package plugin

import (
	"testing"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
)

// TestBuildConsumerSurfaceSnapshotAggregatesHostInputs verifies consumer
// frontend, enablement, version, and tenant governance metadata are grouped by plugin.
func TestBuildConsumerSurfaceSnapshotAggregatesHostInputs(t *testing.T) {
	supportsMultiTenant := true
	manifests := []*catalog.Manifest{
		{
			ID:                  "mall",
			Version:             "0.1.0",
			Type:                catalog.TypeSource.String(),
			ScopeNature:         catalog.ScopeNatureTenantAware.String(),
			SupportsMultiTenant: &supportsMultiTenant,
			DefaultInstallMode:  catalog.InstallModeTenantScoped.String(),
		},
		{
			ID:                 "admin-only",
			Version:            "0.1.0",
			Type:               catalog.TypeSource.String(),
			ScopeNature:        catalog.ScopeNaturePlatformOnly.String(),
			DefaultInstallMode: catalog.InstallModeGlobal.String(),
		},
	}
	registries := []*entity.SysPlugin{
		{
			PluginId:    "mall",
			Version:     "0.2.0",
			ScopeNature: catalog.ScopeNatureTenantAware.String(),
			InstallMode: catalog.InstallModeTenantScoped.String(),
		},
	}
	frontendIndex := &sourceConsumerFrontendResourceIndex{
		mounts: []*sourceConsumerFrontendMountEntry{
			{
				pluginID:    "mall",
				version:     "0.2.0",
				mountPath:   "/mall",
				index:       "index.html",
				spaFallback: true,
				assets: map[string]struct{}{
					"frontend/consumer/index.html":     {},
					"frontend/consumer/assets/app.js":  {},
					"frontend/consumer/assets/app.css": {},
				},
			},
		},
	}

	snapshot := buildConsumerSurfaceSnapshot(
		manifests,
		registries,
		frontendIndex,
		map[string]bool{"mall": true, "admin-only": true},
	)

	if snapshot == nil {
		t.Fatalf("expected consumer frontend snapshot")
	}
	if len(snapshot.Plugins) != 1 {
		t.Fatalf("expected only consumer-frontend capable plugin, got %#v", snapshot.Plugins)
	}
	got := snapshot.Plugins[0]
	if got.PluginID != "mall" || got.Version != "0.2.0" {
		t.Fatalf("unexpected plugin identity: %#v", got)
	}
	if !got.TenantAware ||
		got.ScopeNature != catalog.ScopeNatureTenantAware.String() ||
		got.DefaultInstallMode != catalog.InstallModeTenantScoped.String() {
		t.Fatalf("unexpected tenant governance projection: %#v", got)
	}
	if got.ConsumerFrontend == nil ||
		got.ConsumerFrontend.MountPath != "/mall" ||
		got.ConsumerFrontend.AssetCount != 3 ||
		!got.ConsumerFrontend.SPAFallback {
		t.Fatalf("unexpected frontend snapshot: %#v", got.ConsumerFrontend)
	}
}

// TestBuildConsumerSurfaceSnapshotSortsPlugins verifies governance output
// remains deterministic for review and future API exposure.
func TestBuildConsumerSurfaceSnapshotSortsPlugins(t *testing.T) {
	snapshot := buildConsumerSurfaceSnapshot(
		nil,
		nil,
		&sourceConsumerFrontendResourceIndex{mounts: []*sourceConsumerFrontendMountEntry{
			{pluginID: "portal", version: "0.1.0", mountPath: "/portal", index: "index.html"},
			{pluginID: "mall", version: "0.1.0", mountPath: "/mall", index: "index.html"},
		}},
		nil,
	)

	if len(snapshot.Plugins) != 2 {
		t.Fatalf("expected two plugin snapshots, got %#v", snapshot.Plugins)
	}
	if snapshot.Plugins[0].PluginID != "mall" || snapshot.Plugins[1].PluginID != "portal" {
		t.Fatalf("expected sorted plugins, got %#v", snapshot.Plugins)
	}
}
