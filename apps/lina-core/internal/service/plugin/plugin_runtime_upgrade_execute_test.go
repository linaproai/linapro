// This file covers explicit plugin runtime-upgrade execution paths.

package plugin

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/coordination"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/pluginbridge"
)

// TestExecuteRuntimeUpgradeRequiresConfirmation verifies the side-effecting
// upgrade endpoint rejects requests until the operator explicitly confirms.
func TestExecuteRuntimeUpgradeRequiresConfirmation(t *testing.T) {
	service := newTestService()
	_, err := service.ExecuteRuntimeUpgrade(context.Background(), "plugin-upgrade-unconfirmed", RuntimeUpgradeOptions{})
	if !bizerr.Is(err, CodePluginRuntimeUpgradeConfirmationRequired) {
		t.Fatalf("expected confirmation-required bizerr, got %v", err)
	}
}

// TestExecuteRuntimeUpgradeRejectsNormalPlugin verifies execution re-reads
// server state and refuses plugins that are no longer pending upgrade.
func TestExecuteRuntimeUpgradeRejectsNormalPlugin(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-runtime-upgrade-normal"
		version  = "v0.1.0"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Runtime Upgrade Normal Plugin",
		version,
		nil,
		nil,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic manifest sync to succeed, got error: %v", err)
	}
	if _, err = service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected dynamic plugin install to succeed, got error: %v", err)
	}

	_, err = service.ExecuteRuntimeUpgrade(ctx, pluginID, RuntimeUpgradeOptions{Confirmed: true})
	if !bizerr.Is(err, CodePluginRuntimeUpgradeUnavailable) {
		t.Fatalf("expected upgrade-unavailable bizerr, got %v", err)
	}
}

// TestInstallKeepsDynamicHigherVersionPendingUntilExplicitUpgrade verifies the
// legacy install path no longer switches releases when a higher artifact is staged.
func TestInstallKeepsDynamicHigherVersionPendingUntilExplicitUpgrade(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dynamic-runtime-upgrade-install-pending"
		oldVersion = "v0.1.0"
		newVersion = "v0.2.0"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Runtime Upgrade Install Pending Plugin",
		oldVersion,
		nil,
		nil,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic manifest sync to succeed, got error: %v", err)
	}
	if _, err = service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected dynamic plugin install to succeed, got error: %v", err)
	}
	oldRelease, err := service.getPluginRelease(ctx, pluginID, oldVersion)
	if err != nil {
		t.Fatalf("expected old release lookup to succeed, got error: %v", err)
	}
	if oldRelease == nil {
		t.Fatal("expected old release row")
	}

	testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Runtime Upgrade Install Pending Plugin",
		newVersion,
		nil,
		nil,
	)
	if _, err = service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected install path to keep staged upgrade pending, got error: %v", err)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup to succeed, got error: %v", err)
	}
	if registry == nil || registry.Version != oldVersion || registry.ReleaseId != oldRelease.Id {
		t.Fatalf("expected install path to keep effective release %s/%d, got %#v", oldVersion, oldRelease.Id, registry)
	}
	item := findPluginItemFromService(t, service, ctx, pluginID)
	if item.RuntimeState != RuntimeUpgradeStatePendingUpgrade || !item.UpgradeAvailable {
		t.Fatalf("expected pending runtime upgrade after install path, got %#v", item)
	}
}

// TestExecuteRuntimeUpgradeUpgradesDynamicPlugin verifies the confirmed runtime
// upgrade path switches the active release and records upgrade SQL.
func TestExecuteRuntimeUpgradeUpgradesDynamicPlugin(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dynamic-runtime-upgrade-execute"
		oldVersion = "v0.1.0"
		newVersion = "v0.2.0"
	)

	artifactPath := filepath.Join(testutil.TestDynamicStorageDir(), pluginID+".wasm")
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove runtime upgrade execute artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      pluginID,
			Name:    "Dynamic Runtime Upgrade Execute Plugin",
			Version: oldVersion,
			Type:    catalog.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: pluginbridge.RuntimeKindWasm,
			ABIVersion:  pluginbridge.SupportedABIVersion,
		},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected initial dynamic plugin install to succeed, got error: %v", err)
	}
	oldRelease, err := service.getPluginRelease(ctx, pluginID, oldVersion)
	if err != nil {
		t.Fatalf("expected old release lookup to succeed, got error: %v", err)
	}
	if oldRelease == nil {
		t.Fatal("expected old release row")
	}

	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      pluginID,
			Name:    "Dynamic Runtime Upgrade Execute Plugin",
			Version: newVersion,
			Type:    catalog.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:   pluginbridge.RuntimeKindWasm,
			ABIVersion:    pluginbridge.SupportedABIVersion,
			SQLAssetCount: 1,
		},
		nil,
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-plugin-dynamic-runtime-upgrade-execute.sql",
				Content: "CREATE TABLE IF NOT EXISTS plugin_dynamic_runtime_upgrade_execute(id INTEGER);",
			},
		},
		nil,
		nil,
		nil,
		nil,
	)
	newManifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected new dynamic artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, newManifest); err != nil {
		t.Fatalf("expected new dynamic manifest sync to succeed, got error: %v", err)
	}

	registryBeforeRun, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup before running state mark to succeed, got error: %v", err)
	}
	if err = service.markRuntimeUpgradeRunning(ctx, registryBeforeRun); err != nil {
		t.Fatalf("expected running state mark to succeed, got error: %v", err)
	}
	runningItem := findPluginItemFromService(t, service, ctx, pluginID)
	if runningItem.RuntimeState != RuntimeUpgradeStateUpgradeRunning {
		t.Fatalf("expected running state projection during upgrade, got %#v", runningItem)
	}
	if err = service.catalogSvc.SetRegistryRuntimeState(ctx, pluginID, do.SysPlugin{
		CurrentState: catalog.HostStateInstalled.String(),
	}); err != nil {
		t.Fatalf("expected running state reset to succeed, got error: %v", err)
	}

	result, err := service.ExecuteRuntimeUpgrade(ctx, pluginID, RuntimeUpgradeOptions{Confirmed: true})
	if err != nil {
		t.Fatalf("expected runtime upgrade execution to succeed, got error: %v", err)
	}
	if result == nil || !result.Executed {
		t.Fatalf("expected executed runtime upgrade result, got %#v", result)
	}
	if result.FromVersion != oldVersion || result.ToVersion != newVersion {
		t.Fatalf("expected result versions %s/%s, got %#v", oldVersion, newVersion, result)
	}
	if result.RuntimeState != RuntimeUpgradeStateNormal {
		t.Fatalf("expected post-upgrade runtime state normal, got %#v", result)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup after upgrade to succeed, got error: %v", err)
	}
	if registry == nil || registry.Version != newVersion {
		t.Fatalf("expected effective version %s after upgrade, got %#v", newVersion, registry)
	}
	if registry.ReleaseId == oldRelease.Id {
		t.Fatalf("expected active release to switch away from %d, got %#v", oldRelease.Id, registry)
	}
	item := findPluginItemFromService(t, service, ctx, pluginID)
	if item.RuntimeState != RuntimeUpgradeStateNormal || item.UpgradeAvailable {
		t.Fatalf("expected plugin item to be normal after upgrade, got %#v", item)
	}

	var migrationCount int
	migrationCount, err = dao.SysPluginMigration.Ctx(ctx).
		Where(do.SysPluginMigration{
			PluginId: pluginID,
			Phase:    catalog.MigrationDirectionUpgrade.String(),
			Status:   catalog.MigrationExecutionStatusSucceeded.String(),
		}).
		Count()
	if err != nil {
		t.Fatalf("expected upgrade migration count query to succeed, got error: %v", err)
	}
	if migrationCount != 1 {
		t.Fatalf("expected one successful upgrade migration, got %d", migrationCount)
	}
}

// TestExecuteRuntimeUpgradeFailureKeepsEffectiveDynamicVersion verifies failed
// upgrade execution preserves the current effective release and exposes a failed state.
func TestExecuteRuntimeUpgradeFailureKeepsEffectiveDynamicVersion(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dynamic-runtime-upgrade-execute-failed"
		oldVersion = "v0.1.0"
		newVersion = "v0.2.0"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Runtime Upgrade Execute Failed Plugin",
		oldVersion,
		nil,
		nil,
	)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected initial artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected initial manifest sync to succeed, got error: %v", err)
	}
	if _, err = service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected initial dynamic plugin install to succeed, got error: %v", err)
	}
	oldRegistry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected old registry lookup to succeed, got error: %v", err)
	}
	if oldRegistry == nil {
		t.Fatal("expected old registry row")
	}

	testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Runtime Upgrade Execute Failed Plugin",
		newVersion,
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-plugin-dynamic-runtime-upgrade-execute-failed.sql",
				Content: "THIS IS NOT VALID SQL;",
			},
		},
		nil,
	)
	newManifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected failed target artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, newManifest); err != nil {
		t.Fatalf("expected failed target manifest sync to succeed, got error: %v", err)
	}

	_, err = service.ExecuteRuntimeUpgrade(ctx, pluginID, RuntimeUpgradeOptions{Confirmed: true})
	if !bizerr.Is(err, CodePluginRuntimeUpgradeExecutionFailed) {
		t.Fatalf("expected runtime upgrade execution failure bizerr, got %v", err)
	}
	registryAfterFailure, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup after failed upgrade to succeed, got error: %v", err)
	}
	if registryAfterFailure == nil ||
		registryAfterFailure.Version != oldVersion ||
		registryAfterFailure.ReleaseId != oldRegistry.ReleaseId {
		t.Fatalf("expected effective release to stay %s/%d, got %#v", oldVersion, oldRegistry.ReleaseId, registryAfterFailure)
	}
	item := findPluginItemFromService(t, service, ctx, pluginID)
	if item.RuntimeState != RuntimeUpgradeStateUpgradeFailed || item.LastUpgradeFailure == nil {
		t.Fatalf("expected upgrade_failed projection after failed upgrade, got %#v", item)
	}
}

// TestExecuteRuntimeUpgradeBeforeLifecycleBlocksBeforeRunningState verifies
// dynamic BeforeUpgrade preconditions run before upgrade state markers or
// release-switch side effects.
func TestExecuteRuntimeUpgradeBeforeLifecycleBlocksBeforeRunningState(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dynamic-before-upgrade-fail-closed"
		oldVersion = "v0.1.0"
		newVersion = "v0.2.0"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Before Upgrade Fail Closed Plugin",
		oldVersion,
		nil,
		nil,
	)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected initial artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected initial manifest sync to succeed, got error: %v", err)
	}
	if _, err = service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected initial dynamic plugin install to succeed, got error: %v", err)
	}
	oldRegistry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected old registry lookup to succeed, got error: %v", err)
	}
	if oldRegistry == nil {
		t.Fatal("expected old registry row")
	}

	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      pluginID,
			Name:    "Dynamic Before Upgrade Fail Closed Plugin",
			Version: newVersion,
			Type:    catalog.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: pluginbridge.RuntimeKindWasm,
			ABIVersion:  pluginbridge.SupportedABIVersion,
			LifecycleContracts: []*pluginbridge.LifecycleContract{
				{
					Operation:    pluginbridge.LifecycleOperationBeforeUpgrade,
					RequestType:  "DynamicBeforeUpgradeReq",
					InternalPath: "/__lifecycle/before-upgrade",
					TimeoutMs:    1000,
				},
			},
		},
		nil,
		nil,
		nil,
		nil,
		nil,
		&pluginbridge.BridgeSpec{
			ABIVersion:     pluginbridge.ABIVersionV1,
			RuntimeKind:    pluginbridge.RuntimeKindWasm,
			RouteExecution: true,
			RequestCodec:   pluginbridge.CodecProtobuf,
			ResponseCodec:  pluginbridge.CodecProtobuf,
		},
	)
	newManifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected target artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, newManifest); err != nil {
		t.Fatalf("expected target manifest sync to succeed, got error: %v", err)
	}

	_, err = service.ExecuteRuntimeUpgrade(ctx, pluginID, RuntimeUpgradeOptions{Confirmed: true})
	if !bizerr.Is(err, CodePluginLifecyclePreconditionVetoed) {
		t.Fatalf("expected dynamic BeforeUpgrade precondition bizerr, got %v", err)
	}
	registryAfterFailure, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup after vetoed upgrade to succeed, got error: %v", err)
	}
	if registryAfterFailure == nil ||
		registryAfterFailure.Version != oldVersion ||
		registryAfterFailure.ReleaseId != oldRegistry.ReleaseId ||
		registryAfterFailure.CurrentState == catalog.RuntimeUpgradeStateUpgradeRunning.String() {
		t.Fatalf("expected vetoed upgrade to preserve effective release and avoid running state, got %#v", registryAfterFailure)
	}
	item := findPluginItemFromService(t, service, ctx, pluginID)
	if item.RuntimeState != RuntimeUpgradeStatePendingUpgrade {
		t.Fatalf("expected lifecycle veto to leave plugin pending upgrade, got %#v", item)
	}
}

// TestExecuteRuntimeUpgradeLifecycleCallbackBlocksBeforeUpgradeSQL verifies
// dynamic Upgrade execution callbacks run before target upgrade SQL.
func TestExecuteRuntimeUpgradeLifecycleCallbackBlocksBeforeUpgradeSQL(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dynamic-upgrade-lifecycle-fail-closed"
		oldVersion = "v0.1.0"
		newVersion = "v0.2.0"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Upgrade Lifecycle Fail Closed Plugin",
		oldVersion,
		nil,
		nil,
	)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected initial artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected initial manifest sync to succeed, got error: %v", err)
	}
	if _, err = service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected initial dynamic plugin install to succeed, got error: %v", err)
	}
	oldRegistry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected old registry lookup to succeed, got error: %v", err)
	}
	if oldRegistry == nil {
		t.Fatal("expected old registry row")
	}

	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      pluginID,
			Name:    "Dynamic Upgrade Lifecycle Fail Closed Plugin",
			Version: newVersion,
			Type:    catalog.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:   pluginbridge.RuntimeKindWasm,
			ABIVersion:    pluginbridge.SupportedABIVersion,
			SQLAssetCount: 1,
			LifecycleContracts: []*pluginbridge.LifecycleContract{
				{
					Operation:    pluginbridge.LifecycleOperationUpgrade,
					RequestType:  "DynamicUpgradeReq",
					InternalPath: "/__lifecycle/upgrade",
					TimeoutMs:    1000,
				},
			},
		},
		nil,
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-plugin-dynamic-upgrade-lifecycle-fail-closed.sql",
				Content: "CREATE TABLE IF NOT EXISTS plugin_dynamic_upgrade_lifecycle_fail_closed(id INTEGER);",
			},
		},
		nil,
		nil,
		nil,
		&pluginbridge.BridgeSpec{
			ABIVersion:     pluginbridge.ABIVersionV1,
			RuntimeKind:    pluginbridge.RuntimeKindWasm,
			RouteExecution: true,
			RequestCodec:   pluginbridge.CodecProtobuf,
			ResponseCodec:  pluginbridge.CodecProtobuf,
		},
	)
	newManifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected target artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, newManifest); err != nil {
		t.Fatalf("expected target manifest sync to succeed, got error: %v", err)
	}

	_, err = service.ExecuteRuntimeUpgrade(ctx, pluginID, RuntimeUpgradeOptions{Confirmed: true})
	if !bizerr.Is(err, CodePluginRuntimeUpgradeExecutionFailed) {
		t.Fatalf("expected dynamic Upgrade lifecycle execution failure, got %v", err)
	}
	registryAfterFailure, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup after failed upgrade to succeed, got error: %v", err)
	}
	if registryAfterFailure == nil ||
		registryAfterFailure.Version != oldVersion ||
		registryAfterFailure.ReleaseId != oldRegistry.ReleaseId {
		t.Fatalf("expected effective release to stay %s/%d, got %#v", oldVersion, oldRegistry.ReleaseId, registryAfterFailure)
	}

	migrationCount, err := dao.SysPluginMigration.Ctx(ctx).
		Where(do.SysPluginMigration{
			PluginId: pluginID,
			Phase:    catalog.MigrationDirectionUpgrade.String(),
			Status:   catalog.MigrationExecutionStatusSucceeded.String(),
		}).
		Count()
	if err != nil {
		t.Fatalf("expected upgrade migration count query to succeed, got error: %v", err)
	}
	if migrationCount != 0 {
		t.Fatalf("expected Upgrade lifecycle failure to block upgrade SQL, got successful migration count=%d", migrationCount)
	}
}

// TestRuntimeUpgradeLockSerializesSamePlugin verifies the explicit upgrade
// entrypoint serializes side effects for the same plugin inside the current process.
func TestRuntimeUpgradeLockSerializesSamePlugin(t *testing.T) {
	service := newTestService()
	var (
		started       int32
		inside        int32
		maxConcurrent int32
	)
	firstEntered := make(chan struct{})
	releaseFirst := make(chan struct{})

	runLocked := func() {
		unlock, err := service.lockRuntimeUpgrade(context.Background(), "plugin-runtime-upgrade-lock")
		if err != nil {
			t.Errorf("expected local lock to succeed, got %v", err)
			return
		}
		defer unlock()
		current := atomic.AddInt32(&inside, 1)
		if current > atomic.LoadInt32(&maxConcurrent) {
			atomic.StoreInt32(&maxConcurrent, current)
		}
		if atomic.AddInt32(&started, 1) == 1 {
			close(firstEntered)
			<-releaseFirst
		}
		atomic.AddInt32(&inside, -1)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		runLocked()
	}()
	<-firstEntered
	go func() {
		defer wg.Done()
		runLocked()
	}()

	time.Sleep(20 * time.Millisecond)
	if atomic.LoadInt32(&started) != 1 {
		t.Fatalf("expected second same-plugin upgrade to wait, started=%d", atomic.LoadInt32(&started))
	}
	close(releaseFirst)
	wg.Wait()
	if maxConcurrent != 1 {
		t.Fatalf("expected same-plugin runtime upgrade lock to serialize calls, maxConcurrent=%d", maxConcurrent)
	}
}

// TestRuntimeUpgradeClusterLockRejectsMissingBackend verifies cluster mode
// fails closed instead of falling back to a process-local upgrade lock.
func TestRuntimeUpgradeClusterLockRejectsMissingBackend(t *testing.T) {
	service := newTestServiceWithTopology(&testTopology{
		enabled: true,
		primary: true,
		nodeID:  "cluster-node-missing-lock",
	})
	service.runtimeUpgradeLockStore = nil

	unlock, err := service.lockRuntimeUpgrade(context.Background(), "plugin-cluster-missing-lock")
	if unlock != nil {
		unlock()
	}
	if !bizerr.Is(err, CodePluginRuntimeUpgradeLockUnavailable) {
		t.Fatalf("expected lock-unavailable bizerr, got %v", err)
	}
}

// TestRuntimeUpgradeClusterLockSerializesAcrossServices verifies two service
// instances sharing coordination cannot upgrade the same plugin concurrently.
func TestRuntimeUpgradeClusterLockSerializesAcrossServices(t *testing.T) {
	ctx := context.Background()
	coordSvc := coordination.NewMemory(nil)
	first := newTestServiceWithTopology(&testTopology{
		enabled: true,
		primary: true,
		nodeID:  "cluster-node-a",
	})
	second := newTestServiceWithTopology(&testTopology{
		enabled: true,
		primary: true,
		nodeID:  "cluster-node-b",
	})
	first.runtimeUpgradeLockStore = coordSvc.Lock()
	second.runtimeUpgradeLockStore = coordSvc.Lock()

	unlockFirst, err := first.lockRuntimeUpgrade(ctx, "plugin-cluster-shared-lock")
	if err != nil {
		t.Fatalf("expected first cluster lock acquisition to succeed, got %v", err)
	}
	unlockSecond, err := second.lockRuntimeUpgrade(ctx, "plugin-cluster-shared-lock")
	if unlockSecond != nil {
		unlockSecond()
	}
	if !bizerr.Is(err, CodePluginRuntimeUpgradeAlreadyRunning) {
		t.Fatalf("expected already-running bizerr for second lock, got %v", err)
	}

	unlockFirst()
	unlockSecond, err = second.lockRuntimeUpgrade(ctx, "plugin-cluster-shared-lock")
	if err != nil {
		t.Fatalf("expected second cluster lock acquisition after release to succeed, got %v", err)
	}
	unlockSecond()
}

// findPluginItemFromService reads the plugin list and returns the target item.
func findPluginItemFromService(
	t *testing.T,
	service *serviceImpl,
	ctx context.Context,
	pluginID string,
) *PluginItem {
	t.Helper()

	out, err := service.List(ctx, ListInput{})
	if err != nil {
		t.Fatalf("expected plugin list to succeed, got error: %v", err)
	}
	item := findPluginItem(out, pluginID)
	if item == nil {
		t.Fatalf("expected plugin item %s", pluginID)
	}
	return item
}
