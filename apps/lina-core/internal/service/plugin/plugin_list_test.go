// This file covers root-facade list methods defined in plugin_list.go.

package plugin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/gogf/gf/v2/os/gctx"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/pluginbridge"
)

// TestSyncAndListRetainsMissingRuntimeRegistryAndReconcilesState verifies that
// missing runtime artifacts reconcile registry state without hiding the plugin.
func TestSyncAndListRetainsMissingRuntimeRegistryAndReconcilesState(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-registry-missing"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		"Runtime Registry Missing Plugin",
		"v0.9.4",
		nil,
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
	if err = service.setPluginInstalled(ctx, pluginID, catalog.InstalledYes); err != nil {
		t.Fatalf("expected dynamic plugin install state to be set, got error: %v", err)
	}
	if err = service.setPluginStatus(ctx, pluginID, catalog.StatusEnabled); err != nil {
		t.Fatalf("expected dynamic plugin enable state to be set, got error: %v", err)
	}
	if err = os.Remove(artifactPath); err != nil {
		t.Fatalf("failed to remove dynamic artifact: %v", err)
	}

	out, err := service.SyncAndList(ctx)
	if err != nil {
		t.Fatalf("expected sync-and-list to tolerate missing dynamic artifact, got error: %v", err)
	}

	var item *PluginItem
	for _, current := range out.List {
		if current != nil && current.Id == pluginID {
			item = current
			break
		}
	}
	if item == nil {
		t.Fatalf("expected missing dynamic plugin to remain visible in plugin list")
	}
	if item.Installed != catalog.InstalledNo {
		t.Fatalf("expected missing dynamic plugin installed state to reconcile to %d, got %d", catalog.InstalledNo, item.Installed)
	}
	if item.Enabled != catalog.StatusDisabled {
		t.Fatalf("expected missing dynamic plugin enabled state to reconcile to %d, got %d", catalog.StatusDisabled, item.Enabled)
	}

	runtimeStates, err := service.ListRuntimeStates(ctx)
	if err != nil {
		t.Fatalf("expected runtime state list to succeed, got error: %v", err)
	}
	var runtimeState *PluginDynamicStateItem
	for _, current := range runtimeStates.List {
		if current != nil && current.Id == pluginID {
			runtimeState = current
			break
		}
	}
	if runtimeState == nil {
		t.Fatalf("expected missing dynamic plugin to remain visible in public runtime states")
	}
	if runtimeState.Installed != catalog.InstalledNo || runtimeState.Enabled != catalog.StatusDisabled {
		t.Fatalf("expected public runtime state to reconcile to uninstalled+disabled, got installed=%d enabled=%d", runtimeState.Installed, runtimeState.Enabled)
	}
	if runtimeState.RuntimeState != RuntimeUpgradeStateNormal {
		t.Fatalf("expected missing dynamic plugin public runtime state to stay normal, got %s", runtimeState.RuntimeState)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected runtime registry lookup to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatalf("expected runtime registry row to remain after reconciliation")
	}
	if registry.Installed != catalog.InstalledNo || registry.Status != catalog.StatusDisabled {
		t.Fatalf("expected runtime registry row to reconcile to uninstalled+disabled, got installed=%d enabled=%d", registry.Installed, registry.Status)
	}
}

// TestListProjectsMissingRuntimeRegistryWithoutWriting verifies the GET-list
// path can show a safe missing-artifact state without mutating governance rows.
func TestListProjectsMissingRuntimeRegistryWithoutWriting(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-registry-readonly"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		"Runtime Registry Readonly Plugin",
		"v0.9.5",
		nil,
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
	if err = service.setPluginInstalled(ctx, pluginID, catalog.InstalledYes); err != nil {
		t.Fatalf("expected dynamic plugin install state to be set, got error: %v", err)
	}
	if err = service.setPluginStatus(ctx, pluginID, catalog.StatusEnabled); err != nil {
		t.Fatalf("expected dynamic plugin enable state to be set, got error: %v", err)
	}

	registryBefore, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected runtime registry lookup before list to succeed, got error: %v", err)
	}
	if registryBefore == nil {
		t.Fatalf("expected runtime registry row before list")
	}
	if err = os.Remove(artifactPath); err != nil {
		t.Fatalf("failed to remove dynamic artifact: %v", err)
	}

	out, err := service.List(ctx, ListInput{})
	if err != nil {
		t.Fatalf("expected read-only list to tolerate missing dynamic artifact, got error: %v", err)
	}

	item := findPluginItem(out, pluginID)
	if item == nil {
		t.Fatalf("expected missing dynamic plugin to remain visible in read-only plugin list")
	}
	if item.Installed != catalog.InstalledNo {
		t.Fatalf("expected read-only projection installed state to be %d, got %d", catalog.InstalledNo, item.Installed)
	}
	if item.Enabled != catalog.StatusDisabled {
		t.Fatalf("expected read-only projection enabled state to be %d, got %d", catalog.StatusDisabled, item.Enabled)
	}

	registryAfter, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected runtime registry lookup after list to succeed, got error: %v", err)
	}
	if registryAfter == nil {
		t.Fatalf("expected runtime registry row to remain after read-only list")
	}
	if registryAfter.Installed != registryBefore.Installed ||
		registryAfter.Status != registryBefore.Status ||
		registryAfter.DesiredState != registryBefore.DesiredState ||
		registryAfter.CurrentState != registryBefore.CurrentState ||
		registryAfter.Generation != registryBefore.Generation ||
		registryAfter.ReleaseId != registryBefore.ReleaseId {
		t.Fatalf(
			"expected read-only list not to mutate registry, before installed=%d status=%d desired=%s current=%s generation=%d release=%d after installed=%d status=%d desired=%s current=%s generation=%d release=%d",
			registryBefore.Installed,
			registryBefore.Status,
			registryBefore.DesiredState,
			registryBefore.CurrentState,
			registryBefore.Generation,
			registryBefore.ReleaseId,
			registryAfter.Installed,
			registryAfter.Status,
			registryAfter.DesiredState,
			registryAfter.CurrentState,
			registryAfter.Generation,
			registryAfter.ReleaseId,
		)
	}
}

// TestGetReturnsStableNotFoundBizerr verifies exact detail lookup reports a
// stable business error when no discovered or registered plugin matches.
func TestGetReturnsStableNotFoundBizerr(t *testing.T) {
	service := newTestService()
	_, err := service.Get(context.Background(), "plugin-detail-missing")
	if !bizerr.Is(err, CodePluginNotFound) {
		t.Fatalf("expected plugin not-found bizerr, got %v", err)
	}
}

// TestListMarksInstalledDynamicPluginWithHigherArtifactPendingUpgrade verifies
// dynamic artifact replacement is exposed as a pending runtime upgrade without
// switching the effective registry version.
func TestListMarksInstalledDynamicPluginWithHigherArtifactPendingUpgrade(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dynamic-runtime-upgrade-pending"
		oldVersion = "v0.1.0"
		newVersion = "v0.2.0"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Runtime Upgrade Pending Plugin",
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
		t.Fatalf("expected old dynamic release lookup to succeed, got error: %v", err)
	}
	if oldRelease == nil {
		t.Fatal("expected old dynamic release row")
	}

	testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Runtime Upgrade Pending Plugin",
		newVersion,
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

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected dynamic registry lookup to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatal("expected dynamic registry row")
	}
	if registry.Version != oldVersion {
		t.Fatalf("expected effective version %s to stay pinned, got %s", oldVersion, registry.Version)
	}
	if registry.ReleaseId != oldRelease.Id {
		t.Fatalf("expected effective release_id %d to stay pinned, got %d", oldRelease.Id, registry.ReleaseId)
	}

	out, err := service.List(ctx, ListInput{})
	if err != nil {
		t.Fatalf("expected plugin list to succeed, got error: %v", err)
	}
	item := findPluginItem(out, pluginID)
	if item == nil {
		t.Fatal("expected dynamic plugin list item")
	}
	if item.RuntimeState != RuntimeUpgradeStatePendingUpgrade {
		t.Fatalf("expected runtime state %s, got %#v", RuntimeUpgradeStatePendingUpgrade, item)
	}
	if item.EffectiveVersion != oldVersion || item.DiscoveredVersion != newVersion {
		t.Fatalf("expected effective/discovered versions %s/%s, got %#v", oldVersion, newVersion, item)
	}
	if !item.UpgradeAvailable {
		t.Fatalf("expected dynamic plugin to report upgradeAvailable, got %#v", item)
	}

	detail, err := service.Get(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected plugin detail to succeed, got error: %v", err)
	}
	if detail.RuntimeState != RuntimeUpgradeStatePendingUpgrade {
		t.Fatalf("expected detail runtime state %s, got %#v", RuntimeUpgradeStatePendingUpgrade, detail)
	}
	if detail.EffectiveVersion != oldVersion || detail.DiscoveredVersion != newVersion {
		t.Fatalf("expected detail effective/discovered versions %s/%s, got %#v", oldVersion, newVersion, detail)
	}
	if !detail.UpgradeAvailable {
		t.Fatalf("expected detail to report upgradeAvailable, got %#v", detail)
	}
}

// TestPreviewRuntimeUpgradeReturnsPendingDynamicPlan verifies that preview is
// read-only and exposes manifest snapshots, dependency checks, SQL summary,
// hostServices drift, and stable risk hints for a pending dynamic upgrade.
func TestPreviewRuntimeUpgradeReturnsPendingDynamicPlan(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dynamic-runtime-upgrade-preview"
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
			Type:    catalog.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: pluginbridge.RuntimeKindWasm,
			ABIVersion:  pluginbridge.SupportedABIVersion,
			HostServices: []*pluginbridge.HostServiceSpec{
				{
					Service: pluginbridge.HostServiceStorage,
					Methods: []string{pluginbridge.HostServiceMethodStorageGet},
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
			Type:    catalog.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: pluginbridge.RuntimeKindWasm,
			ABIVersion:  pluginbridge.SupportedABIVersion,
			HostServices: []*pluginbridge.HostServiceSpec{
				{
					Service: pluginbridge.HostServiceStorage,
					Methods: []string{
						pluginbridge.HostServiceMethodStorageGet,
						pluginbridge.HostServiceMethodStoragePut,
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
	if change.Service != pluginbridge.HostServiceStorage {
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
		pluginID = "plugin-dynamic-runtime-upgrade-preview-normal"
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

// TestListMarksInstalledDynamicPluginWithFailedTargetReleaseUpgradeFailed verifies
// failed target releases stay visible as retryable runtime-upgrade failures.
func TestListMarksInstalledDynamicPluginWithFailedTargetReleaseUpgradeFailed(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dynamic-runtime-upgrade-failed"
		oldVersion = "v0.1.0"
		newVersion = "v0.2.0"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Runtime Upgrade Failed Plugin",
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

	testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Runtime Upgrade Failed Plugin",
		newVersion,
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

	targetRelease, err := service.getPluginRelease(ctx, pluginID, newVersion)
	if err != nil {
		t.Fatalf("expected target release lookup to succeed, got error: %v", err)
	}
	if targetRelease == nil {
		t.Fatal("expected target release row")
	}
	if err = service.catalogSvc.UpdateReleaseState(
		ctx,
		targetRelease.Id,
		catalog.ReleaseStatusFailed,
		"",
	); err != nil {
		t.Fatalf("expected target release failure state update to succeed, got error: %v", err)
	}

	out, err := service.SyncAndList(ctx)
	if err != nil {
		t.Fatalf("expected sync-and-list to preserve failed target release, got error: %v", err)
	}
	item := findPluginItem(out, pluginID)
	if item == nil {
		t.Fatal("expected dynamic plugin list item")
	}
	if item.RuntimeState != RuntimeUpgradeStateUpgradeFailed {
		t.Fatalf("expected runtime state %s, got %#v", RuntimeUpgradeStateUpgradeFailed, item)
	}
	if item.LastUpgradeFailure == nil {
		t.Fatalf("expected last upgrade failure details, got %#v", item)
	}
	if item.LastUpgradeFailure.ReleaseID != targetRelease.Id ||
		item.LastUpgradeFailure.ReleaseVersion != newVersion {
		t.Fatalf("expected failed release %d/%s, got %#v", targetRelease.Id, newVersion, item.LastUpgradeFailure)
	}

	failedRelease, err := service.getPluginRelease(ctx, pluginID, newVersion)
	if err != nil {
		t.Fatalf("expected failed release lookup to succeed, got error: %v", err)
	}
	if failedRelease == nil || failedRelease.Status != catalog.ReleaseStatusFailed.String() {
		t.Fatalf("expected target release to remain failed after sync, got %#v", failedRelease)
	}
}

// TestFilterMenusHidesPendingUpgradePluginMenus verifies plugin-owned menus are
// hidden while a plugin waits for runtime upgrade.
func TestFilterMenusHidesPendingUpgradePluginMenus(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-dynamic-menu-pending-upgrade"
		oldVersion = "v0.1.0"
		newVersion = "v0.2.0"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Menu Pending Upgrade Plugin",
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

	testutil.CreateTestRuntimeStorageArtifact(
		t,
		pluginID,
		"Dynamic Menu Pending Upgrade Plugin",
		newVersion,
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
	if err = service.integrationSvc.RefreshEnabledSnapshot(ctx); err != nil {
		t.Fatalf("expected enabled snapshot refresh to succeed, got error: %v", err)
	}

	filtered := service.FilterMenus(ctx, []*entity.SysMenu{
		{
			Id:      1,
			MenuKey: "plugin:" + pluginID + ":entry",
			Name:    "runtime menu",
			Type:    catalog.MenuTypePage.String(),
			Status:  1,
			Visible: 1,
		},
	})
	if len(filtered) != 0 {
		t.Fatalf("expected pending-upgrade plugin menu to be hidden, got %d entries", len(filtered))
	}
}

// TestListLocalizesUninstalledDynamicPluginMetadataInEnglish verifies that
// plugin management can display artifact-owned metadata before installation.
func TestListLocalizesUninstalledDynamicPluginMetadataInEnglish(t *testing.T) {
	var (
		service      = newTestService()
		ctx          = context.WithValue(context.Background(), gctx.StrKey("BizCtx"), &model.Context{Locale: i18nsvc.EnglishLocale})
		pluginID     = "plugin-dynamic-list-i18n"
		artifactPath = filepath.Join(testutil.TestDynamicStorageDir(), pluginID+".wasm")
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		if err := os.Remove(artifactPath); err != nil && !os.IsNotExist(err) {
			t.Fatalf("failed to remove dynamic i18n test artifact %s: %v", artifactPath, err)
		}
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:          pluginID,
			Name:        "动态插件列表中文名",
			Version:     "v0.9.8",
			Type:        catalog.TypeDynamic.String(),
			Description: "未安装动态插件的中文描述",
		},
		&catalog.ArtifactSpec{
			RuntimeKind:        pluginbridge.RuntimeKindWasm,
			ABIVersion:         pluginbridge.SupportedABIVersion,
			FrontendAssetCount: len(testutil.DefaultTestRuntimeFrontendAssets()),
		},
		testutil.DefaultTestRuntimeFrontendAssets(),
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	appendRuntimeI18NSectionForPluginListTest(
		t,
		artifactPath,
		[]map[string]string{
			{
				"locale": "en-US",
				"content": `{
  "plugin": {
    "plugin-dynamic-list-i18n": {
      "name": "Dynamic List I18N Plugin",
      "description": "English dynamic plugin description before installation."
    }
  }
}`,
			},
		},
	)

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic i18n artifact manifest to load, got error: %v", err)
	}
	registry, err := service.syncPluginManifest(ctx, manifest)
	if err != nil {
		t.Fatalf("expected dynamic i18n manifest sync to succeed, got error: %v", err)
	}
	if registry == nil || registry.Installed != catalog.InstalledNo {
		t.Fatalf("expected dynamic i18n plugin to remain uninstalled after sync, got %#v", registry)
	}

	out, err := service.List(ctx, ListInput{})
	if err != nil {
		t.Fatalf("expected plugin list to succeed, got error: %v", err)
	}
	item := findPluginItem(out, pluginID)
	if item == nil {
		t.Fatalf("expected dynamic i18n plugin to appear in plugin list")
	}
	if item.Name != "Dynamic List I18N Plugin" {
		t.Fatalf("expected English plugin name before install, got %q", item.Name)
	}
	if item.Description != "English dynamic plugin description before installation." {
		t.Fatalf("expected English plugin description before install, got %q", item.Description)
	}
	if item.Installed != catalog.InstalledNo {
		t.Fatalf("expected plugin to remain not installed, got %d", item.Installed)
	}
}

// TestSyncAndListDoesNotRestoreUninstalledDynamicGovernanceProjection verifies
// that sync does not recreate release-bound governance after uninstall.
func TestSyncAndListDoesNotRestoreUninstalledDynamicGovernanceProjection(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-uninstall-governance"
	)

	testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithMenus(
		t,
		pluginID,
		"Dynamic Uninstall Governance Plugin",
		"v0.3.1",
		[]*catalog.MenuSpec{
			{
				Key:    "plugin:plugin-dynamic-uninstall-governance:entry",
				Name:   "Dynamic Uninstall Governance Plugin",
				Path:   "plugin-dynamic-uninstall-governance-entry",
				Perms:  "plugin-dynamic-uninstall-governance:view",
				Icon:   "ant-design:appstore-outlined",
				Type:   catalog.MenuTypePage.String(),
				Sort:   1,
				Remark: "Runtime uninstall governance verification menu.",
			},
		},
		nil,
		nil,
	)

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected runtime artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic manifest sync to succeed, got error: %v", err)
	}
	if _, err = service.Install(ctx, pluginID, InstallOptions{}); err != nil {
		t.Fatalf("expected dynamic plugin install to succeed, got error: %v", err)
	}
	if err = service.Uninstall(ctx, pluginID, UninstallOptions{PurgeStorageData: true}); err != nil {
		t.Fatalf("expected dynamic plugin uninstall to succeed, got error: %v", err)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected runtime registry lookup to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatalf("expected runtime registry row to exist after uninstall")
	}
	if registry.ReleaseId != 0 {
		t.Fatalf("expected runtime registry release_id to be cleared after uninstall, got %d", registry.ReleaseId)
	}

	resourceCount, err := dao.SysPluginResourceRef.Ctx(ctx).
		Where(do.SysPluginResourceRef{PluginId: pluginID}).
		Count()
	if err != nil {
		t.Fatalf("expected governance resource count query to succeed, got error: %v", err)
	}
	if resourceCount != 0 {
		t.Fatalf("expected uninstall to clear governance resource refs, got count=%d", resourceCount)
	}

	if _, err = service.SyncAndList(ctx); err != nil {
		t.Fatalf("expected sync-and-list to succeed after uninstall, got error: %v", err)
	}

	registry, err = service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected runtime registry lookup after sync-and-list to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatalf("expected runtime registry row to remain after sync-and-list")
	}
	if registry.ReleaseId != 0 {
		t.Fatalf("expected sync-and-list not to restore release_id for uninstalled plugin, got %d", registry.ReleaseId)
	}

	resourceCount, err = dao.SysPluginResourceRef.Ctx(ctx).
		Where(do.SysPluginResourceRef{PluginId: pluginID}).
		Count()
	if err != nil {
		t.Fatalf("expected governance resource count query after sync-and-list to succeed, got error: %v", err)
	}
	if resourceCount != 0 {
		t.Fatalf("expected sync-and-list not to recreate governance resource refs for uninstalled plugin, got count=%d", resourceCount)
	}
}

// findPluginItem returns one plugin list item by plugin ID for list assertions.
func findPluginItem(out *ListOutput, pluginID string) *PluginItem {
	if out == nil {
		return nil
	}
	for _, current := range out.List {
		if current != nil && current.Id == pluginID {
			return current
		}
	}
	return nil
}

// containsString reports whether values contains target.
func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

// appendRuntimeI18NSectionForPluginListTest appends one runtime i18n custom
// section to the synthetic wasm artifact used by plugin list localization tests.
func appendRuntimeI18NSectionForPluginListTest(
	t *testing.T,
	artifactPath string,
	payload any,
) {
	t.Helper()

	content, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("expected runtime artifact read to succeed, got error: %v", err)
	}
	sectionPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("expected runtime i18n payload marshal to succeed, got error: %v", err)
	}
	content = appendPluginListTestWasmCustomSection(
		content,
		pluginbridge.WasmSectionI18NAssets,
		sectionPayload,
	)
	if err = os.WriteFile(artifactPath, content, 0o644); err != nil {
		t.Fatalf("expected runtime artifact write to succeed, got error: %v", err)
	}
}

// appendPluginListTestWasmCustomSection appends one custom section using WASM
// section-length encoding.
func appendPluginListTestWasmCustomSection(content []byte, name string, payload []byte) []byte {
	sectionPayload := append([]byte{}, encodePluginListTestWasmULEB128(uint32(len(name)))...)
	sectionPayload = append(sectionPayload, []byte(name)...)
	sectionPayload = append(sectionPayload, payload...)

	result := append([]byte{}, content...)
	result = append(result, 0x00)
	result = append(result, encodePluginListTestWasmULEB128(uint32(len(sectionPayload)))...)
	result = append(result, sectionPayload...)
	return result
}

// encodePluginListTestWasmULEB128 encodes one unsigned integer for custom sections.
func encodePluginListTestWasmULEB128(value uint32) []byte {
	result := make([]byte, 0, 5)
	for {
		current := byte(value & 0x7f)
		value >>= 7
		if value != 0 {
			current |= 0x80
		}
		result = append(result, current)
		if value == 0 {
			return result
		}
	}
}
