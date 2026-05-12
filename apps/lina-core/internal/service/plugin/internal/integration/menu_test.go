// This file covers menu and permission-menu synchronization behaviors owned by integration.

package integration_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	menusvc "lina-core/internal/service/menu"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/internal/service/startupstats"
	"lina-core/pkg/pluginbridge"
)

// TestSyncSourcePluginMenusFromManifest verifies source plugin menus are only
// materialized on explicit sync and are deleted when removed from the manifest.
func TestSyncSourcePluginMenusFromManifest(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()
	adminRoleID := mustQueryAdminRoleID(t, ctx)

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
			"scope_nature: tenant_aware\n"+
			"supports_multi_tenant: false\n"+
			"default_install_mode: global\n"+
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
				Type:      catalog.MenuTypePage.String(),
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
		t.Fatalf("expected source plugin manifest sync to succeed, got error: %v", err)
	}

	menu, err := testutil.QueryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected plugin menu query to succeed, got error: %v", err)
	}
	if menu != nil {
		t.Fatalf("expected source plugin menu %s to stay absent before explicit install", menuKey)
	}

	if err := services.Integration.SyncPluginMenusAndPermissions(ctx, manifest); err != nil {
		t.Fatalf("expected explicit source plugin menu sync to succeed, got error: %v", err)
	}

	menu, err = testutil.QueryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected plugin menu query after explicit sync to succeed, got error: %v", err)
	}
	if menu == nil {
		t.Fatalf("expected source plugin menu %s to be created after explicit sync", menuKey)
	}
	if menu.Path != "plugin-source-menu-sync" {
		t.Fatalf("expected source plugin menu path to be synced, got %s", menu.Path)
	}

	roleMenuCount, err := dao.SysRoleMenu.Ctx(ctx).
		Where(do.SysRoleMenu{
			RoleId: adminRoleID,
			MenuId: menu.Id,
		}).
		Count()
	if err != nil {
		t.Fatalf("expected admin role binding query to succeed, got error: %v", err)
	}
	if roleMenuCount != 0 {
		t.Fatalf("expected source plugin menu not to be granted to admin role, got count=%d", roleMenuCount)
	}

	testutil.WriteTestFile(
		t,
		manifestPath,
		"id: "+pluginID+"\nname: Source Menu Sync Plugin\nversion: v0.1.0\ntype: source\nscope_nature: tenant_aware\nsupports_multi_tenant: false\ndefault_install_mode: global\n",
	)
	manifest.Menus = nil
	if err := services.Catalog.ValidateManifest(manifest, manifestPath); err != nil {
		t.Fatalf("expected source plugin manifest without menus to stay valid, got error: %v", err)
	}
	if err := services.Integration.SyncPluginMenusAndPermissions(ctx, manifest); err != nil {
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

// TestDynamicPluginInstallAndUninstallManageMenusFromManifest verifies dynamic
// plugin install/uninstall creates and removes manifest-owned menus.
func TestDynamicPluginInstallAndUninstallManageMenusFromManifest(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()
	adminRoleID := mustQueryAdminRoleID(t, ctx)

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
				Type:      catalog.MenuTypePage.String(),
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
			RoleId: adminRoleID,
			MenuId: menu.Id,
		}).
		Count()
	if err != nil {
		t.Fatalf("expected runtime admin role binding query to succeed, got error: %v", err)
	}
	if roleMenuCount != 0 {
		t.Fatalf("expected runtime plugin menu not to be granted to admin role, got count=%d", roleMenuCount)
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

// TestSyncPluginMenusAndPermissionsNoopSkipsWritesAndTransactions verifies a
// no-op startup sync performs no menu writes and opens no empty transaction.
func TestSyncPluginMenusAndPermissionsNoopSkipsWritesAndTransactions(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const (
		pluginID   = "plugin-menu-noop-startup"
		menuKey    = "plugin:plugin-menu-noop-startup:main-entry"
		permission = "plugin-menu-noop-startup:review:view"
	)

	manifest := &catalog.Manifest{
		ID:          pluginID,
		Name:        "Menu Noop Startup Plugin",
		Version:     "v0.1.0",
		Type:        catalog.TypeDynamic.String(),
		Description: "Menu no-op startup test plugin",
		Menus: []*catalog.MenuSpec{
			{
				Key:       menuKey,
				Name:      "Menu Noop Startup Plugin",
				Path:      "/plugin-assets/plugin-menu-noop-startup/v0.1.0/index.html",
				Perms:     "plugin-menu-noop-startup:view",
				Icon:      "ant-design:deployment-unit-outlined",
				Type:      catalog.MenuTypePage.String(),
				Sort:      -1,
				Component: "system/plugin/dynamic-page",
			},
		},
		Routes: []*pluginbridge.RouteContract{
			{
				Path:       "/review-summary",
				Method:     http.MethodGet,
				Access:     pluginbridge.AccessLogin,
				Permission: permission,
			},
		},
	}

	testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
	})

	if err := services.Integration.SyncPluginMenusAndPermissions(ctx, manifest); err != nil {
		t.Fatalf("expected initial menu sync to succeed, got error: %v", err)
	}

	collector := startupstats.New()
	startupCtx := startupstats.WithCollector(ctx, collector)
	startupCtx, err := services.Integration.WithStartupDataSnapshot(startupCtx)
	if err != nil {
		t.Fatalf("build integration startup snapshot: %v", err)
	}

	sqls, logs, err := captureSQLDuring(t, startupCtx, func(ctx context.Context) error {
		return services.Integration.SyncPluginMenusAndPermissions(ctx, manifest)
	})
	if err != nil {
		t.Fatalf("expected no-op menu sync to succeed, got error: %v", err)
	}
	assertNoMutationSQL(t, sqls)
	assertNoMutationSQL(t, logs)

	snapshot := collector.Snapshot()
	if got := snapshot.CounterValue(startupstats.CounterPluginMenuSyncNoop); got != 1 {
		t.Fatalf("expected one no-op menu sync, got %d", got)
	}
	if got := snapshot.CounterValue(startupstats.CounterPluginMenuSyncChanged); got != 0 {
		t.Fatalf("expected no changed menu sync, got %d", got)
	}
}

// mustQueryAdminRoleID resolves the built-in admin role ID for integration assertions.
func mustQueryAdminRoleID(t *testing.T, ctx context.Context) int {
	t.Helper()

	var adminRole *entity.SysRole
	err := dao.SysRole.Ctx(ctx).
		Where(do.SysRole{Key: "admin"}).
		Scan(&adminRole)
	if err != nil {
		t.Fatalf("expected admin role query to succeed, got error: %v", err)
	}
	if adminRole == nil {
		t.Fatal("expected built-in admin role to exist")
	}
	return adminRole.Id
}

// TestDynamicPluginRoutePermissionsMaterializeHiddenMenus verifies dynamic
// route permissions are projected as hidden button menus.
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
	if menu.Type != catalog.MenuTypeButton.String() || menu.Visible != 0 {
		t.Fatalf("expected synthetic permission menu to be hidden button, got %#v", menu)
	}
	if strings.ContainsAny(menu.Name, "动态路由权限") {
		t.Fatalf("expected synthetic permission menu source name to avoid localized CJK text, got %q", menu.Name)
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

// TestDynamicPluginRoutePermissionMenusAttachToPluginMenu verifies synthetic
// route-permission buttons are nested under the dynamic plugin's own menu.
func TestDynamicPluginRoutePermissionMenusAttachToPluginMenu(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const (
		pluginID   = "plugin-dynamic-route-permission-parent"
		menuKey    = "plugin:plugin-dynamic-route-permission-parent:main-entry"
		permission = "plugin-dynamic-route-permission-parent:review:view"
		version    = "v0.3.0"
	)

	menus := []*catalog.MenuSpec{
		{
			Key:       menuKey,
			Name:      "Runtime Route Permission Parent Plugin",
			Path:      "/plugin-assets/plugin-dynamic-route-permission-parent/v0.3.0/index.html",
			Perms:     "plugin-dynamic-route-permission-parent:view",
			Icon:      "ant-design:deployment-unit-outlined",
			Type:      catalog.MenuTypePage.String(),
			Sort:      -1,
			Component: "system/plugin/dynamic-page",
		},
	}
	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithMenus(
		t,
		pluginID,
		"Runtime Route Permission Parent Plugin",
		version,
		menus,
		nil,
		nil,
	)
	writeRuntimeArtifactWithMenusAndRoutePermissions(
		t,
		artifactPath,
		pluginID,
		"Runtime Route Permission Parent Plugin",
		version,
		menus,
		permission,
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

	parentMenu, err := testutil.QueryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected parent menu query to succeed, got error: %v", err)
	}
	if parentMenu == nil {
		t.Fatal("expected parent plugin menu to be created")
	}

	permissionMenuKey := integration.BuildDynamicRoutePermissionMenuKey(pluginID, permission)
	permissionMenu, err := testutil.QueryMenuByKey(ctx, permissionMenuKey)
	if err != nil {
		t.Fatalf("expected synthetic permission menu query to succeed, got error: %v", err)
	}
	if permissionMenu == nil {
		t.Fatal("expected synthetic permission menu to be created")
	}
	if permissionMenu.ParentId != parentMenu.Id {
		t.Fatalf("expected synthetic permission menu parent %d, got %d", parentMenu.Id, permissionMenu.ParentId)
	}
}

// TestDynamicPluginRoutePermissionMenusDeleteStaleEntriesOnRefresh verifies a
// same-version refresh cleans up superseded synthetic permission menus.
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

// TestDynamicPluginRoutePermissionRefreshIgnoresUnrelatedBrokenRegistry
// verifies target-plugin refresh still succeeds when another staged dynamic
// registry row is broken.
func TestDynamicPluginRoutePermissionRefreshIgnoresUnrelatedBrokenRegistry(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const (
		targetPluginID = "plugin-route-refresh-ok"
		brokenPluginID = "plugin-route-refresh-bad"
		permissionKey  = "plugin-route-refresh-ok:review:view"
		version        = "v0.3.0"
	)

	artifactPath := testutil.CreateTestRuntimeStorageArtifactWithMenus(
		t,
		targetPluginID,
		"Runtime Route Permission Refresh Target Plugin",
		version,
		nil,
		nil,
		nil,
	)
	writeRuntimeArtifactWithRoutePermissions(
		t,
		artifactPath,
		targetPluginID,
		"Runtime Route Permission Refresh Target Plugin",
		version,
		permissionKey,
	)

	testutil.CleanupPluginGovernanceRowsHard(t, ctx, targetPluginID)
	testutil.CleanupPluginMenuRowsHard(t, ctx, targetPluginID)
	testutil.CleanupPluginGovernanceRowsHard(t, ctx, brokenPluginID)
	testutil.CleanupPluginMenuRowsHard(t, ctx, brokenPluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginMenuRowsHard(t, ctx, brokenPluginID)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, brokenPluginID)
		testutil.CleanupPluginMenuRowsHard(t, ctx, targetPluginID)
		testutil.CleanupPluginGovernanceRowsHard(t, ctx, targetPluginID)
	})

	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected target runtime manifest to load, got error: %v", err)
	}
	if _, err = services.Catalog.SyncManifest(ctx, manifest); err != nil {
		t.Fatalf("expected target manifest sync to succeed, got error: %v", err)
	}

	_, err = dao.SysPlugin.Ctx(ctx).Data(do.SysPlugin{
		PluginId:     brokenPluginID,
		Name:         "Broken Dynamic Plugin",
		Version:      "v0.0.1",
		Type:         catalog.TypeDynamic.String(),
		Installed:    catalog.InstalledNo,
		Status:       catalog.StatusDisabled,
		DesiredState: catalog.HostStateInstalled.String(),
		CurrentState: catalog.HostStateReconciling.String(),
		Generation:   int64(1),
		Checksum:     "broken-dynamic-plugin-checksum",
	}).Insert()
	if err != nil {
		t.Fatalf("expected broken dynamic registry seed to succeed, got error: %v", err)
	}

	if err = services.Lifecycle.Install(ctx, targetPluginID); err != nil {
		t.Fatalf("expected target install to ignore unrelated broken registry, got error: %v", err)
	}

	menuKey := integration.BuildDynamicRoutePermissionMenuKey(targetPluginID, permissionKey)
	menu, err := testutil.QueryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected target synthetic permission menu query to succeed, got error: %v", err)
	}
	if menu == nil {
		t.Fatalf("expected target synthetic permission menu %s to exist", menuKey)
	}
}

// TestDynamicRoutePermissionMenuKeyAvoidsCollisions verifies the synthetic menu
// key builder preserves distinct permission identifiers.
func TestDynamicRoutePermissionMenuKeyAvoidsCollisions(t *testing.T) {
	const pluginID = "plugin-dynamic-route-key-collision"

	keyOne := integration.BuildDynamicRoutePermissionMenuKey(pluginID, "plugin-dynamic-route-key-collision:report/a:view")
	keyTwo := integration.BuildDynamicRoutePermissionMenuKey(pluginID, "plugin-dynamic-route-key-collision:report-a:view")
	if keyOne == keyTwo {
		t.Fatalf("expected distinct permission menu keys, got identical key %s", keyOne)
	}
}

// TestFilterMenusHidesRuntimeMenusWhenArtifactIsMissing verifies runtime menus
// disappear when the backing dynamic artifact has been removed.
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
			Type:    catalog.MenuTypePage.String(),
			Status:  1,
			Visible: 1,
		},
	})
	if len(filtered) != 0 {
		t.Fatalf("expected dynamic plugin menu to be hidden after artifact removal, got %d entries", len(filtered))
	}
}

// writeRuntimeArtifactWithRoutePermissions rewrites the test runtime artifact
// so it declares the provided backend route permissions.
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

// writeRuntimeArtifactWithMenusAndRoutePermissions rewrites the test runtime
// artifact with both menu declarations and backend route permissions.
func writeRuntimeArtifactWithMenusAndRoutePermissions(
	t *testing.T,
	artifactPath string,
	pluginID string,
	pluginName string,
	version string,
	menus []*catalog.MenuSpec,
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
			Menus:   menus,
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

// TestSyncPluginMenusResolvesStableHostParent verifies plugin menu sync maps a
// manifest parent_key that targets one host catalog into the persisted parent_id.
func TestSyncPluginMenusResolvesStableHostParent(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const (
		pluginID = "plugin-stable-parent-sync"
		menuKey  = "plugin:plugin-stable-parent-sync:main-entry"
	)

	testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
	t.Cleanup(func() {
		testutil.CleanupPluginMenuRowsHard(t, ctx, pluginID)
	})

	hostParent, err := testutil.QueryMenuByKey(ctx, menusvc.Monitor)
	if err != nil {
		t.Fatalf("expected host parent query to succeed, got error: %v", err)
	}
	if hostParent == nil {
		t.Fatalf("expected host stable parent menu %s to exist", menusvc.Monitor)
	}

	manifest := &catalog.Manifest{
		ID:      pluginID,
		Name:    "Stable Parent Sync Plugin",
		Version: "0.1.0",
		Type:    catalog.TypeSource.String(),
		Menus: []*catalog.MenuSpec{
			{
				Key:       menuKey,
				Name:      "Stable Parent Sync Plugin",
				ParentKey: menusvc.Monitor,
				Path:      "plugin-stable-parent-sync",
				Component: "system/plugin/dynamic-page",
				Type:      catalog.MenuTypePage.String(),
			},
		},
	}

	if err = services.Integration.SyncPluginMenusAndPermissions(ctx, manifest); err != nil {
		t.Fatalf("expected plugin menu sync to succeed, got error: %v", err)
	}

	menu, err := testutil.QueryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected synced plugin menu query to succeed, got error: %v", err)
	}
	if menu == nil {
		t.Fatalf("expected plugin menu %s to be created", menuKey)
	}
	if menu.ParentId != hostParent.Id {
		t.Fatalf("expected plugin menu parent_id=%d, got %d", hostParent.Id, menu.ParentId)
	}
}

// TestSyncMultiTenantPluginMenusResolveAllowedHostParents verifies the
// multi-tenant plugin mounts only under the platform host catalog.
func TestSyncMultiTenantPluginMenusResolveAllowedHostParents(t *testing.T) {
	services := testutil.NewServices()
	ctx := context.Background()

	const (
		platformMenuKey = "plugin:multi-tenant:platform:tenants"
	)

	testutil.CleanupPluginMenuRowsHard(t, ctx, menusvc.MultiTenant)
	t.Cleanup(func() {
		testutil.CleanupPluginMenuRowsHard(t, ctx, menusvc.MultiTenant)
	})

	platformParent := ensureTestStableHostMenu(t, ctx, menusvc.Platform)

	manifest := &catalog.Manifest{
		ID:      menusvc.MultiTenant,
		Name:    "Multi Tenant",
		Version: "0.1.0",
		Type:    catalog.TypeSource.String(),
		Menus: []*catalog.MenuSpec{
			{
				Key:       platformMenuKey,
				Name:      "Tenant Management",
				ParentKey: menusvc.Platform,
				Path:      "platform/tenant",
				Component: "system/plugin/dynamic-page",
				Type:      catalog.MenuTypePage.String(),
			},
		},
	}

	if err := services.Integration.SyncPluginMenusAndPermissions(ctx, manifest); err != nil {
		t.Fatalf("expected multi-tenant plugin menu sync to succeed, got error: %v", err)
	}

	platformMenu, err := testutil.QueryMenuByKey(ctx, platformMenuKey)
	if err != nil {
		t.Fatalf("expected platform plugin menu query to succeed, got error: %v", err)
	}
	if platformMenu == nil {
		t.Fatalf("expected plugin menu %s to be created", platformMenuKey)
	}
	if platformMenu.ParentId != platformParent.Id {
		t.Fatalf("expected platform plugin menu parent_id=%d, got %d", platformParent.Id, platformMenu.ParentId)
	}
}

// ensureTestStableHostMenu ensures a stable host menu exists for integration
// tests running against databases initialized before the current iteration.
func ensureTestStableHostMenu(t *testing.T, ctx context.Context, menuKey string) *entity.SysMenu {
	t.Helper()

	existing, err := testutil.QueryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected host parent query to succeed, got error: %v", err)
	}
	if existing != nil {
		return existing
	}

	menuID, err := dao.SysMenu.Ctx(ctx).Data(do.SysMenu{
		ParentId:   0,
		MenuKey:    menuKey,
		Name:       menuKey,
		Path:       menuKey,
		Type:       catalog.MenuTypeDirectory.String(),
		Sort:       100,
		Visible:    1,
		Status:     1,
		IsFrame:    0,
		IsCache:    0,
		QueryParam: "",
		Remark:     "integration test stable host menu",
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("expected host parent insert to succeed, got error: %v", err)
	}
	t.Cleanup(func() {
		if _, err := dao.SysMenu.Ctx(ctx).Unscoped().Where(do.SysMenu{Id: int(menuID)}).Delete(); err != nil {
			t.Fatalf("expected host parent cleanup to succeed, got error: %v", err)
		}
	})

	created, err := testutil.QueryMenuByKey(ctx, menuKey)
	if err != nil {
		t.Fatalf("expected host parent query after insert to succeed, got error: %v", err)
	}
	if created == nil {
		t.Fatalf("expected host stable parent menu %s to exist", menuKey)
	}
	return created
}
