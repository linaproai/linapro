// This file verifies source-plugin upgrade governance, including effective
// version pinning, startup fail-fast validation, and explicit upgrade flow.

package plugin

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/testutil"
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

// TestValidateSourcePluginUpgradeReadinessFailsForPendingUpgrade verifies startup
// validation blocks host boot when an installed source plugin still has a newer
// discovered version waiting for explicit upgrade.
func TestValidateSourcePluginUpgradeReadinessFailsForPendingUpgrade(t *testing.T) {
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
	if err == nil {
		t.Fatal("expected startup validation to fail for pending source plugin upgrade")
	}
	message := err.Error()
	if !strings.Contains(message, pluginID) ||
		!strings.Contains(message, oldVersion) ||
		!strings.Contains(message, newVersion) ||
		!strings.Contains(message, "action=resolve the source-plugin version before startup") {
		t.Fatalf("expected startup validation error to include plugin, versions, and action hint, got %q", message)
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
