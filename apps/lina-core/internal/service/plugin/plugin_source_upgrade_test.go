// This file verifies source-plugin upgrade governance, including effective
// version pinning, non-blocking runtime drift projection, and explicit upgrade flow.

package plugin

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/plugin/internal/catalog"
	sourceupgradeinternal "lina-core/internal/service/plugin/internal/sourceupgrade"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/pluginhost"
)

// TestSourcePluginDiscoveryKeepsEffectiveVersionAfterHigherSourceVersion verifies
// discovered source versions do not overwrite the current effective registry version.
func TestSourcePluginDiscoveryKeepsEffectiveVersionAfterHigherSourceVersion(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-source-upgrade-drift"
		oldVersion = "v0.1.0"
		newVersion = "v0.5.0"
		oldMenuKey = "plugin:plugin-source-upgrade-drift:old-entry"
		newMenuKey = "plugin:plugin-source-upgrade-drift:new-entry"
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
	if err := service.SyncSourcePlugins(ctx); err != nil {
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
	if preparedRelease.Status != catalog.ReleaseStatusPrepared.String() {
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
		pluginID   = "plugin-source-upgrade-startup-guard"
		oldVersion = "v0.1.0"
		newVersion = "v0.5.0"
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Startup Guard Plugin", oldVersion, "plugin:plugin-source-upgrade-startup-guard:old-entry")

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

	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Startup Guard Plugin", newVersion, "plugin:plugin-source-upgrade-startup-guard:new-entry")
	if err := service.SyncSourcePlugins(ctx); err != nil {
		t.Fatalf("expected source plugin rescan to succeed, got error: %v", err)
	}

	err := service.ValidateSourcePluginUpgradeReadiness(ctx)
	if err != nil {
		t.Fatalf("expected source upgrade readiness scan not to fail for pending runtime upgrade, got error: %v", err)
	}

	out, err := service.List(ctx, ListInput{})
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
		pluginID   = "plugin-source-upgrade-abnormal"
		oldVersion = "v0.1.0"
		newVersion = "v0.5.0"
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Abnormal Plugin", newVersion, "plugin:plugin-source-upgrade-abnormal:new-entry")

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

	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Abnormal Plugin", oldVersion, "plugin:plugin-source-upgrade-abnormal:old-entry")
	if err := service.SyncSourcePlugins(ctx); err != nil {
		t.Fatalf("expected source plugin rescan to succeed, got error: %v", err)
	}

	out, err := service.List(ctx, ListInput{})
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
		pluginID   = "plugin-source-upgrade-apply"
		oldVersion = "v0.1.0"
		newVersion = "v0.5.0"
		oldMenuKey = "plugin:plugin-source-upgrade-apply:old-entry"
		newMenuKey = "plugin:plugin-source-upgrade-apply:new-entry"
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
	if err := service.Enable(ctx, pluginID); err != nil {
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
	if err := service.SyncSourcePlugins(ctx); err != nil {
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
	if registry.Status != catalog.StatusEnabled {
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
	if newRelease.Status != catalog.ReleaseStatusActive.String() {
		t.Fatalf("expected new source plugin release to become active, got %s", newRelease.Status)
	}

	oldRelease, err = service.getPluginRelease(ctx, pluginID, oldVersion)
	if err != nil {
		t.Fatalf("expected old source plugin release lookup after upgrade to succeed, got error: %v", err)
	}
	if oldRelease == nil {
		t.Fatal("expected old source plugin release row after upgrade")
	}
	if oldRelease.Status != catalog.ReleaseStatusInstalled.String() {
		t.Fatalf("expected old source plugin release to be demoted to installed, got %s", oldRelease.Status)
	}

	upgradeMigrationCount, err := dao.SysPluginMigration.Ctx(ctx).
		Where(do.SysPluginMigration{
			PluginId:  pluginID,
			ReleaseId: newRelease.Id,
			Phase:     catalog.MigrationDirectionUpgrade.String(),
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
		pluginID   = "plugin-source-upgrade-callback"
		oldVersion = "v0.1.0"
		newVersion = "v0.5.0"
		events     []string
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Callback Plugin", oldVersion, "plugin:plugin-source-upgrade-callback:old-entry")
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

	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Callback Plugin", newVersion, "plugin:plugin-source-upgrade-callback:new-entry")
	registerSourceUpgradeCallbacksForTest(t, pluginID, &events, false, false)
	if err := service.SyncSourcePlugins(ctx); err != nil {
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
		pluginID   = "plugin-source-upgrade-before-veto"
		oldVersion = "v0.1.0"
		newVersion = "v0.5.0"
		events     []string
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Before Veto Plugin", oldVersion, "plugin:plugin-source-upgrade-before-veto:old-entry")
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

	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Before Veto Plugin", newVersion, "plugin:plugin-source-upgrade-before-veto:new-entry")
	registerSourceUpgradeCallbacksForTest(t, pluginID, &events, true, false)
	if err := service.SyncSourcePlugins(ctx); err != nil {
		t.Fatalf("expected source plugin rescan to succeed, got error: %v", err)
	}

	_, err := service.UpgradeSourcePlugin(ctx, pluginID)
	if !bizerr.Is(err, sourceupgradeCodeLifecycleVetoed()) {
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
	if item.LastUpgradeFailure.Phase != catalog.RuntimeUpgradeFailurePhaseBeforeUpgrade {
		t.Fatalf("expected before_upgrade failure phase, got %#v", item.LastUpgradeFailure)
	}
}

// TestUpgradeSourcePluginCallbackFailureIsRetryable verifies custom upgrade
// callback failures keep the effective release stable and allow a later retry.
func TestUpgradeSourcePluginCallbackFailureIsRetryable(t *testing.T) {
	var (
		service    = newTestService()
		ctx        = context.Background()
		pluginID   = "plugin-source-upgrade-callback-retry"
		oldVersion = "v0.1.0"
		newVersion = "v0.5.0"
		events     []string
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Callback Retry Plugin", oldVersion, "plugin:plugin-source-upgrade-callback-retry:old-entry")
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

	writeTestSourcePluginManifest(t, manifestPath, pluginID, "Source Upgrade Callback Retry Plugin", newVersion, "plugin:plugin-source-upgrade-callback-retry:new-entry")
	registerSourceUpgradeCallbacksForTest(t, pluginID, &events, false, true)
	if err := service.SyncSourcePlugins(ctx); err != nil {
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
	if item.LastUpgradeFailure.Phase != catalog.RuntimeUpgradeFailurePhaseUpgradeCallback {
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

// TestListSourceUpgradeStatusesSkipsDynamicPlugins verifies development-time
// source-plugin upgrade discovery does not include runtime-managed dynamic plugins.
func TestListSourceUpgradeStatusesSkipsDynamicPlugins(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-upgrade-boundary"
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
	plugin := pluginhost.NewSourcePlugin(pluginID)
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

// sourceupgradeCodeLifecycleVetoed returns the internal source-upgrade code
// used by tests without exporting it from the root plugin service package.
func sourceupgradeCodeLifecycleVetoed() *bizerr.Code {
	return sourceupgradeinternal.CodePluginSourceUpgradeLifecycleVetoed
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
