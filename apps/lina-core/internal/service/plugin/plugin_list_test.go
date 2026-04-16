// This file covers root-facade list methods defined in plugin_list.go.

package plugin

import (
	"context"
	"os"
	"testing"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/testutil"
)

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
				Type:   "M",
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
	if err = service.Install(ctx, pluginID, nil); err != nil {
		t.Fatalf("expected dynamic plugin install to succeed, got error: %v", err)
	}
	if err = service.Uninstall(ctx, pluginID); err != nil {
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
