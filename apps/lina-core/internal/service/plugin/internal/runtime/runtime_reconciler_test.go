// This file covers dynamic-plugin reconciler safety boundaries: per-plugin
// locks, stale transient-state recovery, panic isolation, and rollback
// diagnostics.

package runtime

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/cachecoord/revisionctrl"
	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/coordination"
	"lina-core/internal/service/locker"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/migration"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/internal/service/session"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// TestRuntimeReconcilerLockSerializesSamePlugin verifies a held per-plugin lock
// causes the next clustered reconcile attempt for the same plugin to skip
// primary lifecycle side effects.
func TestRuntimeReconcilerLockSerializesSamePlugin(t *testing.T) {
	ctx := context.Background()
	coordSvc := coordination.NewMemory(nil)
	locker.ConfigureCoordination(coordSvc)
	t.Cleanup(func() {
		locker.ConfigureCoordination(nil)
	})

	pluginID := "plugin-dev-reconciler-lock-same"
	held, ok, err := locker.New().Lock(ctx, runtimeReconcilerLockName(pluginID), "node-a", "test", time.Minute)
	if err != nil || !ok || held == nil {
		t.Fatalf("expected initial per-plugin lock, ok=%v err=%v", ok, err)
	}
	defer func() {
		if unlockErr := held.Unlock(ctx); unlockErr != nil {
			t.Fatalf("unlock initial per-plugin lock failed: %v", unlockErr)
		}
	}()

	service := &serviceImpl{
		topology:          reconcilerRevisionTestTopology{cluster: true, primary: true, nodeID: "node-b"},
		reconcilerLockSvc: locker.New(),
	}
	locked, unlock, err := service.lockRuntimeReconcilePlugin(ctx, pluginID)
	if err != nil {
		t.Fatalf("second per-plugin lock attempt failed: %v", err)
	}
	defer unlock()
	if locked {
		t.Fatal("expected same-plugin lock attempt to skip while another holder owns it")
	}
}

// TestRuntimeReconcilerLockNamesArePerPlugin verifies two different plugin IDs
// use independent distributed locks.
func TestRuntimeReconcilerLockNamesArePerPlugin(t *testing.T) {
	ctx := context.Background()
	coordSvc := coordination.NewMemory(nil)
	locker.ConfigureCoordination(coordSvc)
	t.Cleanup(func() {
		locker.ConfigureCoordination(nil)
	})

	held, ok, err := locker.New().Lock(ctx, runtimeReconcilerLockName("plugin-dev-reconciler-lock-a"), "node-a", "test", time.Minute)
	if err != nil || !ok || held == nil {
		t.Fatalf("expected initial plugin A lock, ok=%v err=%v", ok, err)
	}
	defer func() {
		if unlockErr := held.Unlock(ctx); unlockErr != nil {
			t.Fatalf("unlock plugin A lock failed: %v", unlockErr)
		}
	}()

	service := &serviceImpl{
		topology:          reconcilerRevisionTestTopology{cluster: true, primary: true, nodeID: "node-b"},
		reconcilerLockSvc: locker.New(),
	}
	locked, unlock, err := service.lockRuntimeReconcilePlugin(ctx, "plugin-dev-reconciler-lock-b")
	if err != nil {
		t.Fatalf("plugin B per-plugin lock attempt failed: %v", err)
	}
	defer unlock()
	if !locked {
		t.Fatal("expected plugin B lock to be independent from plugin A")
	}
}

// TestRequiredRuntimeReconcilerLockReturnsConflict verifies explicit lifecycle
// callers do not observe success when another node owns the per-plugin lock.
func TestRequiredRuntimeReconcilerLockReturnsConflict(t *testing.T) {
	ctx := context.Background()
	coordSvc := coordination.NewMemory(nil)
	locker.ConfigureCoordination(coordSvc)
	t.Cleanup(func() {
		locker.ConfigureCoordination(nil)
	})

	const pluginID = "plugin-dev-reconciler-lock-required"
	held, ok, err := locker.New().Lock(ctx, runtimeReconcilerLockName(pluginID), "node-a", "test", time.Minute)
	if err != nil || !ok || held == nil {
		t.Fatalf("expected initial per-plugin lock, ok=%v err=%v", ok, err)
	}
	defer func() {
		if unlockErr := held.Unlock(ctx); unlockErr != nil {
			t.Fatalf("unlock initial per-plugin lock failed: %v", unlockErr)
		}
	}()

	service := &serviceImpl{
		topology:          reconcilerRevisionTestTopology{cluster: true, primary: true, nodeID: "node-b"},
		reconcilerLockSvc: locker.New(),
	}
	called := false
	err = service.reconcilePrimaryPluginWithRequiredLock(
		ctx,
		&store.PluginRecord{PluginId: pluginID, Type: plugintypes.TypeDynamic.String()},
		func(context.Context, *store.PluginRecord) error {
			called = true
			return nil
		},
	)
	if err == nil {
		t.Fatal("expected required lock conflict to return an error")
	}
	if called {
		t.Fatal("expected required lock conflict to skip lifecycle callback")
	}
}

// TestRecoverStaleReconcilingRestoresOldState verifies abandoned transient host
// state is restored to the stable state derived from installed/enabled flags.
func TestRecoverStaleReconcilingRestoresOldState(t *testing.T) {
	services := newRuntimeSafetyServices()
	ctx := context.Background()

	const pluginID = "plugin-dev-stale-reconciling-old"
	manifest := newRuntimeSafetyManifest(pluginID, "Stale Reconciling Old Plugin", "v0.9.4", nil)

	cleanupRuntimeSafetyPluginRows(t, ctx, pluginID)
	t.Cleanup(func() {
		cleanupRuntimeSafetyPluginRows(t, ctx, pluginID)
	})

	registry, err := services.Store.SyncManifest(ctx, manifest)
	if err != nil {
		t.Fatalf("expected manifest sync to succeed, got error: %v", err)
	}
	if _, err = dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: pluginID}).
		Data(do.SysPlugin{
			Installed:    plugintypes.InstalledYes,
			Status:       plugintypes.StatusEnabled,
			DesiredState: plugintypes.HostStateEnabled.String(),
			CurrentState: plugintypes.HostStateReconciling.String(),
		}).
		Update(); err != nil {
		t.Fatalf("seed reconciling registry state failed: %v", err)
	}
	staleUpdatedAt := time.Now().Add(-runtimeReconcilerStaleReconcilingAfter - time.Minute)
	if err = seedRegistryUpdatedAt(ctx, pluginID, staleUpdatedAt); err != nil {
		t.Fatalf("seed stale updated_at failed: %v", err)
	}

	registry, err = services.Store.GetRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("reload registry failed: %v", err)
	}
	restored, err := services.Runtime.recoverStaleReconciling(ctx, registry)
	if err != nil {
		t.Fatalf("recover stale reconciling failed: %v", err)
	}
	if restored == nil || restored.CurrentState != plugintypes.HostStateEnabled.String() {
		t.Fatalf("expected stale reconciling to restore enabled state, got %#v", restored)
	}
}

// TestRecoverFreshReconcilingSkipsPrimaryWork verifies fresh transient state is
// not reset by another reconcile attempt.
func TestRecoverFreshReconcilingSkipsPrimaryWork(t *testing.T) {
	services := newRuntimeSafetyServices()
	ctx := context.Background()

	const pluginID = "plugin-dev-stale-reconciling-fresh"
	manifest := newRuntimeSafetyManifest(pluginID, "Fresh Reconciling Plugin", "v0.9.5", nil)

	cleanupRuntimeSafetyPluginRows(t, ctx, pluginID)
	t.Cleanup(func() {
		cleanupRuntimeSafetyPluginRows(t, ctx, pluginID)
	})

	if _, err := services.Store.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected manifest sync to succeed, got error: %v", err)
	}
	if _, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: pluginID}).
		Data(do.SysPlugin{
			Installed:    plugintypes.InstalledYes,
			Status:       plugintypes.StatusEnabled,
			DesiredState: plugintypes.HostStateEnabled.String(),
			CurrentState: plugintypes.HostStateReconciling.String(),
		}).
		Update(); err != nil {
		t.Fatalf("seed reconciling registry state failed: %v", err)
	}

	registry, err := services.Store.GetRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("reload registry failed: %v", err)
	}
	restored, err := services.Runtime.recoverStaleReconciling(ctx, registry)
	if err != nil {
		t.Fatalf("recover fresh reconciling failed: %v", err)
	}
	if restored != nil {
		t.Fatalf("expected fresh reconciling state to be skipped, got %#v", restored)
	}
	latest, err := services.Store.GetRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("reload latest registry failed: %v", err)
	}
	if latest.CurrentState != plugintypes.HostStateReconciling.String() {
		t.Fatalf("expected fresh reconciling state to remain, got %s", latest.CurrentState)
	}
}

// TestRunReconcilerTickSafelyRecoversPanic verifies tick-level panic recovery
// keeps the caller from seeing the panic.
func TestRunReconcilerTickSafelyRecoversPanic(t *testing.T) {
	service := &serviceImpl{
		topology: reconcilerRevisionTestTopology{cluster: true, primary: true, nodeID: "node-primary"},
		reconcilerRevisionCtrl: newTestReconcilerRevisionController(
			&panicReconcilerRevisionCacheCoord{},
			revisionctrl.NewObservedRevision(),
		),
	}
	service.runReconcilerTickSafely(context.Background())
}

// TestRollbackInstallOrUpgradeReturnsRollbackDiagnostics verifies rollback
// errors are included in the authoritative returned error and registry state.
func TestRollbackInstallOrUpgradeReturnsRollbackDiagnostics(t *testing.T) {
	services := newRuntimeSafetyServices()
	ctx := context.Background()

	const pluginID = "plugin-dev-rollback-diagnostics"
	manifest := newRuntimeSafetyManifest(
		pluginID,
		"Rollback Diagnostics Plugin",
		"v0.9.6",
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-plugin-dev-rollback-diagnostics.sql",
				Content: "INSERT INTO plugin_dev_missing_rollback_table(id) VALUES (1);",
			},
		},
	)

	cleanupRuntimeSafetyPluginRows(t, ctx, pluginID)
	t.Cleanup(func() {
		cleanupRuntimeSafetyPluginRows(t, ctx, pluginID)
	})

	registry, err := services.Store.SyncManifest(ctx, manifest)
	if err != nil {
		t.Fatalf("expected manifest sync to succeed, got error: %v", err)
	}

	failedRelease, err := services.Store.GetRelease(ctx, pluginID, manifest.Version)
	if err != nil {
		t.Fatalf("expected release lookup to succeed, got error: %v", err)
	}
	if failedRelease == nil {
		t.Fatal("expected synced release row")
	}

	reconcileErr := tagDynamicUpgradeFailure(
		plugintypes.RuntimeUpgradeFailurePhaseSQL,
		errors.New("original install failure"),
	)
	resultErr := services.Runtime.rollbackInstallOrUpgrade(
		ctx,
		registry,
		nil,
		manifest,
		failedRelease.Id,
		reconcileErr,
	)
	if resultErr == nil {
		t.Fatal("expected rollback diagnostics error")
	}
	if !strings.Contains(resultErr.Error(), "original install failure") ||
		!strings.Contains(resultErr.Error(), "001-plugin-dev-rollback-diagnostics.sql") {
		t.Fatalf("expected original and rollback SQL diagnostics, got: %v", resultErr)
	}
	if phase := DynamicUpgradeFailurePhase(resultErr); phase != plugintypes.RuntimeUpgradeFailurePhaseSQL {
		t.Fatalf("expected rollback diagnostics to preserve original SQL phase, got %s", phase)
	}

	releaseAfter, err := services.Store.GetReleaseByID(ctx, failedRelease.Id)
	if err != nil {
		t.Fatalf("expected release reload to succeed, got error: %v", err)
	}
	if releaseAfter.Status != plugintypes.ReleaseStatusFailed.String() {
		t.Fatalf("expected failed release status, got %s", releaseAfter.Status)
	}
	registryAfter, err := services.Store.GetRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry reload to succeed, got error: %v", err)
	}
	if registryAfter.CurrentState != plugintypes.HostStateFailed.String() {
		t.Fatalf("expected registry failed state, got %s", registryAfter.CurrentState)
	}
}

// runtimeSafetyServices groups the narrow service set needed by same-package
// runtime safety tests without importing testutil, which depends on runtime.
type runtimeSafetyServices struct {
	Catalog catalog.Service
	Store   store.Service
	Runtime *serviceImpl
}

// runtimeSafetyLifecycleReconciler delegates lifecycle reconciliation to runtime
// after the runtime service is constructed.
type runtimeSafetyLifecycleReconciler struct {
	service *serviceImpl
}

// ReconcileDynamicPluginRequest delegates desired-state transitions.
func (p *runtimeSafetyLifecycleReconciler) ReconcileDynamicPluginRequest(
	ctx context.Context,
	pluginID string,
	desiredState string,
	options DynamicReconcileOptions,
) error {
	if p == nil || p.service == nil {
		return nil
	}
	return p.service.ReconcileDynamicPluginRequest(ctx, pluginID, desiredState, options)
}

// ShouldRefreshInstalledDynamicRelease delegates refresh decisions.
func (p *runtimeSafetyLifecycleReconciler) ShouldRefreshInstalledDynamicRelease(ctx context.Context, registry interface{}, manifest *catalog.Manifest) bool {
	return p != nil && p.service != nil && p.service.ShouldRefreshInstalledDynamicRelease(ctx, registry, manifest)
}

// EnsureRuntimeArtifactAvailable delegates artifact availability checks.
func (p *runtimeSafetyLifecycleReconciler) EnsureRuntimeArtifactAvailable(manifest *catalog.Manifest, actionLabel string) error {
	if p == nil || p.service == nil {
		return nil
	}
	return p.service.EnsureRuntimeArtifactAvailable(manifest, actionLabel)
}

// runtimeSafetyStorageCleanupProvider returns no storage cleanup directory for
// same-package safety tests.
type runtimeSafetyStorageCleanupProvider struct{}

// StorageCleanupServices returns no capability directory.
func (runtimeSafetyStorageCleanupProvider) StorageCleanupServices() capability.Services {
	return nil
}

// newRuntimeSafetyServices wires catalog, lifecycle, and runtime services with
// only the dependencies exercised by reconciler safety tests.
func newRuntimeSafetyServices() *runtimeSafetyServices {
	configProvider := configsvc.New()
	catalogSvc := catalog.New(configProvider)
	topology := reconcilerRevisionTestTopology{cluster: false, primary: true, nodeID: "test-node"}
	storeSvc := store.New(catalogSvc, topology)
	migrationSvc := migration.New(catalogSvc, storeSvc)
	lifecycleHook := &runtimeSafetyLifecycleReconciler{}
	runtimeSvc := New(
		catalogSvc,
		storeSvc,
		migrationSvc,
		nil,
		nil,
		nil,
		locker.New(),
		topology,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		session.NewDBStore(),
		testRoleAccessProjector{},
		nil,
		nil,
		nil,
		runtimeSafetyStorageCleanupProvider{},
		nil,
	).(*serviceImpl)
	lifecycleHook.service = runtimeSvc
	runtimeSvc.reconcilerRevisionCtrl = newTestReconcilerRevisionController(
		&reconcilerRevisionCacheCoord{},
		revisionctrl.NewObservedRevision(),
	)
	return &runtimeSafetyServices{
		Catalog: catalogSvc,
		Store:   storeSvc,
		Runtime: runtimeSvc,
	}
}

// newRuntimeSafetyManifest builds the minimal dynamic manifest required for
// registry, release, SQL, and rollback diagnostics tests.
func newRuntimeSafetyManifest(
	pluginID string,
	name string,
	version string,
	installSQLAssets []*catalog.ArtifactSQLAsset,
) *catalog.Manifest {
	supportsMultiTenant := true
	artifact := &catalog.ArtifactSpec{
		Path:             pluginID + ".wasm",
		Checksum:         pluginID + "-checksum-" + version,
		RuntimeKind:      protocol.RuntimeKindWasm,
		ABIVersion:       protocol.SupportedABIVersion,
		SQLAssetCount:    len(installSQLAssets),
		InstallSQLAssets: installSQLAssets,
		Manifest: &catalog.ArtifactManifest{
			ID:                  pluginID,
			Name:                name,
			Version:             version,
			Type:                plugintypes.TypeDynamic.String(),
			ScopeNature:         plugintypes.ScopeNatureTenantAware.String(),
			SupportsMultiTenant: &supportsMultiTenant,
			DefaultInstallMode:  plugintypes.InstallModeTenantScoped.String(),
		},
	}
	return &catalog.Manifest{
		ID:                  pluginID,
		Name:                name,
		Version:             version,
		Type:                plugintypes.TypeDynamic.String(),
		ScopeNature:         plugintypes.ScopeNatureTenantAware.String(),
		SupportsMultiTenant: &supportsMultiTenant,
		DefaultInstallMode:  plugintypes.InstallModeTenantScoped.String(),
		RuntimeArtifact:     artifact,
	}
}

// cleanupRuntimeSafetyPluginRows removes governance rows created by tests in
// dependency order so each test stays isolated and rerunnable.
func cleanupRuntimeSafetyPluginRows(t *testing.T, ctx context.Context, pluginID string) {
	t.Helper()

	if _, err := dao.SysPluginNodeState.Ctx(ctx).
		Unscoped().
		Where(dao.SysPluginNodeState.Columns().PluginId, pluginID).
		Delete(); err != nil {
		t.Fatalf("failed to delete sys_plugin_node_state rows for %s: %v", pluginID, err)
	}
	if _, err := dao.SysPluginResourceRef.Ctx(ctx).
		Unscoped().
		Where(dao.SysPluginResourceRef.Columns().PluginId, pluginID).
		Delete(); err != nil {
		t.Fatalf("failed to delete sys_plugin_resource_ref rows for %s: %v", pluginID, err)
	}
	if _, err := dao.SysPluginState.Ctx(ctx).
		Unscoped().
		Where(dao.SysPluginState.Columns().PluginId, pluginID).
		Delete(); err != nil {
		t.Fatalf("failed to delete sys_plugin_state rows for %s: %v", pluginID, err)
	}
	if _, err := dao.SysPluginMigration.Ctx(ctx).
		Unscoped().
		Where(dao.SysPluginMigration.Columns().PluginId, pluginID).
		Delete(); err != nil {
		t.Fatalf("failed to delete sys_plugin_migration rows for %s: %v", pluginID, err)
	}
	if _, err := dao.SysPluginRelease.Ctx(ctx).
		Unscoped().
		Where(dao.SysPluginRelease.Columns().PluginId, pluginID).
		Delete(); err != nil {
		t.Fatalf("failed to delete sys_plugin_release rows for %s: %v", pluginID, err)
	}
	if _, err := dao.SysPlugin.Ctx(ctx).
		Unscoped().
		Where(dao.SysPlugin.Columns().PluginId, pluginID).
		Delete(); err != nil {
		t.Fatalf("failed to delete sys_plugin rows for %s: %v", pluginID, err)
	}
}

// panicReconcilerRevisionCacheCoord forces a background tick panic before any
// database work so panic recovery can be tested cheaply.
type panicReconcilerRevisionCacheCoord struct{}

// ConfigureDomain is unused by panic tests.
func (p *panicReconcilerRevisionCacheCoord) ConfigureDomain(_ cachecoord.DomainSpec) error {
	return nil
}

// MarkChanged is unused by panic tests.
func (p *panicReconcilerRevisionCacheCoord) MarkChanged(context.Context, cachecoord.Domain, cachecoord.Scope, cachecoord.ChangeReason) (int64, error) {
	return 0, nil
}

// MarkTenantChanged is unused by panic tests.
func (p *panicReconcilerRevisionCacheCoord) MarkTenantChanged(context.Context, cachecoord.Domain, cachecoord.Scope, cachecoord.InvalidationScope, cachecoord.ChangeReason) (int64, error) {
	return 0, nil
}

// EnsureFresh is unused by panic tests.
func (p *panicReconcilerRevisionCacheCoord) EnsureFresh(context.Context, cachecoord.Domain, cachecoord.Scope, cachecoord.Refresher) (int64, error) {
	return 0, nil
}

// CurrentRevision panics to exercise runReconcilerTickSafely.
func (p *panicReconcilerRevisionCacheCoord) CurrentRevision(context.Context, cachecoord.Domain, cachecoord.Scope) (int64, error) {
	panic("reconciler revision panic")
}

// Snapshot is unused by panic tests.
func (p *panicReconcilerRevisionCacheCoord) Snapshot(context.Context) ([]cachecoord.SnapshotItem, error) {
	return nil, nil
}

// seedRegistryUpdatedAt moves updated_at in tests that need stale state.
func seedRegistryUpdatedAt(ctx context.Context, pluginID string, updatedAt time.Time) error {
	_, err := g.DB().Exec(ctx, `UPDATE sys_plugin SET updated_at = ? WHERE plugin_id = ?`, updatedAt, pluginID)
	return err
}
