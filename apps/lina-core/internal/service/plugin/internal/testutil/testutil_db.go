// This file contains database cleanup and lookup helpers for plugin tests.

package testutil

import (
	"context"
	"testing"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// CleanupPluginGovernanceRowsHard removes plugin governance records created during tests.
func CleanupPluginGovernanceRowsHard(t *testing.T, ctx context.Context, pluginID string) {
	t.Helper()

	if _, err := dao.SysPluginNodeState.Ctx(ctx).
		Unscoped().
		Where(do.SysPluginNodeState{PluginId: pluginID}).
		Delete(); err != nil {
		t.Fatalf("failed to delete sys_plugin_node_state rows for %s: %v", pluginID, err)
	}
	if _, err := dao.SysPluginResourceRef.Ctx(ctx).
		Unscoped().
		Where(do.SysPluginResourceRef{PluginId: pluginID}).
		Delete(); err != nil {
		t.Fatalf("failed to delete sys_plugin_resource_ref rows for %s: %v", pluginID, err)
	}
	if _, err := dao.SysPluginMigration.Ctx(ctx).
		Unscoped().
		Where(do.SysPluginMigration{PluginId: pluginID}).
		Delete(); err != nil {
		t.Fatalf("failed to delete sys_plugin_migration rows for %s: %v", pluginID, err)
	}
	if _, err := dao.SysPluginRelease.Ctx(ctx).
		Unscoped().
		Where(do.SysPluginRelease{PluginId: pluginID}).
		Delete(); err != nil {
		t.Fatalf("failed to delete sys_plugin_release rows for %s: %v", pluginID, err)
	}
	if _, err := dao.SysPlugin.Ctx(ctx).
		Unscoped().
		Where(do.SysPlugin{PluginId: pluginID}).
		Delete(); err != nil {
		t.Fatalf("failed to delete sys_plugin rows for %s: %v", pluginID, err)
	}
}

// CleanupPluginMenuRowsHard removes plugin-owned menu rows and admin bindings created during tests.
func CleanupPluginMenuRowsHard(t *testing.T, ctx context.Context, pluginID string) {
	t.Helper()

	services := NewServices()
	menus, err := services.Integration.ListPluginMenusByPlugin(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected plugin menu cleanup query to succeed, got error: %v", err)
	}
	if len(menus) == 0 {
		return
	}

	menuIDs := make([]interface{}, 0, len(menus))
	menuKeys := make([]string, 0, len(menus))
	for _, item := range menus {
		if item == nil {
			continue
		}
		menuIDs = append(menuIDs, item.Id)
		menuKeys = append(menuKeys, item.MenuKey)
	}

	if len(menuIDs) > 0 {
		if _, err := dao.SysRoleMenu.Ctx(ctx).
			WhereIn(dao.SysRoleMenu.Columns().MenuId, menuIDs).
			Delete(); err != nil {
			t.Fatalf("failed to delete sys_role_menu rows for %s: %v", pluginID, err)
		}
	}
	if len(menuKeys) > 0 {
		if _, err := dao.SysMenu.Ctx(ctx).
			Unscoped().
			WhereIn(dao.SysMenu.Columns().MenuKey, menuKeys).
			Delete(); err != nil {
			t.Fatalf("failed to delete sys_menu rows for %s: %v", pluginID, err)
		}
	}
}

// QueryMenuByKey returns one sys_menu row by menu_key.
func QueryMenuByKey(ctx context.Context, menuKey string) (*entity.SysMenu, error) {
	var menu *entity.SysMenu
	err := dao.SysMenu.Ctx(ctx).
		Where(do.SysMenu{MenuKey: menuKey}).
		Scan(&menu)
	return menu, err
}
