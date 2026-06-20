// This file verifies role menu assignment filtering and write-time rejection
// for platform-only permissions in tenant contexts.

package role

import (
	"context"
	"testing"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/menutype"
)

// TestTenantRoleRejectsPlatformOnlyMenuAssignment verifies role creation fails
// before any role or role-menu row is written when tenant context submits a
// platform-control-plane menu ID.
func TestTenantRoleRejectsPlatformOnlyMenuAssignment(t *testing.T) {
	ctx := datascope.WithTenantForTest(context.Background(), 62101)
	svc := newDefaultRoleTestService()
	setRoleTestBizCtx(svc, roleScopeStaticBizCtx{ctx: &model.Context{TenantId: 62101}})
	enableRoleAssignableTenantCapability(t, svc)
	platformMenuID := insertRoleAssignableMenu(t, ctx, roleAssignableMenuSeed{
		Label:   "tenant-deny-plugin-install",
		MenuKey: uniqueRoleTenantBoundaryName("extension-plugin-install"),
		Perms:   "plugin:install",
		Type:    menutype.Button.String(),
	})
	t.Cleanup(func() {
		cleanupRoleTestRows(t, ctx, nil, nil, []int{platformMenuID})
	})

	roleID, err := svc.Create(ctx, CreateInput{
		Name:      uniqueRoleTenantBoundaryName("tenant-deny-platform-menu"),
		Key:       uniqueRoleTenantBoundaryName("tenant-deny-platform-menu-key"),
		Sort:      1,
		DataScope: roleDataScopeSelf,
		Status:    roleTenantBoundaryStatusNormal,
		MenuIds:   []int{platformMenuID},
	})

	if !bizerr.Is(err, CodeRoleMenuAssignmentForbidden) {
		t.Fatalf("expected role menu assignment denial, got role=%d err=%v", roleID, err)
	}
	if roleID != 0 {
		t.Fatalf("expected no role ID on forbidden assignment, got %d", roleID)
	}
	if count := mustCountRoleTenantBoundaryRoleMenu(t, ctx, roleID, 62101); count != 0 {
		t.Fatalf("expected no forbidden role-menu row, got %d", count)
	}
}

// TestTenantRoleUpdateRejectsPlatformOnlyMenuAssignment verifies update keeps
// the existing role-menu grants when a tenant submits platform-only menu IDs.
func TestTenantRoleUpdateRejectsPlatformOnlyMenuAssignment(t *testing.T) {
	ctx := datascope.WithTenantForTest(context.Background(), 62102)
	svc := newDefaultRoleTestService()
	setRoleTestBizCtx(svc, roleScopeStaticBizCtx{ctx: &model.Context{TenantId: 62102}})
	enableRoleAssignableTenantCapability(t, svc)
	roleID := insertRoleTenantBoundaryRole(t, ctx, "tenant-update-deny", 62102)
	allowedMenuID := insertRoleAssignableMenu(t, ctx, roleAssignableMenuSeed{
		Label:   "tenant-allowed-plugin-list",
		MenuKey: uniqueRoleAssignablePluginMenuKey("linapro-tenant-core", "tenant-plugin-list"),
		Perms:   "system:tenant:plugin:list",
		Type:    menutype.Button.String(),
	})
	platformMenuID := insertRoleAssignableMenu(t, ctx, roleAssignableMenuSeed{
		Label:   "tenant-deny-menu-edit",
		MenuKey: uniqueRoleTenantBoundaryName("system-menu-edit"),
		Perms:   "system:menu:edit",
		Type:    menutype.Button.String(),
	})
	t.Cleanup(func() {
		cleanupRoleTestRows(t, ctx, []int{roleID}, nil, []int{allowedMenuID, platformMenuID})
	})
	insertRoleTenantBoundaryRoleMenu(t, ctx, roleID, allowedMenuID, 62102)
	sortValue := 2
	dataScope := roleDataScopeSelf
	statusValue := roleTenantBoundaryStatusNormal
	remark := "deny platform menu update"

	err := svc.Update(ctx, UpdateInput{
		Id:        roleID,
		Name:      uniqueRoleTenantBoundaryName("tenant-update-deny-new"),
		Key:       uniqueRoleTenantBoundaryName("tenant-update-deny-new-key"),
		Sort:      &sortValue,
		DataScope: &dataScope,
		Status:    &statusValue,
		Remark:    &remark,
		MenuIds:   []int{allowedMenuID, platformMenuID},
	})

	if !bizerr.Is(err, CodeRoleMenuAssignmentForbidden) {
		t.Fatalf("expected role menu assignment denial, got %v", err)
	}
	if count := mustCountRoleTenantBoundaryRoleMenu(t, ctx, roleID, 62102); count != 1 {
		t.Fatalf("expected existing tenant role-menu row to remain, got %d", count)
	}
}

// TestTenantAssignableMenusFiltersPlatformOnlyNodes verifies tenant contexts
// remove platform tenant-management, platform plugin governance, global menu
// governance, and platform-only plugin menus while preserving tenant plugin
// self-service permissions.
func TestTenantAssignableMenusFiltersPlatformOnlyNodes(t *testing.T) {
	ctx := datascope.WithTenantForTest(context.Background(), 62103)
	svc := newDefaultRoleTestService()
	setRoleTestBizCtx(svc, roleScopeStaticBizCtx{ctx: &model.Context{TenantId: 62103}})
	enableRoleAssignableTenantCapability(t, svc)
	pluginID := uniqueRoleTenantBoundaryName("platform-only-plugin")
	insertRoleAssignablePlugin(t, ctx, pluginID, rolePluginScopePlatformOnly)
	tenantPlatformID := insertRoleAssignableMenu(t, ctx, roleAssignableMenuSeed{
		Label:   "tenant-platform",
		MenuKey: uniqueRoleAssignablePluginMenuKey("linapro-tenant-core", "platform-tenants"),
		Perms:   "system:tenant:list",
		Type:    menutype.Menu.String(),
	})
	tenantSelfServiceID := insertRoleAssignableMenu(t, ctx, roleAssignableMenuSeed{
		Label:   "tenant-self-service",
		MenuKey: uniqueRoleAssignablePluginMenuKey("linapro-tenant-core", "tenant-plugin-list"),
		Perms:   "system:tenant:plugin:list",
		Type:    menutype.Button.String(),
	})
	menuWriteID := insertRoleAssignableMenu(t, ctx, roleAssignableMenuSeed{
		Label:   "menu-write",
		MenuKey: uniqueRoleTenantBoundaryName("system-menu-add"),
		Perms:   "system:menu:add",
		Type:    menutype.Button.String(),
	})
	pluginGovernanceID := insertRoleAssignableMenu(t, ctx, roleAssignableMenuSeed{
		Label:   "plugin-governance",
		MenuKey: uniqueRoleTenantBoundaryName("extension-plugin-enable"),
		Perms:   "plugin:enable",
		Type:    menutype.Button.String(),
	})
	platformPluginMenuID := insertRoleAssignableMenu(t, ctx, roleAssignableMenuSeed{
		Label:   "platform-plugin-menu",
		MenuKey: uniqueRoleAssignablePluginMenuKey(pluginID, "entry"),
		Perms:   "platform-plugin:view",
		Type:    menutype.Menu.String(),
	})
	t.Cleanup(func() {
		cleanupRoleAssignablePlugin(t, ctx, pluginID)
		cleanupRoleTestRows(t, ctx, nil, nil, []int{
			tenantPlatformID,
			tenantSelfServiceID,
			menuWriteID,
			pluginGovernanceID,
			platformPluginMenuID,
		})
	})

	filtered, err := svc.FilterAssignableMenuIDs(ctx, []int{
		tenantPlatformID,
		tenantSelfServiceID,
		menuWriteID,
		pluginGovernanceID,
		platformPluginMenuID,
	})
	if err != nil {
		t.Fatalf("filter tenant assignable menu IDs: %v", err)
	}

	if len(filtered) != 1 || filtered[0] != tenantSelfServiceID {
		t.Fatalf("expected only tenant self-service menu to remain, got %#v", filtered)
	}
}

// TestPlatformContextKeepsPlatformOnlyMenuAssignment verifies platform context
// can still assign platform governance permissions.
func TestPlatformContextKeepsPlatformOnlyMenuAssignment(t *testing.T) {
	ctx := datascope.WithTenantForTest(context.Background(), datascope.PlatformTenantID)
	svc := newDefaultRoleTestService()
	setRoleTestBizCtx(svc, roleScopeStaticBizCtx{ctx: &model.Context{
		TenantId:  datascope.PlatformTenantID,
		DataScope: roleDataScopeAll,
	}})
	enableRoleAssignableTenantCapability(t, svc)
	platformMenuID := insertRoleAssignableMenu(t, ctx, roleAssignableMenuSeed{
		Label:   "platform-allow-plugin-install",
		MenuKey: uniqueRoleTenantBoundaryName("extension-plugin-install"),
		Perms:   "plugin:install",
		Type:    menutype.Button.String(),
	})
	t.Cleanup(func() {
		cleanupRoleTestRows(t, ctx, nil, nil, []int{platformMenuID})
	})

	if err := svc.EnsureAssignableMenuIDs(ctx, []int{platformMenuID}); err != nil {
		t.Fatalf("expected platform context to allow platform menu assignment, got %v", err)
	}
}

// TestRoleMenuPluginIDUsesMenuKeyOnly verifies plugin ownership classification
// does not read marker-like text from menu remarks.
func TestRoleMenuPluginIDUsesMenuKeyOnly(t *testing.T) {
	t.Parallel()

	if got := roleMenuPluginID(&entity.SysMenu{Remark: "plugin:remark-owner:entry"}); got != "" {
		t.Fatalf("expected remark-only plugin marker to be ignored, got %q", got)
	}
	if got := roleMenuPluginID(&entity.SysMenu{
		MenuKey: "plugin:current-owner:entry",
		Remark:  "plugin:remark-owner:entry",
	}); got != "current-owner" {
		t.Fatalf("expected menu_key plugin owner, got %q", got)
	}
}

// roleAssignableMenuSeed carries the menu fields relevant to assignability
// tests.
type roleAssignableMenuSeed struct {
	Label    string
	ParentID int
	MenuKey  string
	Perms    string
	Type     string
}

// insertRoleAssignableMenu inserts one menu row for assignability tests.
func insertRoleAssignableMenu(t *testing.T, ctx context.Context, seed roleAssignableMenuSeed) int {
	t.Helper()
	menuType := seed.Type
	if menuType == "" {
		menuType = menutype.Menu.String()
	}
	id, err := dao.SysMenu.Ctx(ctx).Data(do.SysMenu{
		ParentId: seed.ParentID,
		MenuKey:  seed.MenuKey,
		Name:     uniqueRoleTenantBoundaryName(seed.Label),
		Perms:    seed.Perms,
		Type:     menuType,
		Sort:     99,
		Visible:  roleTenantBoundaryStatusNormal,
		Status:   roleTenantBoundaryStatusNormal,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert role assignable menu: %v", err)
	}
	return int(id)
}

// insertRoleAssignablePlugin inserts plugin scope metadata for plugin-owned
// menu classification.
func insertRoleAssignablePlugin(t *testing.T, ctx context.Context, pluginID string, scopeNature string) {
	t.Helper()
	_, err := dao.SysPlugin.Ctx(ctx).Data(do.SysPlugin{
		PluginId:    pluginID,
		Name:        pluginID,
		Version:     "v0.0.1",
		Type:        "source",
		ScopeNature: scopeNature,
		InstallMode: "global",
		Installed:   1,
		Status:      1,
	}).Insert()
	if err != nil {
		t.Fatalf("insert role assignable plugin: %v", err)
	}
}

// cleanupRoleAssignablePlugin removes plugin rows created by assignability
// tests.
func cleanupRoleAssignablePlugin(t *testing.T, ctx context.Context, pluginID string) {
	t.Helper()
	if _, err := dao.SysPlugin.Ctx(ctx).Unscoped().Where(dao.SysPlugin.Columns().PluginId, pluginID).Delete(); err != nil {
		t.Fatalf("cleanup role assignable plugin: %v", err)
	}
}

// enableRoleAssignableTenantCapability activates the real tenantcap service so
// role assignability tests exercise tenant/platform context decisions instead
// of the single-tenant branch.
func enableRoleAssignableTenantCapability(t *testing.T, svc *serviceImpl) {
	t.Helper()
	svc.tenantSvc = activateRoleTenantBoundaryProvider(t)
}

// uniqueRoleAssignablePluginMenuKey builds a unique plugin-owned menu key while
// keeping the plugin ID parseable by the assignability classifier.
func uniqueRoleAssignablePluginMenuKey(pluginID string, suffix string) string {
	return "plugin:" + pluginID + ":" + uniqueRoleTenantBoundaryName(suffix)
}
