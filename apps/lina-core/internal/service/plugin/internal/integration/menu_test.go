// This file covers menu and permission-menu synchronization behaviors owned by integration.

package integration_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/pluginbridge"
)

const testDefaultAdminRoleID = 1

func TestSyncSourcePluginMenusFromManifest(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const (
		pluginID = "plugin-source-menu-sync"
		menuKey  = "plugin:plugin-source-menu-sync:sidebar-entry"
	)

	pluginDir := testutil.CreateTestPluginDir(t, pluginID)
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	testutil.WriteTestFile(
		t,
		manifestPath,
		"id: "+pluginID+"\n"+
			"name: Source Menu Sync Plugin\n"+
			"version: v0.1.0\n"+
			"type: source\n"+
			"menus:\n"+
			"  - key: "+menuKey+"\n"+
			"    name: Source Menu Sync Plugin\n"+
			"    path: plugin-source-menu-sync\n"+
			"    component: system/plugin/dynamic-page\n"+
			"    perms: plugin-source-menu-sync:view\n"+
			"    icon: ant-design:appstore-outlined\n"+
			"    type: M\n"+
			"    sort: -1\n",
	)

	manifest := &catalog.Manifest{
		ID:      pluginID,
		Name:    "Source Menu Sync Plugin",
		Version: "v0.1.0",
		Type:    catalog.TypeSource.String(),
		Menus: []*catalog.MenuSpec{
			{
				Key:       menuKey,
				Name:      "Source Menu Sync Plugin",
				Path:      "plugin-source-menu-sync",
				Component: "system/plugin/dynamic-page",
				Perms:     "plugin-source-menu-sync:view",
				Icon:      "ant-design:appstore-outlined",
				Type:      "M",
				Sort:      -1,
			},
		},
		ManifestPath: manifestPath,
		RootDir:      pluginDir,
	}
	if err := services.Catalog.ValidateManifest(manifest, manifestPath); err != nil {
		t.Fatalf("expected source plugin manifest with menus to be valid, got error: %v", err)
	}

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err := services.Catalog.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected source plugin menu sync to succeed, got error: %v", err)
	}

	menu, err := testutil.QueryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected plugin menu query to succeed, got error: %v", err)
	}
	if menu == nil {
		t.Fatalf("expected source plugin menu %s to be created", menuKey)
	}
	if menu.Path != "plugin-source-menu-sync" {
		t.Fatalf("expected source plugin menu path to be synced, got %s", menu.Path)
	}

	roleMenuCount, err := dao.SysRoleMenu.Ctx(ctx).
		Where(do.SysRoleMenu{
			RoleId: testDefaultAdminRoleID,
			MenuId: menu.Id,
		}).
		Count()
	if err != nil {
		t.Fatalf("expected admin role binding query to succeed, got error: %v", err)
	}
	if roleMenuCount != 1 {
		t.Fatalf("expected source plugin menu to be granted to admin role, got count=%d", roleMenuCount)
	}

	testutil.WriteTestFile(
		t,
		manifestPath,
		"id: "+pluginID+"\nname: Source Menu Sync Plugin\nversion: v0.1.0\ntype: source\n",
	)
	manifest.Menus = nil
	if err := services.Catalog.ValidateManifest(manifest, manifestPath); err != nil {
		t.Fatalf("expected source plugin manifest without menus to stay valid, got error: %v", err)
	}
	if _, err := services.Catalog.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected source plugin stale menu cleanup to succeed, got error: %v", err)
	}

	menu, err = testutil.QueryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected plugin menu cleanup query to succeed, got error: %v", err)
	}
	if menu != nil {
		t.Fatalf("expected source plugin menu %s to be deleted after manifest removed it", menuKey)
	}
}

func TestDynamicPluginInstallAndUninstallManageMenusFromManifest(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const (
		pluginID = "plugin-dynamic-menu-metadata"
		menuKey  = "plugin:plugin-dynamic-menu-metadata:main-entry"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithMenus(
		t,
		pluginID,
		"Runtime Menu Metadata Plugin",
		"v0.3.0",
		[]*catalog.MenuSpec{
			{
				Key:       menuKey,
				Name:      "Runtime Menu Metadata Plugin",
				Path:      "/plugin-assets/plugin-dynamic-menu-metadata/v0.3.0/index.html",
				Perms:     "plugin-dynamic-menu-metadata:view",
				Icon:      "ant-design:deployment-unit-outlined",
				Type:      "M",
				Sort:      -1,
				Query:     map[string]interface{}{"pluginAccessMode": "embedded-mount"},
				Component: "system/plugin/dynamic-page",
			},
		},
		nil,
		nil,
	)

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic artifact with manifest menus to load, got error: %v", err)
	}

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err = services.Catalog.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected runtime plugin manifest sync to succeed, got error: %v", err)
	}
	if err = services.Lifecycle.Install(ctx, pluginID); err != nil {
		t.Fatalf("expected runtime plugin install to succeed, got error: %v", err)
	}

	menu, err := testutil.QueryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected runtime plugin menu query to succeed, got error: %v", err)
	}
	if menu == nil {
		t.Fatalf("expected runtime plugin menu %s to be created on install", menuKey)
	}

	roleMenuCount, err := dao.SysRoleMenu.Ctx(ctx).
		Where(do.SysRoleMenu{
			RoleId: testDefaultAdminRoleID,
			MenuId: menu.Id,
		}).
		Count()
	if err != nil {
		t.Fatalf("expected runtime admin role binding query to succeed, got error: %v", err)
	}
	if roleMenuCount != 1 {
		t.Fatalf("expected runtime plugin menu to be granted to admin role, got count=%d", roleMenuCount)
	}

	if err = services.Lifecycle.Uninstall(ctx, pluginID); err != nil {
		t.Fatalf("expected runtime plugin uninstall to succeed, got error: %v", err)
	}

	menu, err = testutil.QueryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected runtime plugin menu cleanup query to succeed, got error: %v", err)
	}
	if menu != nil {
		t.Fatalf("expected runtime plugin menu %s to be deleted on uninstall", menuKey)
	}
}

func TestDynamicPluginRoutePermissionsMaterializeHiddenMenus(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const pluginID = "plugin-dynamic-route-permission"

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithMenus(
		t,
		pluginID,
		"Runtime Route Permission Plugin",
		"v0.3.0",
		nil,
		nil,
		nil,
	)
	writeRuntimeArtifactWithRoutePermissions(
		t,
		artifactPath,
		pluginID,
		"Runtime Route Permission Plugin",
		"v0.3.0",
		"plugin-dynamic-route-permission:review:view",
	)

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic runtime manifest to load, got error: %v", err)
	}

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	if _, err = services.Catalog.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected runtime plugin manifest sync to succeed, got error: %v", err)
	}
	if err = services.Lifecycle.Install(ctx, pluginID); err != nil {
		t.Fatalf("expected runtime plugin install to succeed, got error: %v", err)
	}

	menuKey := integration.BuildDynamicRoutePermissionMenuKey(
		pluginID,
		"plugin-dynamic-route-permission:review:view",
	)
	menu, err := testutil.QueryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected synthetic permission menu query to succeed, got error: %v", err)
	}
	if menu == nil {
		t.Fatal("expected synthetic permission menu to be created")
	}
	if menu.Type != "B" || menu.Visible != 0 {
		t.Fatalf("expected synthetic permission menu to be hidden button, got %#v", menu)
	}

	if err = services.Lifecycle.Uninstall(ctx, pluginID); err != nil {
		t.Fatalf("expected runtime plugin uninstall to succeed, got error: %v", err)
	}

	menu, err = testutil.QueryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected synthetic permission menu cleanup query to succeed, got error: %v", err)
	}
	if menu != nil {
		t.Fatal("expected synthetic permission menu to be deleted on uninstall")
	}
}

func TestDynamicPluginRoutePermissionMenusDeleteStaleEntriesOnRefresh(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const (
		pluginID      = "plugin-dynamic-route-permission-refresh"
		permissionOne = "plugin-dynamic-route-permission-refresh:review/view:read"
		permissionTwo = "plugin-dynamic-route-permission-refresh:review-view:read"
		version       = "v0.3.0"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithMenus(
		t,
		pluginID,
		"Runtime Route Permission Refresh Plugin",
		version,
		nil,
		nil,
		nil,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	writeRuntimeArtifactWithRoutePermissions(
		t,
		artifactPath,
		pluginID,
		"Runtime Route Permission Refresh Plugin",
		version,
		permissionOne,
	)

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected initial runtime manifest to load, got error: %v", err)
	}
	if _, err = services.Catalog.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected initial manifest sync to succeed, got error: %v", err)
	}
	if err = services.Lifecycle.Install(ctx, pluginID); err != nil {
		t.Fatalf("expected initial install to succeed, got error: %v", err)
	}

	menuOneKey := integration.BuildDynamicRoutePermissionMenuKey(pluginID, permissionOne)
	menuTwoKey := integration.BuildDynamicRoutePermissionMenuKey(pluginID, permissionTwo)

	menuOne, err := testutil.QueryMenuByKey(ctx, menuOneKey)
	if err != nil {
		t.Fatalf("expected initial synthetic permission menu query to succeed, got error: %v", err)
	}
	if menuOne == nil {
		t.Fatal("expected initial synthetic permission menu to exist")
	}

	writeRuntimeArtifactWithRoutePermissions(
		t,
		artifactPath,
		pluginID,
		"Runtime Route Permission Refresh Plugin",
		version,
		permissionTwo,
	)

	if err = services.Lifecycle.Install(ctx, pluginID); err != nil {
		t.Fatalf("expected refresh install to succeed, got error: %v", err)
	}

	menuOne, err = testutil.QueryMenuByKey(ctx, menuOneKey)
	if err != nil {
		t.Fatalf("expected stale synthetic permission menu query to succeed, got error: %v", err)
	}
	if menuOne != nil {
		t.Fatalf("expected stale synthetic permission menu %s to be deleted on refresh", menuOneKey)
	}

	menuTwo, err := testutil.QueryMenuByKey(ctx, menuTwoKey)
	if err != nil {
		t.Fatalf("expected refreshed synthetic permission menu query to succeed, got error: %v", err)
	}
	if menuTwo == nil {
		t.Fatalf("expected refreshed synthetic permission menu %s to exist", menuTwoKey)
	}
}

func TestDynamicRoutePermissionMenuKeyAvoidsCollisions(t *testing.T) {
	const pluginID = "plugin-dynamic-route-key-collision"

	keyOne := integration.BuildDynamicRoutePermissionMenuKey(pluginID, "plugin-dynamic-route-key-collision:report/a:view")
	keyTwo := integration.BuildDynamicRoutePermissionMenuKey(pluginID, "plugin-dynamic-route-key-collision:report-a:view")
	if keyOne == keyTwo {
		t.Fatalf("expected distinct permission menu keys, got identical key %s", keyOne)
	}
}

func TestFilterMenusHidesRuntimeMenusWhenArtifactIsMissing(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const pluginID = "plugin-dynamic-menu-hidden"

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithMenus(
		t,
		pluginID,
		"Runtime Menu Hidden Plugin",
		"v0.9.5",
		nil,
		nil,
		nil,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, pluginID)
	})

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected dynamic artifact manifest to load, got error: %v", err)
	}
	if _, err = services.Catalog.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected dynamic manifest sync to succeed, got error: %v", err)
	}
	if err = services.Catalog.SetPluginInstalled(ctx, pluginID, catalog.InstalledYes); err != nil {
		t.Fatalf("expected dynamic plugin install state to be set, got error: %v", err)
	}
	if err = services.Catalog.SetPluginStatus(ctx, pluginID, catalog.StatusEnabled); err != nil {
		t.Fatalf("expected dynamic plugin enable state to be set, got error: %v", err)
	}
	if err = os.Remove(artifactPath); err != nil {
		t.Fatalf("failed to remove dynamic artifact: %v", err)
	}

	filtered := services.Integration.FilterMenus(ctx, []*entity.SysMenu{
		{
			Id:      1,
			MenuKey: "plugin:" + pluginID + ":entry",
			Name:    "runtime menu",
			Type:    "M",
			Status:  1,
			Visible: 1,
		},
	})
	if len(filtered) != 0 {
		t.Fatalf("expected dynamic plugin menu to be hidden after artifact removal, got %d entries", len(filtered))
	}
}

func writeRuntimeArtifactWithRoutePermissions(
	t *testing.T,
	artifactPath string,
	pluginID string,
	pluginName string,
	version string,
	permissions ...string,
) {
	t.Helper()

	routes := make([]*pluginbridge.RouteContract, 0, len(permissions))
	for _, permission := range permissions {
		routes = append(routes, &pluginbridge.RouteContract{
			Path:       "/review-summary",
			Method:     http.MethodGet,
			Access:     pluginbridge.AccessLogin,
			Permission: permission,
		})
	}

	testutil.WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      pluginID,
			Name:    pluginName,
			Version: version,
			Type:    catalog.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:        pluginbridge.RuntimeKindWasm,
			ABIVersion:         pluginbridge.SupportedABIVersion,
			FrontendAssetCount: len(testutil.DefaultTestRuntimeFrontendAssets()),
			RouteCount:         len(routes),
		},
		testutil.DefaultTestRuntimeFrontendAssets(),
		nil,
		nil,
		routes,
		&pluginbridge.BridgeSpec{
			ABIVersion:     pluginbridge.ABIVersionV1,
			RuntimeKind:    pluginbridge.RuntimeKindWasm,
			RouteExecution: true,
			RequestCodec:   pluginbridge.CodecProtobuf,
			ResponseCodec:  pluginbridge.CodecProtobuf,
		},
	)
}
