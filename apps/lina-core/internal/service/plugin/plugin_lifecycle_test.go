// This file covers root-facade lifecycle methods defined in plugin_lifecycle.go.

package plugin

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/pluginbridge"
)

func TestUpdateStatusEnablesBackendOnlyDynamicPluginWithoutFrontendAssets(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-backend-only"
	)

	frontend.ResetBundleCache()
	t.Cleanup(frontend.ResetBundleCache)

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		"Backend Only Dynamic Plugin",
		"v0.4.1",
		nil,
		nil,
		nil,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	manifest, err := service.loadRuntimePluginManifestFromArtifact(artifactPath)
	if err != nil {
		t.Fatalf("expected backend-only artifact manifest to load, got error: %v", err)
	}
	if _, err = service.syncPluginManifest(ctx, manifest); err != nil {
		t.Fatalf("expected backend-only artifact sync to succeed, got error: %v", err)
	}
	if err = service.setPluginInstalled(ctx, pluginID, catalog.InstalledYes); err != nil {
		t.Fatalf("expected backend-only plugin install state to be set, got error: %v", err)
	}

	if err = service.UpdateStatus(ctx, pluginID, catalog.StatusEnabled, nil); err != nil {
		t.Fatalf("expected backend-only dynamic plugin enable to succeed, got error: %v", err)
	}
	if !service.IsEnabled(ctx, pluginID) {
		t.Fatalf("expected backend-only dynamic plugin to be enabled after status update")
	}
}

func TestSyncAndListReportsPendingHostServiceAuthorization(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-host-auth-pending"
	)

	artifactPath := filepath.Join(
		testutil.TestDynamicStorageDir(),
		pluginID+".wasm",
	)
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      pluginID,
			Name:    "Pending Authorization Plugin",
			Version: "v0.5.0",
			Type:    catalog.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: pluginbridge.RuntimeKindWasm,
			ABIVersion:  pluginbridge.SupportedABIVersion,
			HostServices: []*pluginbridge.HostServiceSpec{
				{
					Service: pluginbridge.HostServiceRuntime,
					Methods: []string{pluginbridge.HostServiceMethodRuntimeInfoNow},
				},
				{
					Service: pluginbridge.HostServiceNetwork,
					Methods: []string{pluginbridge.HostServiceMethodNetworkRequest},
					Resources: []*pluginbridge.HostServiceResourceSpec{
						{
							Ref: "https://example.com/api",
						},
					},
				},
			},
		},
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	out, err := service.SyncAndList(ctx)
	if err != nil {
		t.Fatalf("expected sync-and-list to succeed, got error: %v", err)
	}

	var item *PluginItem
	for _, current := range out.List {
		if current != nil && current.Id == pluginID {
			item = current
			break
		}
	}
	if item == nil {
		t.Fatalf("expected pending authorization plugin in list")
	}
	if !item.AuthorizationRequired {
		t.Fatalf("expected pending authorization plugin to require review")
	}
	if item.AuthorizationStatus != runtime.AuthorizationStatusPending {
		t.Fatalf("expected authorization status pending, got %s", item.AuthorizationStatus)
	}
	if len(item.RequestedHostServices) != 2 {
		t.Fatalf("expected requested host services to be exposed, got %#v", item.RequestedHostServices)
	}
	if len(item.AuthorizedHostServices) != 0 {
		t.Fatalf("expected no authorized host services before confirmation, got %#v", item.AuthorizedHostServices)
	}
}

func TestEnableWithAuthorizationAppliesConfirmedHostServiceSnapshot(t *testing.T) {
	var (
		service  = newTestService()
		ctx      = context.Background()
		pluginID = "plugin-dynamic-host-auth-enabled"
	)

	artifactPath := filepath.Join(
		testutil.TestDynamicStorageDir(),
		pluginID+".wasm",
	)
	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      pluginID,
			Name:    "Confirmed Authorization Plugin",
			Version: "v0.5.1",
			Type:    catalog.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind: pluginbridge.RuntimeKindWasm,
			ABIVersion:  pluginbridge.SupportedABIVersion,
			HostServices: []*pluginbridge.HostServiceSpec{
				{
					Service: pluginbridge.HostServiceRuntime,
					Methods: []string{pluginbridge.HostServiceMethodRuntimeInfoNow},
				},
				{
					Service: pluginbridge.HostServiceNetwork,
					Methods: []string{pluginbridge.HostServiceMethodNetworkRequest},
					Resources: []*pluginbridge.HostServiceResourceSpec{
						{
							Ref: "https://example.com/api",
						},
					},
				},
				{
					Service: pluginbridge.HostServiceStorage,
					Methods: []string{pluginbridge.HostServiceMethodStorageGet},
					Paths:   []string{"private-files/"},
				},
			},
		},
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	authorization := &HostServiceAuthorizationInput{
		Services: []*HostServiceAuthorizationDecision{
			{
				Service: pluginbridge.HostServiceStorage,
				Paths:   []string{"private-files/"},
			},
		},
	}

	if err := service.Install(ctx, pluginID, authorization); err != nil {
		t.Fatalf("expected install with authorization to succeed, got error: %v", err)
	}
	if err := service.UpdateStatus(ctx, pluginID, catalog.StatusEnabled, authorization); err != nil {
		t.Fatalf("expected enable with authorization to succeed, got error: %v", err)
	}

	release, err := service.getPluginRelease(ctx, pluginID, "v0.5.1")
	if err != nil {
		t.Fatalf("expected release lookup to succeed, got error: %v", err)
	}
	if release == nil {
		t.Fatalf("expected release row after enable")
	}

	snapshot, err := service.catalogSvc.ParseManifestSnapshot(release.ManifestSnapshot)
	if err != nil {
		t.Fatalf("expected manifest snapshot parse to succeed, got error: %v", err)
	}
	if snapshot == nil || !snapshot.HostServiceAuthConfirmed {
		t.Fatalf("expected confirmed host service authorization snapshot, got %#v", snapshot)
	}

	activeManifest, err := service.getActivePluginManifest(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected active manifest lookup to succeed, got error: %v", err)
	}
	if activeManifest == nil {
		t.Fatalf("expected active manifest after enable")
	}
	if len(activeManifest.HostServices) != 2 {
		t.Fatalf("expected active manifest to use narrowed host services, got %#v", activeManifest.HostServices)
	}
	if activeManifest.HostServices[0].Service != pluginbridge.HostServiceRuntime &&
		activeManifest.HostServices[1].Service != pluginbridge.HostServiceRuntime {
		t.Fatalf("expected runtime host service to remain authorized, got %#v", activeManifest.HostServices)
	}
	for _, spec := range activeManifest.HostServices {
		if spec == nil {
			continue
		}
		if spec.Service == pluginbridge.HostServiceNetwork {
			t.Fatalf("expected network host service to be removed from authorized snapshot, got %#v", activeManifest.HostServices)
		}
	}
	if _, ok := activeManifest.HostCapabilities[pluginbridge.CapabilityHTTPRequest]; ok {
		t.Fatalf("expected network capability to be removed with rejected authorization")
	}
}

func TestSourcePluginInstallAndUninstallRequireExplicitLifecycle(t *testing.T) {
	var (
		service = newTestService()
		ctx     = context.Background()
	)

	const (
		pluginID = "plugin-source-explicit-lifecycle"
		menuKey  = "plugin:plugin-source-explicit-lifecycle:entry"
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	testutil.WriteTestFile(
		t,
		manifestPath,
		"id: "+pluginID+"\n"+
			"name: Source Explicit Lifecycle Plugin\n"+
			"version: v0.1.0\n"+
			"type: source\n"+
			"menus:\n"+
			"  - key: "+menuKey+"\n"+
			"    name: Source Explicit Lifecycle Plugin\n"+
			"    path: plugin-source-explicit-lifecycle\n"+
			"    component: system/plugin/dynamic-page\n"+
			"    perms: plugin-source-explicit-lifecycle:view\n"+
			"    icon: ant-design:appstore-outlined\n"+
			"    type: M\n"+
			"    sort: -1\n",
	)

	testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := service.SyncAndList(ctx); err != nil {
		t.Fatalf("expected source plugin discovery to succeed, got error: %v", err)
	}

	registry, err := service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected source plugin registry lookup to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatalf("expected source plugin registry row to exist after discovery")
	}
	if registry.Installed != catalog.InstalledNo || registry.Status != catalog.StatusDisabled {
		t.Fatalf("expected source plugin to stay uninstalled+disabled after discovery, got installed=%d enabled=%d", registry.Installed, registry.Status)
	}

	menu, err := testutil.QueryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected source plugin menu query to succeed, got error: %v", err)
	}
	if menu != nil {
		t.Fatalf("expected source plugin menu to remain absent before install")
	}

	release, err := service.getPluginRelease(ctx, pluginID, "v0.1.0")
	if err != nil {
		t.Fatalf("expected source plugin release lookup after discovery to succeed, got error: %v", err)
	}
	if release == nil {
		t.Fatalf("expected source plugin release row after discovery")
	}
	if release.Status != catalog.ReleaseStatusUninstalled.String() {
		t.Fatalf("expected discovered source plugin release to stay uninstalled, got %s", release.Status)
	}

	if err = service.Install(ctx, pluginID, nil); err != nil {
		t.Fatalf("expected source plugin install to succeed, got error: %v", err)
	}

	registry, err = service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected source plugin registry lookup after install to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatalf("expected source plugin registry row after install")
	}
	if registry.Installed != catalog.InstalledYes || registry.Status != catalog.StatusDisabled {
		t.Fatalf("expected source plugin install to yield installed+disabled, got installed=%d enabled=%d", registry.Installed, registry.Status)
	}
	if registry.InstalledAt == nil {
		t.Fatalf("expected source plugin install to record installed_at")
	}

	menu, err = testutil.QueryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected source plugin menu query after install to succeed, got error: %v", err)
	}
	if menu == nil {
		t.Fatalf("expected source plugin menu to be created on install")
	}

	release, err = service.getPluginRelease(ctx, pluginID, "v0.1.0")
	if err != nil {
		t.Fatalf("expected source plugin release lookup after install to succeed, got error: %v", err)
	}
	if release == nil {
		t.Fatalf("expected source plugin release row after install")
	}
	if release.Status != catalog.ReleaseStatusInstalled.String() {
		t.Fatalf("expected source plugin release to become installed, got %s", release.Status)
	}

	migrationCount, err := dao.SysPluginMigration.Ctx(ctx).
		Where(do.SysPluginMigration{
			PluginId: pluginID,
			Phase:    catalog.MigrationDirectionInstall.String(),
		}).
		Count()
	if err != nil {
		t.Fatalf("expected source plugin install migration count query to succeed, got error: %v", err)
	}
	if migrationCount != 1 {
		t.Fatalf("expected one source plugin install migration row, got count=%d", migrationCount)
	}

	resourceCount, err := dao.SysPluginResourceRef.Ctx(ctx).
		Where(do.SysPluginResourceRef{PluginId: pluginID, ReleaseId: release.Id}).
		Count()
	if err != nil {
		t.Fatalf("expected source plugin resource ref count query to succeed, got error: %v", err)
	}
	if resourceCount == 0 {
		t.Fatalf("expected source plugin install to materialize governance resource refs")
	}

	if err = service.Uninstall(ctx, pluginID); err != nil {
		t.Fatalf("expected source plugin uninstall to succeed, got error: %v", err)
	}

	registry, err = service.getPluginRegistry(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected source plugin registry lookup after uninstall to succeed, got error: %v", err)
	}
	if registry == nil {
		t.Fatalf("expected source plugin registry row after uninstall")
	}
	if registry.Installed != catalog.InstalledNo || registry.Status != catalog.StatusDisabled {
		t.Fatalf("expected source plugin uninstall to yield uninstalled+disabled, got installed=%d enabled=%d", registry.Installed, registry.Status)
	}

	menu, err = testutil.QueryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected source plugin menu query after uninstall to succeed, got error: %v", err)
	}
	if menu != nil {
		t.Fatalf("expected source plugin menu to be removed on uninstall")
	}

	release, err = service.getPluginRelease(ctx, pluginID, "v0.1.0")
	if err != nil {
		t.Fatalf("expected source plugin release lookup after uninstall to succeed, got error: %v", err)
	}
	if release == nil {
		t.Fatalf("expected source plugin release row after uninstall")
	}
	if release.Status != catalog.ReleaseStatusUninstalled.String() {
		t.Fatalf("expected source plugin release to become uninstalled, got %s", release.Status)
	}

	resourceCount, err = dao.SysPluginResourceRef.Ctx(ctx).
		Where(do.SysPluginResourceRef{PluginId: pluginID, ReleaseId: release.Id}).
		Count()
	if err != nil {
		t.Fatalf("expected source plugin resource ref count query after uninstall to succeed, got error: %v", err)
	}
	if resourceCount != 0 {
		t.Fatalf("expected source plugin uninstall to clear governance resource refs, got count=%d", resourceCount)
	}

	migrationCount, err = dao.SysPluginMigration.Ctx(ctx).
		Where(do.SysPluginMigration{
			PluginId: pluginID,
			Phase:    catalog.MigrationDirectionUninstall.String(),
		}).
		Count()
	if err != nil {
		t.Fatalf("expected source plugin uninstall migration count query to succeed, got error: %v", err)
	}
	if migrationCount != 1 {
		t.Fatalf("expected one source plugin uninstall migration row, got count=%d", migrationCount)
	}
}
