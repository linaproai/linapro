// This file verifies store-owned registry, release, startup snapshot, and
// runtime-upgrade projections.

package store_test

import (
	"context"
	"strings"
	"testing"

	_ "lina-core/pkg/dbdriver"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// TestGetRegistryReleaseFallsBackWhenReleasePointerIsDangling verifies that
// store reads tolerate registry rows whose release_id no longer points to an
// existing release row.
func TestGetRegistryReleaseFallsBackWhenReleasePointerIsDangling(t *testing.T) {
	var (
		ctx      = context.Background()
		svcs     = testutil.NewServices()
		pluginID = "acme-demo-dangling-release-pointer"
		version  = "9.9.9"
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := dao.SysPlugin.Ctx(ctx).Data(do.SysPlugin{
		PluginId:     pluginID,
		Name:         "Dangling Release Pointer Plugin",
		Version:      version,
		Type:         plugintypes.TypeDynamic.String(),
		Installed:    plugintypes.InstalledYes,
		Status:       plugintypes.StatusEnabled,
		DesiredState: plugintypes.LifecycleStateRuntimeEnabled.String(),
		CurrentState: plugintypes.LifecycleStateRuntimeEnabled.String(),
		Generation:   int64(1),
		ReleaseId:    987654321,
		ScopeNature:  plugintypes.ScopeNatureTenantAware.String(),
		InstallMode:  plugintypes.InstallModeTenantScoped.String(),
		ManifestPath: "runtime/acme-demo-dangling-release-pointer/plugin.yaml",
		Checksum:     "dangling-release-pointer",
		Remark:       "Dangling release pointer test plugin",
	}).InsertAndGetId(); err != nil {
		t.Fatalf("failed to insert plugin registry row: %v", err)
	}
	insertID, err := dao.SysPluginRelease.Ctx(ctx).Data(do.SysPluginRelease{
		PluginId:       pluginID,
		ReleaseVersion: version,
		Type:           plugintypes.TypeDynamic.String(),
		RuntimeKind:    protocol.RuntimeKindWasm,
		Status:         plugintypes.ReleaseStatusActive.String(),
		ManifestPath:   "runtime/acme-demo-dangling-release-pointer/plugin.yaml",
		PackagePath:    "runtime/acme-demo-dangling-release-pointer.wasm",
		Checksum:       "dangling-release-pointer",
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("failed to insert fallback plugin release row: %v", err)
	}

	registry, err := svcs.Store.GetRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup to succeed, got error: %v", err)
	}
	release, err := svcs.Store.GetRegistryRelease(ctx, registry)
	if err != nil {
		t.Fatalf("expected dangling release pointer to fall back to plugin version, got error: %v", err)
	}
	if release == nil {
		t.Fatalf("expected fallback release to be returned")
	}
	if release.Id != int(insertID) {
		t.Fatalf("expected fallback release id %d, got %d", insertID, release.Id)
	}
}

// TestStartupDataSnapshotReusesReleaseByIDAndVersion verifies one store
// snapshot backs both release lookup shapes and can be explicitly refreshed
// after the authority database row changes.
func TestStartupDataSnapshotReusesReleaseByIDAndVersion(t *testing.T) {
	var (
		ctx      = context.Background()
		svcs     = testutil.NewServices()
		pluginID = "acme-demo-release-snapshot-reuse"
		version  = "v0.1.0"
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest := &catalog.Manifest{
		ID:                 pluginID,
		Name:               "Release Snapshot Reuse",
		Version:            version,
		Type:               plugintypes.TypeDynamic.String(),
		ScopeNature:        plugintypes.ScopeNatureTenantAware.String(),
		DefaultInstallMode: plugintypes.InstallModeTenantScoped.String(),
	}
	registry, err := svcs.Store.SyncManifest(ctx, manifest)
	if err != nil {
		t.Fatalf("expected manifest sync to create registry and release, got error: %v", err)
	}
	if err = svcs.Store.SetPluginInstalled(ctx, pluginID, plugintypes.InstalledYes); err != nil {
		t.Fatalf("expected installed marker update to succeed, got error: %v", err)
	}
	registry, err = svcs.Store.GetRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup to succeed, got error: %v", err)
	}
	registry, err = svcs.Store.SyncRegistryReleaseReference(ctx, registry, manifest)
	if err != nil {
		t.Fatalf("expected release reference sync to succeed, got error: %v", err)
	}
	if registry == nil || registry.ReleaseId <= 0 {
		t.Fatalf("expected registry to point at release, got %#v", registry)
	}

	snapshotCtx, err := svcs.Store.WithStartupDataSnapshot(ctx)
	if err != nil {
		t.Fatalf("expected store snapshot to build, got error: %v", err)
	}
	byID, err := svcs.Store.GetReleaseByID(snapshotCtx, registry.ReleaseId)
	if err != nil {
		t.Fatalf("expected release lookup by id to succeed, got error: %v", err)
	}
	byVersion, err := svcs.Store.GetRelease(snapshotCtx, pluginID, version)
	if err != nil {
		t.Fatalf("expected release lookup by plugin version to succeed, got error: %v", err)
	}
	if byID == nil || byVersion == nil || byID.Id != byVersion.Id {
		t.Fatalf("expected both lookup shapes to return the same release, got byID=%#v byVersion=%#v", byID, byVersion)
	}

	const refreshedChecksum = "release-snapshot-refreshed"
	if _, err = dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{Id: registry.ReleaseId}).
		Data(do.SysPluginRelease{Checksum: refreshedChecksum}).
		Update(); err != nil {
		t.Fatalf("expected release checksum update to succeed, got error: %v", err)
	}
	staleByVersion, err := svcs.Store.GetRelease(snapshotCtx, pluginID, version)
	if err != nil {
		t.Fatalf("expected stale snapshot lookup to succeed, got error: %v", err)
	}
	if staleByVersion == nil || staleByVersion.Checksum == refreshedChecksum {
		t.Fatalf("expected request snapshot to remain stable before explicit refresh, got %#v", staleByVersion)
	}

	refreshed, err := svcs.Store.RefreshStartupReleaseByID(snapshotCtx, registry.ReleaseId)
	if err != nil {
		t.Fatalf("expected snapshot refresh to succeed, got error: %v", err)
	}
	if refreshed == nil || refreshed.Checksum != refreshedChecksum {
		t.Fatalf("expected refreshed release checksum %s, got %#v", refreshedChecksum, refreshed)
	}
	refreshedByVersion, err := svcs.Store.GetRelease(snapshotCtx, pluginID, version)
	if err != nil {
		t.Fatalf("expected refreshed version lookup to succeed, got error: %v", err)
	}
	if refreshedByVersion == nil || refreshedByVersion.Checksum != refreshedChecksum {
		t.Fatalf("expected version index to use refreshed release checksum %s, got %#v", refreshedChecksum, refreshedByVersion)
	}
}

// TestParseManifestSnapshotRejectsUnsupportedHostServiceSnapshots verifies
// persisted snapshots use the same strict host-service names as fresh manifests.
func TestParseManifestSnapshotRejectsUnsupportedHostServiceSnapshots(t *testing.T) {
	svcs := testutil.NewServices()
	testCases := []struct {
		name     string
		snapshot string
	}{
		{
			name: "hostruntime",
			snapshot: `
id: acme-demo-reject-hostruntime-snapshot
requestedHostServices:
  - service: hostruntime
    methods:
      - get
    keys:
      - workspace.basePath
`,
		},
		{
			name: "standalone config",
			snapshot: `
id: acme-demo-reject-config-snapshot
requestedHostServices:
  - service: config
    methods:
      - get
`,
		},
		{
			name: "standalone cron",
			snapshot: `
id: acme-demo-reject-cron-snapshot
requestedHostServices:
  - service: cron
    methods:
      - cron.register
`,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := svcs.Store.ParseManifestSnapshot(testCase.snapshot); err == nil {
				t.Fatalf("expected unsupported host-service snapshot %q to be rejected", testCase.name)
			}
		})
	}
}

// TestRuntimeUpgradeStateReportsExplicitRunningMarker verifies management
// projections expose upgrade_running while an explicit runtime upgrade is in progress.
func TestRuntimeUpgradeStateReportsExplicitRunningMarker(t *testing.T) {
	var (
		ctx        = context.Background()
		svcs       = testutil.NewServices()
		pluginID   = "acme-demo-runtime-upgrade-running-marker"
		oldVersion = "v0.1.0"
		newVersion = "v0.2.0"
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	oldManifest := &catalog.Manifest{
		ID:                 pluginID,
		Name:               "Runtime Upgrade Running Marker",
		Version:            oldVersion,
		Type:               plugintypes.TypeDynamic.String(),
		ScopeNature:        plugintypes.ScopeNatureTenantAware.String(),
		DefaultInstallMode: plugintypes.InstallModeTenantScoped.String(),
	}
	registry, err := svcs.Store.SyncManifest(ctx, oldManifest)
	if err != nil {
		t.Fatalf("expected old manifest sync to succeed, got error: %v", err)
	}
	oldRelease, err := svcs.Store.GetRelease(ctx, pluginID, oldVersion)
	if err != nil {
		t.Fatalf("expected old release lookup to succeed, got error: %v", err)
	}
	if oldRelease == nil {
		t.Fatal("expected old release row")
	}
	if err = svcs.Store.SetPluginInstalled(ctx, pluginID, plugintypes.InstalledYes); err != nil {
		t.Fatalf("expected installed state update to succeed, got error: %v", err)
	}
	registry, err = svcs.Store.GetRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup after install marker to succeed, got error: %v", err)
	}
	registry, err = svcs.Store.SyncRegistryReleaseReference(ctx, registry, oldManifest)
	if err != nil {
		t.Fatalf("expected registry release reference sync to succeed, got error: %v", err)
	}
	if err = svcs.Store.UpdateReleaseState(ctx, oldRelease.Id, plugintypes.ReleaseStatusInstalled, ""); err != nil {
		t.Fatalf("expected old release state update to succeed, got error: %v", err)
	}

	newManifest := &catalog.Manifest{
		ID:                 pluginID,
		Name:               "Runtime Upgrade Running Marker",
		Version:            newVersion,
		Type:               plugintypes.TypeDynamic.String(),
		ScopeNature:        plugintypes.ScopeNatureTenantAware.String(),
		DefaultInstallMode: plugintypes.InstallModeTenantScoped.String(),
	}
	if _, err = svcs.Store.SyncManifest(ctx, newManifest); err != nil {
		t.Fatalf("expected new manifest sync to succeed, got error: %v", err)
	}
	if err = svcs.Store.SetRegistryRuntimeState(ctx, pluginID, store.RuntimeStatePatch{
		CurrentState: plugintypes.RuntimeUpgradeStateUpgradeRunning.String(),
	}); err != nil {
		t.Fatalf("expected running marker update to succeed, got error: %v", err)
	}

	registry, err = svcs.Store.GetRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup to succeed, got error: %v", err)
	}
	projection, err := svcs.Store.BuildRuntimeUpgradeState(ctx, registry, newManifest)
	if err != nil {
		t.Fatalf("expected runtime state projection to succeed, got error: %v", err)
	}
	if projection.State != plugintypes.RuntimeUpgradeStateUpgradeRunning {
		t.Fatalf("expected upgrade_running projection, got %#v", projection)
	}
}

// TestRuntimeUpgradeStateBlocksFailedTargetRelease verifies failed releases do
// not project as normal runtime state even when their semantic version matches
// the current effective registry version.
func TestRuntimeUpgradeStateBlocksFailedTargetRelease(t *testing.T) {
	var (
		ctx      = context.Background()
		svcs     = testutil.NewServices()
		pluginID = "acme-demo-runtime-upgrade-failed-target"
		version  = "v0.1.0"
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest := &catalog.Manifest{
		ID:                 pluginID,
		Name:               "Runtime Upgrade Failed Target",
		Version:            version,
		Type:               plugintypes.TypeDynamic.String(),
		ScopeNature:        plugintypes.ScopeNatureTenantAware.String(),
		DefaultInstallMode: plugintypes.InstallModeTenantScoped.String(),
	}
	registry, err := svcs.Store.SyncManifest(ctx, manifest)
	if err != nil {
		t.Fatalf("expected manifest sync to succeed, got error: %v", err)
	}
	if err = svcs.Store.SetPluginInstalled(ctx, pluginID, plugintypes.InstalledYes); err != nil {
		t.Fatalf("expected installed marker update to succeed, got error: %v", err)
	}
	if err = svcs.Store.SetPluginStatus(ctx, pluginID, plugintypes.StatusEnabled); err != nil {
		t.Fatalf("expected enabled marker update to succeed, got error: %v", err)
	}
	registry, err = svcs.Store.GetRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup to succeed, got error: %v", err)
	}
	registry, err = svcs.Store.SyncRegistryReleaseReference(ctx, registry, manifest)
	if err != nil {
		t.Fatalf("expected registry release reference sync to succeed, got error: %v", err)
	}
	if registry == nil || registry.ReleaseId <= 0 {
		t.Fatalf("expected registry to point at release, got %#v", registry)
	}
	if err = svcs.Store.UpdateReleaseState(ctx, registry.ReleaseId, plugintypes.ReleaseStatusFailed, ""); err != nil {
		t.Fatalf("expected failed release state update to succeed, got error: %v", err)
	}

	release, err := svcs.Store.GetReleaseByID(ctx, registry.ReleaseId)
	if err != nil {
		t.Fatalf("expected release lookup to succeed, got error: %v", err)
	}
	projection, err := svcs.Store.BuildRuntimeUpgradeState(ctx, registry, manifest)
	if err != nil {
		t.Fatalf("expected runtime state projection to succeed, got error: %v", err)
	}
	if projection.State != plugintypes.RuntimeUpgradeStateUpgradeFailed {
		t.Fatalf("expected failed target release to block runtime state, got %#v", projection)
	}
	if projection.LastFailure == nil || projection.LastFailure.ReleaseID != release.Id {
		t.Fatalf("expected failed release diagnostic for release %d, got %#v", release.Id, projection.LastFailure)
	}
}

// TestRuntimeUpgradeStateMapsUnifiedFailurePhaseLedger verifies synthetic
// upgrade failure rows written by the unified upgrade owner drive the same
// phase and message-key projection for dynamic release-switch and cache phases.
func TestRuntimeUpgradeStateMapsUnifiedFailurePhaseLedger(t *testing.T) {
	cases := []struct {
		name          string
		migrationKey  string
		expectedPhase plugintypes.RuntimeUpgradeFailurePhase
	}{
		{
			name:          "release switch",
			migrationKey:  "upgrade-phase-release_switch",
			expectedPhase: plugintypes.RuntimeUpgradeFailurePhaseReleaseSwitch,
		},
		{
			name:          "cache invalidation",
			migrationKey:  "upgrade-phase-cache_invalidation",
			expectedPhase: plugintypes.RuntimeUpgradeFailurePhaseCacheInvalidation,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var (
				ctx      = context.Background()
				svcs     = testutil.NewServices()
				pluginID = "acme-demo-runtime-upgrade-" + strings.ReplaceAll(tt.name, " ", "-")
				version  = "v0.1.0"
			)

			testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
			t.Cleanup(func() {
				testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
			})

			manifest := &catalog.Manifest{
				ID:                 pluginID,
				Name:               "Runtime Upgrade Failed Phase",
				Version:            version,
				Type:               plugintypes.TypeDynamic.String(),
				ScopeNature:        plugintypes.ScopeNatureTenantAware.String(),
				DefaultInstallMode: plugintypes.InstallModeTenantScoped.String(),
			}
			registry, err := svcs.Store.SyncManifest(ctx, manifest)
			if err != nil {
				t.Fatalf("expected manifest sync to succeed, got error: %v", err)
			}
			if err = svcs.Store.SetPluginInstalled(ctx, pluginID, plugintypes.InstalledYes); err != nil {
				t.Fatalf("expected installed marker update to succeed, got error: %v", err)
			}
			if err = svcs.Store.SetPluginStatus(ctx, pluginID, plugintypes.StatusEnabled); err != nil {
				t.Fatalf("expected enabled marker update to succeed, got error: %v", err)
			}
			registry, err = svcs.Store.GetRegistry(ctx, pluginID)
			if err != nil {
				t.Fatalf("expected registry lookup to succeed, got error: %v", err)
			}
			registry, err = svcs.Store.SyncRegistryReleaseReference(ctx, registry, manifest)
			if err != nil {
				t.Fatalf("expected registry release reference sync to succeed, got error: %v", err)
			}
			if registry == nil || registry.ReleaseId <= 0 {
				t.Fatalf("expected registry to point at release, got %#v", registry)
			}
			if err = svcs.Store.UpdateReleaseState(ctx, registry.ReleaseId, plugintypes.ReleaseStatusFailed, ""); err != nil {
				t.Fatalf("expected failed release state update to succeed, got error: %v", err)
			}
			if _, err = dao.SysPluginMigration.Ctx(ctx).Data(do.SysPluginMigration{
				PluginId:       pluginID,
				ReleaseId:      registry.ReleaseId,
				Phase:          plugintypes.MigrationDirectionUpgrade.String(),
				MigrationKey:   tt.migrationKey,
				ExecutionOrder: 0,
				Checksum:       tt.migrationKey,
				Status:         plugintypes.MigrationExecutionStatusFailed.String(),
				ErrorMessage:   tt.name + " failed",
			}).Insert(); err != nil {
				t.Fatalf("expected failure migration insert to succeed, got error: %v", err)
			}

			projection, err := svcs.Store.BuildRuntimeUpgradeState(ctx, registry, manifest)
			if err != nil {
				t.Fatalf("expected runtime state projection to succeed, got error: %v", err)
			}
			if projection.LastFailure == nil {
				t.Fatalf("expected last failure projection, got %#v", projection)
			}
			if projection.LastFailure.Phase != tt.expectedPhase {
				t.Fatalf("expected phase %s, got %#v", tt.expectedPhase, projection.LastFailure)
			}
			if projection.LastFailure.MessageKey != store.RuntimeUpgradeFailureMessageKeyMigrationFailed {
				t.Fatalf("expected unified migration failure key, got %#v", projection.LastFailure)
			}
		})
	}
}
