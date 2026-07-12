// This file covers plugin runtime-upgrade preview and execution paths.

package plugin

import (
	"context"
	pluginv1 "lina-core/api/plugin/v1"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/coordination"
	"lina-core/internal/service/plugin/internal/catalog"
	plugindep "lina-core/internal/service/plugin/internal/dependency"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/internal/service/plugin/internal/upgrade"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/pluginbridge/protocol"
	"lina-core/pkg/plugin/pluginhost"
	"lina-core/pkg/statusflag"
)

// containsString reports whether values contains target.
func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

// TestPreviewRuntimeUpgradeReturnsPendingDynamicPlan verifies that preview is
// read-only and exposes manifest snapshots, dependency checks, SQL summary,
// hostServices drift, and stable risk hints for a pending dynamic upgrade.
func TestPreviewRuntimeUpgradeReturnsPendingDynamicPlan(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dev-dynamic-runtime-upgrade-preview"
		oldVersion = "v0.1.0"
		newVersion = "v0.2.0"
	)

	artifactPath := filepath.Join(testutil.TestDynamicStorageDir(), pluginID+".wasm")
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove runtime upgrade preview artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      pluginID,
			Name:    "Dynamic Runtime Upgrade Preview Plugin",
			Version: oldVersion,
			Type:    pluginv1.PluginTypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: protocol.RuntimeKindWasm,
			ABIVersion:  protocol.SupportedABIVersion,
			HostServices: []*protocol.HostServiceSpec{
				{
					Service: protocol.HostServiceStorage,
					Methods: []string{protocol.HostServiceMethodStorageGet},
					Paths:   []string{"reports/"},
				},
			},
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
			Name:    "Dynamic Runtime Upgrade Preview Plugin",
			Version: newVersion,
			Type:    pluginv1.PluginTypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: protocol.RuntimeKindWasm,
			ABIVersion:  protocol.SupportedABIVersion,
			HostServices: []*protocol.HostServiceSpec{
				{
					Service: protocol.HostServiceStorage,
					Methods: []string{
						protocol.HostServiceMethodStorageGet,
						protocol.HostServiceMethodStoragePut,
					},
					Paths: []string{"reports/", "exports/"},
				},
			},
		},
		nil,
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-upgrade-preview.sql",
				Content: "CREATE TABLE IF NOT EXISTS plugin_dynamic_runtime_upgrade_preview(id INTEGER);",
			},
		},
		nil,
		nil,
		nil,
		nil,
	)
	newManifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected target dynamic artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, newManifest); err != nil {
		t.Fatalf("expected target manifest sync to succeed, got error: %v", err)
	}

	preview, err := service.PreviewRuntimeUpgrade(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected runtime upgrade preview to succeed, got error: %v", err)
	}
	if preview.PluginID != pluginID || preview.RuntimeState != RuntimeUpgradeStatePendingUpgrade {
		t.Fatalf("expected pending preview for %s, got %#v", pluginID, preview)
	}
	if preview.EffectiveVersion != oldVersion || preview.DiscoveredVersion != newVersion {
		t.Fatalf("expected versions %s/%s, got %#v", oldVersion, newVersion, preview)
	}
	if preview.FromManifest == nil || preview.FromManifest.Version != oldVersion {
		t.Fatalf("expected from manifest version %s, got %#v", oldVersion, preview.FromManifest)
	}
	if preview.ToManifest == nil || preview.ToManifest.Version != newVersion {
		t.Fatalf("expected to manifest version %s, got %#v", newVersion, preview.ToManifest)
	}
	if preview.SQLSummary.InstallSQLCount != 1 || preview.SQLSummary.RuntimeSQLAssetCount != 1 {
		t.Fatalf("expected target SQL summary to include one SQL asset, got %#v", preview.SQLSummary)
	}
	if !preview.HostServicesDiff.AuthorizationRequired || !preview.HostServicesDiff.AuthorizationChanged {
		t.Fatalf("expected host service authorization to be required and changed, got %#v", preview.HostServicesDiff)
	}
	if len(preview.HostServicesDiff.Changed) != 1 {
		t.Fatalf("expected one changed host service, got %#v", preview.HostServicesDiff)
	}
	change := preview.HostServicesDiff.Changed[0]
	if change.Service != protocol.HostServiceStorage {
		t.Fatalf("expected storage host service change, got %#v", change)
	}
	if len(change.FromPaths) != 1 || change.FromPaths[0] != "reports/" {
		t.Fatalf("expected from paths to contain reports/, got %#v", change.FromPaths)
	}
	if len(change.ToPaths) != 2 || change.ToPaths[0] != "exports/" || change.ToPaths[1] != "reports/" {
		t.Fatalf("expected target paths to contain exports/ and reports/, got %#v", change.ToPaths)
	}
	if preview.DependencyCheck == nil || preview.DependencyCheck.TargetID != pluginID {
		t.Fatalf("expected dependency check for target plugin, got %#v", preview.DependencyCheck)
	}
	if !containsString(preview.RiskHints, RuntimeUpgradeRiskHintUpgradeSQLRequiresReview) {
		t.Fatalf("expected SQL review risk hint, got %#v", preview.RiskHints)
	}
	if !containsString(preview.RiskHints, RuntimeUpgradeRiskHintHostServiceAuthorizationChanged) {
		t.Fatalf("expected host service authorization risk hint, got %#v", preview.RiskHints)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup after preview to succeed, got error: %v", err)
	}
	if registry == nil || registry.Version != oldVersion || registry.ReleaseId != oldRelease.Id {
		t.Fatalf("expected preview not to switch effective release, got %#v", registry)
	}
}

// TestPreviewRuntimeUpgradeRejectsNormalPlugin verifies preview does not turn a
// non-pending plugin into an upgrade action.
func TestPreviewRuntimeUpgradeRejectsNormalPlugin(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dev-dynamic-runtime-upgrade-preview-normal"
		version  = "v0.1.0"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Runtime Upgrade Preview Normal Plugin",
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

	_, err = service.PreviewRuntimeUpgrade(ctx, pluginID)
	if !bizerr.Is(err, CodePluginRuntimeUpgradePreviewUnavailable) {
		t.Fatalf("expected preview unavailable bizerr, got %v", err)
	}
}

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
		pluginID = "plugin-dev-dynamic-runtime-upgrade-normal"
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
// install path keeps a staged higher version pending until an explicit upgrade.
func TestInstallKeepsDynamicHigherVersionPendingUntilExplicitUpgrade(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dev-dynamic-runtime-upgrade-install-pending"
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
		pluginID   = "plugin-dev-dynamic-runtime-upgrade-execute"
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
			Type:    pluginv1.PluginTypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: protocol.RuntimeKindWasm,
			ABIVersion:  protocol.SupportedABIVersion,
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
			Type:    pluginv1.PluginTypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:   protocol.RuntimeKindWasm,
			ABIVersion:    protocol.SupportedABIVersion,
			SQLAssetCount: 1,
		},
		nil,
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-plugin-dev-dynamic-runtime-upgrade-execute.sql",
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
	if err = service.storeSvc.SetRegistryRuntimeState(ctx, registryBeforeRun.PluginId, store.RuntimeStatePatch{
		CurrentState: plugintypes.RuntimeUpgradeStateUpgradeRunning.String(),
	}); err != nil {
		t.Fatalf("expected running state mark to succeed, got error: %v", err)
	}
	runningItem := findPluginItemFromService(t, service, ctx, pluginID)
	if runningItem.RuntimeState != RuntimeUpgradeStateUpgradeRunning {
		t.Fatalf("expected running state projection during upgrade, got %#v", runningItem)
	}
	if err = service.storeSvc.SetRegistryRuntimeState(ctx, pluginID, store.RuntimeStatePatch{
		CurrentState: plugintypes.HostStateInstalled.String(),
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
			Phase:    plugintypes.MigrationDirectionUpgrade.String(),
			Status:   plugintypes.MigrationExecutionStatusSucceeded.String(),
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
		pluginID   = "plugin-dev-dynamic-runtime-upgrade-execute-failed"
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
				Key:     "001-plugin-dev-dynamic-runtime-upgrade-execute-failed.sql",
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
	if item.LastUpgradeFailure.Phase != plugintypes.RuntimeUpgradeFailurePhaseSQL {
		t.Fatalf("expected dynamic SQL failure phase, got %#v", item.LastUpgradeFailure)
	}
	if item.LastUpgradeFailure.MessageKey != store.RuntimeUpgradeFailureMessageKeyMigrationFailed {
		t.Fatalf("expected migration failure message key, got %#v", item.LastUpgradeFailure)
	}
	failedRelease, err := service.getPluginRelease(ctx, pluginID, newVersion)
	if err != nil {
		t.Fatalf("expected failed dynamic release lookup to succeed, got error: %v", err)
	}
	if failedRelease == nil {
		t.Fatal("expected failed dynamic target release")
	}
	var failedMigrationCount int
	failedMigrationCount, err = dao.SysPluginMigration.Ctx(ctx).
		Where(do.SysPluginMigration{
			PluginId:     pluginID,
			ReleaseId:    failedRelease.Id,
			Phase:        plugintypes.MigrationDirectionUpgrade.String(),
			MigrationKey: "upgrade-phase-sql",
			Status:       plugintypes.MigrationExecutionStatusFailed.String(),
		}).
		Count()
	if err != nil {
		t.Fatalf("expected dynamic upgrade failure migration query to succeed, got error: %v", err)
	}
	if failedMigrationCount != 1 {
		t.Fatalf("expected one dynamic SQL failure migration, got %d", failedMigrationCount)
	}
}

// TestExecuteRuntimeUpgradeDynamicCachePublisherFailureReturnsError verifies
// the unified runtime-upgrade owner surfaces final cache publication failures
// after the dynamic release switch has completed.
func TestExecuteRuntimeUpgradeDynamicCachePublisherFailureReturnsError(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dev-dynamic-runtime-upgrade-cache-failed"
		oldVersion = "v0.1.0"
		newVersion = "v0.2.0"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Runtime Upgrade Cache Failed Plugin",
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

	testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Runtime Upgrade Cache Failed Plugin",
		newVersion,
		nil,
		nil,
	)
	newManifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected target artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, newManifest); err != nil {
		t.Fatalf("expected target manifest sync to succeed, got error: %v", err)
	}
	service.replaceUpgradeServiceForTest(t, service.integrationSvc, failingCachePublisher{
		syncErr: gerror.New("runtime upgrade cache publication failed"),
	}, service.runtimeUpgradeLockStore)

	_, err = service.ExecuteRuntimeUpgrade(ctx, pluginID, RuntimeUpgradeOptions{Confirmed: true})
	if err == nil {
		t.Fatal("expected dynamic runtime upgrade cache publication failure")
	}
	registry, lookupErr := service.getPluginRegistry(ctx, pluginID)
	if lookupErr != nil {
		t.Fatalf("expected registry lookup after cache publication failure to succeed, got error: %v", lookupErr)
	}
	if registry == nil || registry.Version != newVersion {
		t.Fatalf("expected dynamic upgrade side effects to complete before cache error, got %#v", registry)
	}
	item := findPluginItemFromService(t, service, ctx, pluginID)
	if item.RuntimeState != RuntimeUpgradeStateNormal || item.UpgradeAvailable {
		t.Fatalf("expected dynamic plugin to be effective despite cache error, got %#v", item)
	}
	newRelease, err := service.getPluginRelease(ctx, pluginID, newVersion)
	if err != nil {
		t.Fatalf("expected dynamic target release lookup after cache failure to succeed, got error: %v", err)
	}
	if newRelease == nil {
		t.Fatal("expected dynamic target release after cache failure")
	}
	var failedMigrationCount int
	failedMigrationCount, err = dao.SysPluginMigration.Ctx(ctx).
		Where(do.SysPluginMigration{
			PluginId:  pluginID,
			ReleaseId: newRelease.Id,
			Phase:     plugintypes.MigrationDirectionUpgrade.String(),
			Status:    plugintypes.MigrationExecutionStatusFailed.String(),
		}).
		Count()
	if err != nil {
		t.Fatalf("expected dynamic cache failure migration query to succeed, got error: %v", err)
	}
	if failedMigrationCount != 0 {
		t.Fatalf("expected final cache publisher failure not to create failed release ledger, got %d", failedMigrationCount)
	}
}

// TestExecuteRuntimeUpgradeDynamicSameVersionRefreshUnavailable verifies the
// explicit runtime-upgrade entry does not turn same-version dynamic refreshes
// into upgrade executions.
func TestExecuteRuntimeUpgradeDynamicSameVersionRefreshUnavailable(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dev-dynamic-runtime-upgrade-same-version"
		version  = "v0.1.0"
	)

	testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Runtime Upgrade Same Version Plugin",
		version,
		nil,
		nil,
	)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected dynamic plugin install to succeed, got error: %v", err)
	}
	registryBefore, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup before same-version upgrade request to succeed, got error: %v", err)
	}
	if registryBefore == nil {
		t.Fatal("expected registry row before same-version upgrade request")
	}

	_, err = service.ExecuteRuntimeUpgrade(ctx, pluginID, RuntimeUpgradeOptions{Confirmed: true})
	if !bizerr.Is(err, CodePluginRuntimeUpgradeUnavailable) {
		t.Fatalf("expected same-version runtime upgrade to be unavailable, got %v", err)
	}
	registryAfter, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup after same-version upgrade request to succeed, got error: %v", err)
	}
	if registryAfter == nil ||
		registryAfter.Version != registryBefore.Version ||
		registryAfter.ReleaseId != registryBefore.ReleaseId ||
		registryAfter.Generation != registryBefore.Generation {
		t.Fatalf("expected same-version upgrade request not to refresh release, before=%#v after=%#v", registryBefore, registryAfter)
	}
}

// TestExecuteRuntimeUpgradeBeforeLifecycleBlocksBeforeRunningState verifies
// dynamic BeforeUpgrade preconditions run before upgrade state markers or
// release-switch side effects.
func TestExecuteRuntimeUpgradeBeforeLifecycleBlocksBeforeRunningState(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dev-dynamic-before-upgrade-fail-closed"
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
			Type:    pluginv1.PluginTypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: protocol.RuntimeKindWasm,
			ABIVersion:  protocol.SupportedABIVersion,
			LifecycleContracts: []*protocol.LifecycleContract{
				{
					Operation:    protocol.LifecycleOperationBeforeUpgrade,
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
		&protocol.BridgeSpec{
			ABIVersion:     protocol.ABIVersionV1,
			RuntimeKind:    protocol.RuntimeKindWasm,
			RouteExecution: true,
			RequestCodec:   protocol.CodecProtobuf,
			ResponseCodec:  protocol.CodecProtobuf,
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
		registryAfterFailure.CurrentState == plugintypes.RuntimeUpgradeStateUpgradeRunning.String() {
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
		pluginID   = "plugin-dev-dynamic-upgrade-lifecycle-fail-closed"
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
			Type:    pluginv1.PluginTypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:   protocol.RuntimeKindWasm,
			ABIVersion:    protocol.SupportedABIVersion,
			SQLAssetCount: 1,
			LifecycleContracts: []*protocol.LifecycleContract{
				{
					Operation:    protocol.LifecycleOperationUpgrade,
					RequestType:  "DynamicUpgradeReq",
					InternalPath: "/__lifecycle/upgrade",
					TimeoutMs:    1000,
				},
			},
		},
		nil,
		[]*catalog.ArtifactSQLAsset{
			{
				Key:     "001-plugin-dev-dynamic-upgrade-lifecycle-fail-closed.sql",
				Content: "CREATE TABLE IF NOT EXISTS plugin_dynamic_upgrade_lifecycle_fail_closed(id INTEGER);",
			},
		},
		nil,
		nil,
		nil,
		&protocol.BridgeSpec{
			ABIVersion:     protocol.ABIVersionV1,
			RuntimeKind:    protocol.RuntimeKindWasm,
			RouteExecution: true,
			RequestCodec:   protocol.CodecProtobuf,
			ResponseCodec:  protocol.CodecProtobuf,
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
			Phase:    plugintypes.MigrationDirectionUpgrade.String(),
			Status:   plugintypes.MigrationExecutionStatusSucceeded.String(),
		}).
		Count()
	if err != nil {
		t.Fatalf("expected upgrade migration count query to succeed, got error: %v", err)
	}
	if migrationCount != 0 {
		t.Fatalf("expected Upgrade lifecycle failure to block upgrade SQL, got successful migration count=%d", migrationCount)
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
	service.replaceUpgradeServiceForTest(t, service.integrationSvc, upgradeCachePublisher{service: service}, nil)

	_, err := service.ExecuteRuntimeUpgrade(
		context.Background(),
		"plugin-cluster-missing-lock",
		RuntimeUpgradeOptions{Confirmed: true},
	)
	if !bizerr.Is(err, CodePluginRuntimeUpgradeLockUnavailable) {
		t.Fatalf("expected lock-unavailable bizerr, got %v", err)
	}
}

// TestRuntimeUpgradeClusterLockSerializesAcrossServices verifies two service
// instances sharing coordination cannot upgrade the same plugin concurrently.
func TestRuntimeUpgradeClusterLockSerializesAcrossServices(t *testing.T) {
	ctx := context.Background()
	coordSvc := coordination.NewMemory(nil)
	second := newTestServiceWithTopology(&testTopology{
		enabled: true,
		primary: true,
		nodeID:  "cluster-node-b",
	})
	sharedLockStore := coordSvc.Lock()
	second.replaceUpgradeServiceForTest(t, second.integrationSvc, upgradeCachePublisher{service: second}, sharedLockStore)

	handle, ok, err := sharedLockStore.Acquire(
		ctx,
		"plugin-runtime-upgrade:plugin-cluster-shared-lock",
		"external-owner",
		"test",
		time.Minute,
	)
	if err != nil || !ok || handle == nil {
		t.Fatalf("expected external cluster lock acquisition to succeed, ok=%v err=%v", ok, err)
	}
	defer func() {
		if releaseErr := sharedLockStore.Release(ctx, handle); releaseErr != nil {
			t.Fatalf("expected external cluster lock release to succeed, got %v", releaseErr)
		}
	}()

	_, err = second.ExecuteRuntimeUpgrade(
		ctx,
		"plugin-cluster-shared-lock",
		RuntimeUpgradeOptions{Confirmed: true},
	)
	if !bizerr.Is(err, CodePluginRuntimeUpgradeAlreadyRunning) {
		t.Fatalf("expected already-running bizerr for second lock, got %v", err)
	}
}

// findPluginItemFromService reads the plugin list and returns the target item.
// ID filter is required because the management list default page size is smaller
// than the official plugin workspace once multi-cloud storage plugins are present.
func findPluginItemFromService(
	t *testing.T,
	service *serviceImpl,
	ctx context.Context,
	pluginID string,
) *PluginItem {
	t.Helper()

	out, err := service.List(ctx, ListInput{ID: pluginID})
	if err != nil {
		t.Fatalf("expected plugin list to succeed, got error: %v", err)
	}
	item := findPluginItem(out, pluginID)
	if item == nil {
		t.Fatalf("expected plugin item %s", pluginID)
	}
	return item
}

// TestSourcePluginDiscoveryKeepsEffectiveVersionAfterHigherSourceVersion verifies
// discovered source versions do not overwrite the current effective registry version.
func TestSourcePluginDiscoveryKeepsEffectiveVersionAfterHigherSourceVersion(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dev-source-upgrade-drift"
		oldVersion = "v0.1.0"
		newVersion = "v0.5.0"
		oldMenuKey = "plugin:plugin-dev-source-upgrade-drift:old-entry"
		newMenuKey = "plugin:plugin-dev-source-upgrade-drift:new-entry"
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Drift Plugin", oldVersion, oldMenuKey)

	testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := service.SyncAndList(ctx); err != nil {
		t.Fatalf("expected source plugin discovery to succeed, got error: %v", err)
	}
	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected source plugin install to succeed, got error: %v", err)
	}

	registryBefore, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected source plugin registry lookup before drift to succeed, got error: %v", err)
	}
	if registryBefore == nil {
		t.Fatal("expected source plugin registry row before drift")
	}
	oldRelease, err := service.getPluginRelease(ctx, pluginID, oldVersion)
	if err != nil {
		t.Fatalf("expected old source plugin release lookup to succeed, got error: %v", err)
	}
	if oldRelease == nil {
		t.Fatal("expected old source plugin release row before drift")
	}

	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Drift Plugin", newVersion, newMenuKey)
	if _, err := service.SyncSourcePluginsStrict(ctx); err != nil {
		t.Fatalf("expected source plugin rescan to succeed, got error: %v", err)
	}

	registryAfter, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected source plugin registry lookup after drift to succeed, got error: %v", err)
	}
	if registryAfter == nil {
		t.Fatal("expected source plugin registry row after drift")
	}
	if registryAfter.Version != oldVersion {
		t.Fatalf("expected effective version %s to stay pinned, got %s", oldVersion, registryAfter.Version)
	}
	if registryAfter.ReleaseId != oldRelease.Id {
		t.Fatalf("expected release_id to stay pinned to old release %d, got %d", oldRelease.Id, registryAfter.ReleaseId)
	}

	preparedRelease, err := service.getPluginRelease(ctx, pluginID, newVersion)
	if err != nil {
		t.Fatalf("expected prepared source plugin release lookup to succeed, got error: %v", err)
	}
	if preparedRelease == nil {
		t.Fatal("expected prepared source plugin release row after drift")
	}
	if preparedRelease.Status != plugintypes.ReleaseStatusPrepared.String() {
		t.Fatalf("expected prepared release status, got %s", preparedRelease.Status)
	}

	oldMenu, err := testutil.QueryMenuByKey(ctx, oldMenuKey)
	if err != nil {
		t.Fatalf("expected old source plugin menu query to succeed, got error: %v", err)
	}
	if oldMenu == nil {
		t.Fatal("expected old source plugin menu to remain effective before explicit upgrade")
	}
	newMenu, err := testutil.QueryMenuByKey(ctx, newMenuKey)
	if err != nil {
		t.Fatalf("expected new source plugin menu query to succeed, got error: %v", err)
	}
	if newMenu != nil {
		t.Fatal("expected new source plugin menu to stay absent before explicit upgrade")
	}

	statuses, err := service.ListSourceUpgradeStatuses(ctx)
	if err != nil {
		t.Fatalf("expected source upgrade status listing to succeed, got error: %v", err)
	}
	status := findSourceUpgradeStatusByID(statuses, pluginID)
	if status == nil {
		t.Fatal("expected source plugin upgrade status after drift")
	}
	if !status.NeedsUpgrade {
		t.Fatalf("expected source plugin drift to report needsUpgrade, got %#v", status)
	}
	if status.EffectiveVersion != oldVersion || status.DiscoveredVersion != newVersion {
		t.Fatalf("expected effective/discovered versions %s/%s, got %#v", oldVersion, newVersion, status)
	}
}

// TestValidateSourcePluginUpgradeReadinessAllowsPendingUpgrade verifies startup
// drift scanning does not block boot when an installed source plugin has a newer
// discovered version waiting for explicit runtime upgrade.
func TestValidateSourcePluginUpgradeReadinessAllowsPendingUpgrade(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dev-source-upgrade-startup-guard"
		oldVersion = "v0.1.0"
		newVersion = "v0.5.0"
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Startup Guard Plugin", oldVersion, "plugin:plugin-dev-source-upgrade-startup-guard:old-entry")

	testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := service.SyncAndList(ctx); err != nil {
		t.Fatalf("expected source plugin discovery to succeed, got error: %v", err)
	}
	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected source plugin install to succeed, got error: %v", err)
	}

	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Startup Guard Plugin", newVersion, "plugin:plugin-dev-source-upgrade-startup-guard:new-entry")
	if _, err := service.SyncSourcePluginsStrict(ctx); err != nil {
		t.Fatalf("expected source plugin rescan to succeed, got error: %v", err)
	}

	err := service.ValidateSourcePluginUpgradeReadiness(ctx)
	if err != nil {
		t.Fatalf("expected source upgrade readiness scan not to fail for pending runtime upgrade, got error: %v", err)
	}

	out, err := service.List(ctx, ListInput{ID: pluginID})
	if err != nil {
		t.Fatalf("expected plugin list to succeed after pending source drift, got error: %v", err)
	}
	item := findPluginItem(out, pluginID)
	if item == nil {
		t.Fatal("expected source plugin list item after pending drift")
	}
	if item.RuntimeState != RuntimeUpgradeStatePendingUpgrade {
		t.Fatalf("expected runtime state %s, got %#v", RuntimeUpgradeStatePendingUpgrade, item)
	}
	if item.EffectiveVersion != oldVersion || item.DiscoveredVersion != newVersion {
		t.Fatalf("expected effective/discovered versions %s/%s, got %#v", oldVersion, newVersion, item)
	}
	if !item.UpgradeAvailable {
		t.Fatalf("expected pending source plugin to report upgradeAvailable, got %#v", item)
	}
}

// TestSourcePluginListMarksLowerDiscoveredVersionAbnormal verifies a file
// version lower than the effective registry version is exposed for manual repair.
func TestSourcePluginListMarksLowerDiscoveredVersionAbnormal(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dev-source-upgrade-abnormal"
		oldVersion = "v0.1.0"
		newVersion = "v0.5.0"
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Abnormal Plugin", newVersion, "plugin:plugin-dev-source-upgrade-abnormal:new-entry")

	testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := service.SyncAndList(ctx); err != nil {
		t.Fatalf("expected source plugin discovery to succeed, got error: %v", err)
	}
	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected source plugin install to succeed, got error: %v", err)
	}

	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Abnormal Plugin", oldVersion, "plugin:plugin-dev-source-upgrade-abnormal:old-entry")
	if _, err := service.SyncSourcePluginsStrict(ctx); err != nil {
		t.Fatalf("expected source plugin rescan to succeed, got error: %v", err)
	}

	out, err := service.List(ctx, ListInput{ID: pluginID})
	if err != nil {
		t.Fatalf("expected plugin list to succeed after lower source version drift, got error: %v", err)
	}
	item := findPluginItem(out, pluginID)
	if item == nil {
		t.Fatal("expected source plugin list item after lower source drift")
	}
	if item.RuntimeState != RuntimeUpgradeStateAbnormal {
		t.Fatalf("expected runtime state %s, got %#v", RuntimeUpgradeStateAbnormal, item)
	}
	if item.AbnormalReason != RuntimeUpgradeAbnormalReasonDiscoveredVersionLowerThanEffective {
		t.Fatalf("expected abnormal reason %s, got %#v", RuntimeUpgradeAbnormalReasonDiscoveredVersionLowerThanEffective, item)
	}
	if item.EffectiveVersion != newVersion || item.DiscoveredVersion != oldVersion {
		t.Fatalf("expected effective/discovered versions %s/%s, got %#v", newVersion, oldVersion, item)
	}
}

// TestUpgradeSourcePluginAppliesPreparedRelease verifies explicit source-plugin
// upgrade moves the effective registry version, records upgrade migrations, and
// switches host-owned menu governance to the new manifest.
func TestUpgradeSourcePluginAppliesPreparedRelease(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dev-source-upgrade-apply"
		oldVersion = "v0.1.0"
		newVersion = "v0.5.0"
		oldMenuKey = "plugin:plugin-dev-source-upgrade-apply:old-entry"
		newMenuKey = "plugin:plugin-dev-source-upgrade-apply:new-entry"
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Apply Plugin", oldVersion, oldMenuKey)

	testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := service.SyncAndList(ctx); err != nil {
		t.Fatalf("expected source plugin discovery to succeed, got error: %v", err)
	}
	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected source plugin install to succeed, got error: %v", err)
	}
	if err := service.UpdateStatus(ctx, pluginID, UpdateStatusOptions{Status: statusflag.EnabledValue.Int()}); err != nil {
		t.Fatalf("expected source plugin enable to succeed, got error: %v", err)
	}

	oldRelease, err := service.getPluginRelease(ctx, pluginID, oldVersion)
	if err != nil {
		t.Fatalf("expected old source plugin release lookup to succeed, got error: %v", err)
	}
	if oldRelease == nil {
		t.Fatal("expected old source plugin release row before upgrade")
	}

	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Apply Plugin", newVersion, newMenuKey)
	if _, err := service.SyncSourcePluginsStrict(ctx); err != nil {
		t.Fatalf("expected source plugin rescan to succeed, got error: %v", err)
	}

	result, err := service.UpgradeSourcePlugin(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected source plugin upgrade to succeed, got error: %v", err)
	}
	if result == nil || !result.Executed {
		t.Fatalf("expected source plugin upgrade to execute, got %#v", result)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected upgraded source plugin registry lookup to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatal("expected upgraded source plugin registry row")
	}
	if registry.Version != newVersion {
		t.Fatalf("expected upgraded effective version %s, got %s", newVersion, registry.Version)
	}
	if registry.Status != statusflag.EnabledValue.Int() {
		t.Fatalf("expected upgraded source plugin to stay enabled, got status=%d", registry.Status)
	}

	newRelease, err := service.getPluginRelease(ctx, pluginID, newVersion)
	if err != nil {
		t.Fatalf("expected new source plugin release lookup to succeed, got error: %v", err)
	}
	if newRelease == nil {
		t.Fatal("expected upgraded source plugin release row")
	}
	if registry.ReleaseId != newRelease.Id {
		t.Fatalf("expected registry release_id %d, got %d", newRelease.Id, registry.ReleaseId)
	}
	if newRelease.Status != plugintypes.ReleaseStatusActive.String() {
		t.Fatalf("expected new source plugin release to become active, got %s", newRelease.Status)
	}

	oldRelease, err = service.getPluginRelease(ctx, pluginID, oldVersion)
	if err != nil {
		t.Fatalf("expected old source plugin release lookup after upgrade to succeed, got error: %v", err)
	}
	if oldRelease == nil {
		t.Fatal("expected old source plugin release row after upgrade")
	}
	if oldRelease.Status != plugintypes.ReleaseStatusInstalled.String() {
		t.Fatalf("expected old source plugin release to be demoted to installed, got %s", oldRelease.Status)
	}

	upgradeMigrationCount, err := dao.SysPluginMigration.Ctx(ctx).
		Where(do.SysPluginMigration{
			PluginId:  pluginID,
			ReleaseId: newRelease.Id,
			Phase:     plugintypes.MigrationDirectionUpgrade.String(),
		}).
		Count()
	if err != nil {
		t.Fatalf("expected source plugin upgrade migration count query to succeed, got error: %v", err)
	}
	if upgradeMigrationCount != 1 {
		t.Fatalf("expected one source plugin upgrade migration row, got count=%d", upgradeMigrationCount)
	}

	oldMenu, err := testutil.QueryMenuByKey(ctx, oldMenuKey)
	if err != nil {
		t.Fatalf("expected old source plugin menu query after upgrade to succeed, got error: %v", err)
	}
	if oldMenu != nil {
		t.Fatal("expected old source plugin menu to be removed after explicit upgrade")
	}
	newMenu, err := testutil.QueryMenuByKey(ctx, newMenuKey)
	if err != nil {
		t.Fatalf("expected new source plugin menu query after upgrade to succeed, got error: %v", err)
	}
	if newMenu == nil {
		t.Fatal("expected new source plugin menu to be created after explicit upgrade")
	}

	resourceCount, err := dao.SysPluginResourceRef.Ctx(ctx).
		Where(do.SysPluginResourceRef{PluginId: pluginID, ReleaseId: newRelease.Id}).
		Count()
	if err != nil {
		t.Fatalf("expected upgraded source plugin resource ref count query to succeed, got error: %v", err)
	}
	if resourceCount == 0 {
		t.Fatal("expected upgraded source plugin release to retain governance resource refs")
	}
}

// TestUpgradeSourcePluginInvokesLifecycleCallbacks verifies explicit source
// upgrade passes before/upgrade/after callbacks the effective and target
// manifest snapshots.
func TestUpgradeSourcePluginInvokesLifecycleCallbacks(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dev-source-upgrade-callback"
		oldVersion = "v0.1.0"
		newVersion = "v0.5.0"
		events     []string
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Callback Plugin", oldVersion, "plugin:plugin-dev-source-upgrade-callback:old-entry")
	registerSourceUpgradeCallbacksForTest(t, pluginID, &events, false, false)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := service.SyncAndList(ctx); err != nil {
		t.Fatalf("expected source plugin discovery to succeed, got error: %v", err)
	}
	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected source plugin install to succeed, got error: %v", err)
	}

	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Callback Plugin", newVersion, "plugin:plugin-dev-source-upgrade-callback:new-entry")
	registerSourceUpgradeCallbacksForTest(t, pluginID, &events, false, false)
	if _, err := service.SyncSourcePluginsStrict(ctx); err != nil {
		t.Fatalf("expected source plugin rescan to succeed, got error: %v", err)
	}

	result, err := service.UpgradeSourcePlugin(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected source plugin upgrade to succeed, got error: %v", err)
	}
	if result == nil || !result.Executed {
		t.Fatalf("expected source upgrade to execute, got %#v", result)
	}
	expectedEvents := []string{"before:v0.1.0->v0.5.0", "upgrade:v0.1.0->v0.5.0", "after:v0.1.0->v0.5.0"}
	if !sourceUpgradeTestStringSlicesEqual(events, expectedEvents) {
		t.Fatalf("expected lifecycle events %#v, got %#v", expectedEvents, events)
	}
}

// TestUpgradeSourcePluginBeforeCallbackVetoes verifies unified lifecycle
// before-upgrade callbacks can block the upgrade before host side effects.
func TestUpgradeSourcePluginBeforeCallbackVetoes(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dev-source-upgrade-before-veto"
		oldVersion = "v0.1.0"
		newVersion = "v0.5.0"
		events     []string
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Before Veto Plugin", oldVersion, "plugin:plugin-dev-source-upgrade-before-veto:old-entry")
	registerSourceUpgradeCallbacksForTest(t, pluginID, &events, false, false)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := service.SyncAndList(ctx); err != nil {
		t.Fatalf("expected source plugin discovery to succeed, got error: %v", err)
	}
	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected source plugin install to succeed, got error: %v", err)
	}

	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Before Veto Plugin", newVersion, "plugin:plugin-dev-source-upgrade-before-veto:new-entry")
	registerSourceUpgradeCallbacksForTest(t, pluginID, &events, true, false)
	if _, err := service.SyncSourcePluginsStrict(ctx); err != nil {
		t.Fatalf("expected source plugin rescan to succeed, got error: %v", err)
	}

	_, err := service.UpgradeSourcePlugin(ctx, pluginID)
	if !bizerr.Is(err, sourceUpgradeCodeLifecycleVetoed()) {
		t.Fatalf("expected lifecycle veto bizerr, got %v", err)
	}
	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup after veto to succeed, got error: %v", err)
	}
	if registry == nil || registry.Version != oldVersion {
		t.Fatalf("expected effective version to stay %s after veto, got %#v", oldVersion, registry)
	}
	item := findPluginItemFromService(t, service, ctx, pluginID)
	if item.RuntimeState != RuntimeUpgradeStateUpgradeFailed || item.LastUpgradeFailure == nil {
		t.Fatalf("expected veto to be diagnosable as upgrade_failed, got %#v", item)
	}
	if item.LastUpgradeFailure.Phase != plugintypes.RuntimeUpgradeFailurePhaseBeforeUpgrade {
		t.Fatalf("expected before_upgrade failure phase, got %#v", item.LastUpgradeFailure)
	}
}

// TestUpgradeSourcePluginCallbackFailureIsRetryable verifies custom upgrade
// callback failures keep the effective release stable and allow a later retry.
func TestUpgradeSourcePluginCallbackFailureIsRetryable(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dev-source-upgrade-callback-retry"
		oldVersion = "v0.1.0"
		newVersion = "v0.5.0"
		events     []string
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Callback Retry Plugin", oldVersion, "plugin:plugin-dev-source-upgrade-callback-retry:old-entry")
	registerSourceUpgradeCallbacksForTest(t, pluginID, &events, false, false)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := service.SyncAndList(ctx); err != nil {
		t.Fatalf("expected source plugin discovery to succeed, got error: %v", err)
	}
	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected source plugin install to succeed, got error: %v", err)
	}

	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Callback Retry Plugin", newVersion, "plugin:plugin-dev-source-upgrade-callback-retry:new-entry")
	registerSourceUpgradeCallbacksForTest(t, pluginID, &events, false, true)
	if _, err := service.SyncSourcePluginsStrict(ctx); err != nil {
		t.Fatalf("expected source plugin rescan to succeed, got error: %v", err)
	}

	_, err := service.ExecuteRuntimeUpgrade(ctx, pluginID, RuntimeUpgradeOptions{Confirmed: true})
	if !bizerr.Is(err, CodePluginRuntimeUpgradeExecutionFailed) {
		t.Fatalf("expected runtime upgrade execution failure, got %v", err)
	}
	item := findPluginItemFromService(t, service, ctx, pluginID)
	if item.RuntimeState != RuntimeUpgradeStateUpgradeFailed || item.LastUpgradeFailure == nil {
		t.Fatalf("expected callback failure to be upgrade_failed, got %#v", item)
	}
	if item.LastUpgradeFailure.Phase != plugintypes.RuntimeUpgradeFailurePhaseUpgradeCallback {
		t.Fatalf("expected upgrade_callback failure phase, got %#v", item.LastUpgradeFailure)
	}

	registerSourceUpgradeCallbacksForTest(t, pluginID, &events, false, false)
	result, err := service.ExecuteRuntimeUpgrade(ctx, pluginID, RuntimeUpgradeOptions{Confirmed: true})
	if err != nil {
		t.Fatalf("expected retry after callback fix to succeed, got error: %v", err)
	}
	if result == nil || !result.Executed || result.RuntimeState != RuntimeUpgradeStateNormal {
		t.Fatalf("expected retry to execute and return normal state, got %#v", result)
	}
	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup after retry to succeed, got error: %v", err)
	}
	if registry == nil || registry.Version != newVersion {
		t.Fatalf("expected effective version %s after retry, got %#v", newVersion, registry)
	}
}

// TestExecuteRuntimeUpgradeSourceSQLFailureKeepsEffectiveRelease verifies the
// unified runtime-upgrade entry can dispatch directly into the source strategy,
// preserve the active release, and expose SQL failure diagnostics.
func TestExecuteRuntimeUpgradeSourceSQLFailureKeepsEffectiveRelease(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dev-source-upgrade-sql-failed"
		oldVersion = "v0.1.0"
		newVersion = "v0.5.0"
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade SQL Failed Plugin", oldVersion, "plugin:plugin-dev-source-upgrade-sql-failed:old-entry")

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := service.SyncAndList(ctx); err != nil {
		t.Fatalf("expected source plugin discovery to succeed, got error: %v", err)
	}
	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected source plugin install to succeed, got error: %v", err)
	}
	oldRegistry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected source plugin registry lookup to succeed, got error: %v", err)
	}
	if oldRegistry == nil {
		t.Fatal("expected source plugin registry before failed SQL upgrade")
	}

	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade SQL Failed Plugin", newVersion, "plugin:plugin-dev-source-upgrade-sql-failed:new-entry")
	testutil.WriteTestFile(t, filepath.Join(pluginDir, "manifest", "sql", "001-"+pluginID+".sql"), "THIS IS NOT VALID SQL;")
	if _, err = service.SyncSourcePluginsStrict(ctx); err != nil {
		t.Fatalf("expected source plugin rescan to succeed, got error: %v", err)
	}

	_, err = service.ExecuteRuntimeUpgrade(ctx, pluginID, RuntimeUpgradeOptions{Confirmed: true})
	if !bizerr.Is(err, CodePluginRuntimeUpgradeExecutionFailed) {
		t.Fatalf("expected runtime upgrade execution failure, got %v", err)
	}
	registryAfterFailure, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup after failed source SQL upgrade to succeed, got error: %v", err)
	}
	if registryAfterFailure == nil ||
		registryAfterFailure.Version != oldVersion ||
		registryAfterFailure.ReleaseId != oldRegistry.ReleaseId {
		t.Fatalf("expected effective release to stay %s/%d, got %#v", oldVersion, oldRegistry.ReleaseId, registryAfterFailure)
	}
	item := findPluginItemFromService(t, service, ctx, pluginID)
	if item.RuntimeState != RuntimeUpgradeStateUpgradeFailed || item.LastUpgradeFailure == nil {
		t.Fatalf("expected source SQL failure to be upgrade_failed, got %#v", item)
	}
	if item.LastUpgradeFailure.Phase != plugintypes.RuntimeUpgradeFailurePhaseSQL {
		t.Fatalf("expected SQL failure phase, got %#v", item.LastUpgradeFailure)
	}
	if item.LastUpgradeFailure.MessageKey != store.RuntimeUpgradeFailureMessageKeyMigrationFailed {
		t.Fatalf("expected migration failure message key, got %#v", item.LastUpgradeFailure)
	}
}

// TestUpgradeSourcePluginGovernanceSyncFailureIsDiagnosable verifies source
// strategy governance failures keep the effective release stable and write a
// unified governance failure projection.
func TestUpgradeSourcePluginGovernanceSyncFailureIsDiagnosable(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dev-source-upgrade-governance-failed"
		oldVersion = "v0.1.0"
		newVersion = "v0.5.0"
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Governance Failed Plugin", oldVersion, "plugin:plugin-dev-source-upgrade-governance-failed:old-entry")

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := service.SyncAndList(ctx); err != nil {
		t.Fatalf("expected source plugin discovery to succeed, got error: %v", err)
	}
	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected source plugin install to succeed, got error: %v", err)
	}
	oldRegistry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected source plugin registry lookup to succeed, got error: %v", err)
	}
	if oldRegistry == nil {
		t.Fatal("expected source plugin registry before governance failure")
	}

	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Governance Failed Plugin", newVersion, "plugin:plugin-dev-source-upgrade-governance-failed:new-entry")
	if _, err = service.SyncSourcePluginsStrict(ctx); err != nil {
		t.Fatalf("expected source plugin rescan to succeed, got error: %v", err)
	}
	service.replaceUpgradeServiceForTest(t, failingIntegrationService{
		Service:     service.integrationSvc,
		resourceErr: gerror.New("resource reference sync failed"),
	}, upgradeCachePublisher{service: service}, service.runtimeUpgradeLockStore)

	_, err = service.UpgradeSourcePlugin(ctx, pluginID)
	if err == nil {
		t.Fatal("expected source plugin governance sync failure")
	}
	registryAfterFailure, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected registry lookup after governance failure to succeed, got error: %v", err)
	}
	if registryAfterFailure == nil ||
		registryAfterFailure.Version != oldVersion ||
		registryAfterFailure.ReleaseId != oldRegistry.ReleaseId {
		t.Fatalf("expected effective release to stay %s/%d, got %#v", oldVersion, oldRegistry.ReleaseId, registryAfterFailure)
	}
	item := findPluginItemFromService(t, service, ctx, pluginID)
	if item.RuntimeState != RuntimeUpgradeStateUpgradeFailed || item.LastUpgradeFailure == nil {
		t.Fatalf("expected governance failure to be upgrade_failed, got %#v", item)
	}
	if item.LastUpgradeFailure.Phase != plugintypes.RuntimeUpgradeFailurePhaseGovernance {
		t.Fatalf("expected governance failure phase, got %#v", item.LastUpgradeFailure)
	}
	if item.LastUpgradeFailure.MessageKey != store.RuntimeUpgradeFailureMessageKeyMigrationFailed {
		t.Fatalf("expected migration failure message key, got %#v", item.LastUpgradeFailure)
	}
}

// TestUpgradeSourcePluginCachePublisherFailureReturnsError verifies source
// upgrade success still reports cache publication failures to the caller after
// switching the effective release.
func TestUpgradeSourcePluginCachePublisherFailureReturnsError(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dev-source-upgrade-cache-failed"
		oldVersion = "v0.1.0"
		newVersion = "v0.5.0"
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Cache Failed Plugin", oldVersion, "plugin:plugin-dev-source-upgrade-cache-failed:old-entry")

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := service.SyncAndList(ctx); err != nil {
		t.Fatalf("expected source plugin discovery to succeed, got error: %v", err)
	}
	if _, err := service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected source plugin install to succeed, got error: %v", err)
	}

	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Cache Failed Plugin", newVersion, "plugin:plugin-dev-source-upgrade-cache-failed:new-entry")
	if _, err := service.SyncSourcePluginsStrict(ctx); err != nil {
		t.Fatalf("expected source plugin rescan to succeed, got error: %v", err)
	}
	service.replaceUpgradeServiceForTest(t, service.integrationSvc, failingCachePublisher{
		syncErr: gerror.New("cache publication failed"),
	}, service.runtimeUpgradeLockStore)

	_, err := service.UpgradeSourcePlugin(ctx, pluginID)
	if err == nil {
		t.Fatal("expected source plugin cache publisher failure")
	}
	registry, lookupErr := service.getPluginRegistry(ctx, pluginID)
	if lookupErr != nil {
		t.Fatalf("expected registry lookup after cache publication failure to succeed, got error: %v", lookupErr)
	}
	if registry == nil || registry.Version != newVersion {
		t.Fatalf("expected source upgrade side effects to complete before cache error, got %#v", registry)
	}
}

// TestListSourceUpgradeStatusesSkipsDynamicPlugins verifies development-time
// source-plugin upgrade discovery does not include runtime-managed dynamic plugins.
func TestListSourceUpgradeStatusesSkipsDynamicPlugins(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dev-dynamic-upgrade-boundary"
	)

	testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Upgrade Boundary Plugin",
		"v0.6.0",
		nil,
		nil,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	statuses, err := service.ListSourceUpgradeStatuses(ctx)
	if err != nil {
		t.Fatalf("expected source upgrade status listing to succeed, got error: %v", err)
	}
	if status := findSourceUpgradeStatusByID(statuses, pluginID); status != nil {
		t.Fatalf("expected dynamic plugin to stay outside source upgrade statuses, got %#v", status)
	}
}

// writeTestSourcePluginManifest writes one menu-bearing source plugin manifest for upgrade tests.
func writeTestSourcePluginManifest(
	t *testing.T,
	manifestPath string,
	pluginID string,
	pluginName string,
	version string,
	menuKey string,
) {
	t.Helper()

	testutil.WriteTestFile(
		t,
		manifestPath,
		"id: "+pluginID+"\n"+
			"name: "+pluginName+"\n"+
			"version: "+version+"\n"+
			"type: source\n"+
			"scope_nature: tenant_aware\n"+
			"supports_multi_tenant: false\n"+
			"default_install_mode: global\n"+
			"menus:\n"+
			"  - key: "+menuKey+"\n"+
			"    name: "+pluginName+"\n"+
			"    path: "+pluginID+"\n"+
			"    component: system/plugin/dynamic-page\n"+
			"    perms: "+pluginID+":view\n"+
			"    icon: ant-design:appstore-outlined\n"+
			"    type: M\n"+
			"    sort: -1\n",
	)
}

// registerSourceUpgradeCallbacksForTest replaces the source-plugin fixture
// callbacks while preserving its embedded filesystem declaration.
func registerSourceUpgradeCallbacksForTest(
	t *testing.T,
	pluginID string,
	events *[]string,
	vetoBefore bool,
	failUpgrade bool,
) {
	t.Helper()

	previous, ok := pluginhost.GetSourcePlugin(pluginID)
	if !ok || previous == nil {
		t.Fatalf("expected source plugin fixture %s to be registered", pluginID)
	}
	plugin := pluginhost.NewDeclarations(pluginID)
	plugin.Assets().UseEmbeddedFiles(previous.GetEmbeddedFiles())
	if err := plugin.Lifecycle().RegisterBeforeUpgradeHandler(func(ctx context.Context, input pluginhost.SourcePluginUpgradeInput) (bool, string, error) {
		*events = append(*events, "before:"+input.FromVersion()+"->"+input.ToVersion())
		if input.FromManifest() == nil || input.ToManifest() == nil {
			t.Fatalf("expected upgrade manifest snapshots to be published")
		}
		if input.FromManifest().Version() != input.FromVersion() || input.ToManifest().Version() != input.ToVersion() {
			t.Fatalf("expected snapshot versions to match callback versions")
		}
		if vetoBefore {
			return false, "plugin." + pluginID + ".beforeUpgrade.veto", nil
		}
		return true, "", nil
	}); err != nil {
		t.Fatalf("failed to register before-upgrade callback: %v", err)
	}
	if err := plugin.Lifecycle().RegisterUpgradeHandler(func(ctx context.Context, input pluginhost.SourcePluginUpgradeInput) error {
		*events = append(*events, "upgrade:"+input.FromVersion()+"->"+input.ToVersion())
		if failUpgrade {
			return gerror.New("custom upgrade callback failed")
		}
		return nil
	}); err != nil {
		t.Fatalf("failed to register upgrade callback: %v", err)
	}
	if err := plugin.Lifecycle().RegisterAfterUpgradeHandler(func(ctx context.Context, input pluginhost.SourcePluginUpgradeInput) error {
		*events = append(*events, "after:"+input.FromVersion()+"->"+input.ToVersion())
		return nil
	}); err != nil {
		t.Fatalf("failed to register after-upgrade callback: %v", err)
	}
	cleanup, err := pluginhost.RegisterSourcePluginForTest(plugin)
	if err != nil {
		t.Fatalf("failed to replace source plugin fixture %s: %v", pluginID, err)
	}
	t.Cleanup(cleanup)
}

// sourceUpgradeCodeLifecycleVetoed returns the internal source upgrade code
// used by tests without exporting it from the root plugin service package.
func sourceUpgradeCodeLifecycleVetoed() *bizerr.Code {
	return CodePluginSourceUpgradeLifecycleVetoed
}

// sourceUpgradeTestStringSlicesEqual reports whether two ordered string slices are equal.
func sourceUpgradeTestStringSlicesEqual(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

// findSourceUpgradeStatusByID returns one source-plugin upgrade status by plugin ID.
func findSourceUpgradeStatusByID(
	items []*SourceUpgradeStatus,
	pluginID string,
) *SourceUpgradeStatus {
	for _, item := range items {
		if item != nil && item.PluginID == pluginID {
			return item
		}
	}
	return nil
}

// replaceUpgradeServiceForTest rebuilds the unified upgrade owner with one
// injected integration service and cache publisher.
func (s *serviceImpl) replaceUpgradeServiceForTest(
	t *testing.T,
	integrationSvc integration.Service,
	cachePublisher interface {
		PublishPluginChange(context.Context, string, string, string) error
		SyncEnabledSnapshotAndPublishRuntimeChange(context.Context, string, string) error
	},
	lockStore coordination.LockStore,
) {
	t.Helper()
	upgradeSvc, err := upgrade.New(
		s.catalogSvc,
		s.storeSvc,
		s.runtimeSvc,
		integrationSvc,
		s.migrationSvc,
		plugindep.New(),
		s.i18nSvc,
		lockStore,
		cachePublisher,
		upgradeCacheFreshener{service: s},
		s.topology,
		s.configSvc,
	)
	if err != nil {
		t.Fatalf("expected replacement upgrade service to build, got error: %v", err)
	}
	if err = s.lifecycleSvc.BindUpgrade(upgradeSvc); err != nil {
		t.Fatalf("expected replacement upgrade service to bind, got error: %v", err)
	}
}

// failingIntegrationService injects source-upgrade governance failures while
// delegating every unrelated integration method.
type failingIntegrationService struct {
	integration.Service
	menuErr     error
	resourceErr error
}

// SyncPluginMenusAndPermissions optionally fails menu and permission sync.
func (s failingIntegrationService) SyncPluginMenusAndPermissions(ctx context.Context, manifest *catalog.Manifest) error {
	if s.menuErr != nil {
		return s.menuErr
	}
	return s.Service.SyncPluginMenusAndPermissions(ctx, manifest)
}

// SyncPluginResourceReferences optionally fails resource-reference sync.
func (s failingIntegrationService) SyncPluginResourceReferences(ctx context.Context, manifest *catalog.Manifest) error {
	if s.resourceErr != nil {
		return s.resourceErr
	}
	return s.Service.SyncPluginResourceReferences(ctx, manifest)
}

// failingCachePublisher injects source-upgrade cache publication failures.
type failingCachePublisher struct {
	publishErr error
	syncErr    error
}

// PublishPluginChange optionally fails plugin-scoped mutation publication.
func (p failingCachePublisher) PublishPluginChange(context.Context, string, string, string) error {
	return p.publishErr
}

// SyncEnabledSnapshotAndPublishRuntimeChange optionally fails source upgrade success publication.
func (p failingCachePublisher) SyncEnabledSnapshotAndPublishRuntimeChange(context.Context, string, string) error {
	return p.syncErr
}
