// This file verifies role authorization trees use the shared role assignable
// menu projection before returning menus and checked keys to role forms.

package menu

import (
	"context"
	"testing"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	rolesvc "lina-core/internal/service/role"
	"lina-core/pkg/menutype"
)

// TestGetTreeSelectFiltersThroughRoleAssignableMenus verifies menu tree
// selection excludes platform-only menu rows supplied by the role service.
func TestGetTreeSelectFiltersThroughRoleAssignableMenus(t *testing.T) {
	ctx := context.Background()
	allowedID := insertMenuRoleTreeMenu(t, ctx, "role-tree-allowed", "system:tenant:plugin:list")
	blockedID := insertMenuRoleTreeMenu(t, ctx, "role-tree-blocked", "plugin:install")
	t.Cleanup(func() {
		cleanupMenuRoleTreeRows(t, ctx, nil, []int{allowedID, blockedID})
	})
	svc := &serviceImpl{
		roleSvc: menuRoleTreeRoleService{
			assignableIDs: map[int]struct{}{allowedID: {}},
		},
	}

	tree, err := svc.GetTreeSelect(ctx)
	if err != nil {
		t.Fatalf("get tree select: %v", err)
	}

	if !menuRoleTreeContainsID(tree, allowedID) {
		t.Fatalf("expected allowed menu %d in role tree %#v", allowedID, tree)
	}
	if menuRoleTreeContainsID(tree, blockedID) {
		t.Fatalf("did not expect blocked menu %d in role tree %#v", blockedID, tree)
	}
}

// TestGetRoleMenuTreeFiltersCheckedKeys verifies historical role-menu grants
// outside the assignable set are not returned as editable checked keys.
func TestGetRoleMenuTreeFiltersCheckedKeys(t *testing.T) {
	ctx := datascope.WithTenantForTest(context.Background(), 62201)
	roleID := insertMenuRoleTreeRole(t, ctx, 62201)
	allowedID := insertMenuRoleTreeMenu(t, ctx, "role-tree-checked-allowed", "system:tenant:plugin:list")
	blockedID := insertMenuRoleTreeMenu(t, ctx, "role-tree-checked-blocked", "plugin:install")
	insertMenuRoleTreeRoleMenu(t, ctx, roleID, allowedID, 62201)
	insertMenuRoleTreeRoleMenu(t, ctx, roleID, blockedID, 62201)
	t.Cleanup(func() {
		cleanupMenuRoleTreeRows(t, ctx, []int{roleID}, []int{allowedID, blockedID})
	})
	svc := &serviceImpl{
		roleSvc: menuRoleTreeRoleService{
			assignableIDs: map[int]struct{}{allowedID: {}},
		},
	}

	out, err := svc.GetRoleMenuTree(ctx, roleID)
	if err != nil {
		t.Fatalf("get role menu tree: %v", err)
	}

	if len(out.CheckedKeys) != 1 || out.CheckedKeys[0] != allowedID {
		t.Fatalf("expected only allowed checked key %d, got %#v", allowedID, out.CheckedKeys)
	}
	if menuRoleTreeContainsID(out.Menus, blockedID) {
		t.Fatalf("did not expect blocked menu %d in role tree %#v", blockedID, out.Menus)
	}
}

// menuRoleTreeRoleService is the narrow role facade used by menu tree tests.
type menuRoleTreeRoleService struct {
	rolesvc.Service
	assignableIDs map[int]struct{}
}

// FilterAssignableMenus returns only menus present in assignableIDs.
func (s menuRoleTreeRoleService) FilterAssignableMenus(_ context.Context, menus []*entity.SysMenu) ([]*entity.SysMenu, error) {
	filtered := make([]*entity.SysMenu, 0, len(menus))
	for _, menu := range menus {
		if menu == nil {
			continue
		}
		if _, ok := s.assignableIDs[menu.Id]; ok {
			filtered = append(filtered, menu)
		}
	}
	return filtered, nil
}

// FilterAssignableMenuIDs returns only IDs present in assignableIDs.
func (s menuRoleTreeRoleService) FilterAssignableMenuIDs(_ context.Context, menuIDs []int) ([]int, error) {
	filtered := make([]int, 0, len(menuIDs))
	for _, menuID := range menuIDs {
		if _, ok := s.assignableIDs[menuID]; ok {
			filtered = append(filtered, menuID)
		}
	}
	return filtered, nil
}

// insertMenuRoleTreeMenu inserts one minimal menu row for role-tree tests.
func insertMenuRoleTreeMenu(t *testing.T, ctx context.Context, label string, permission string) int {
	t.Helper()
	key := label
	id, err := dao.SysMenu.Ctx(ctx).Data(do.SysMenu{
		MenuKey: key,
		Name:    key,
		Perms:   permission,
		Type:    menutype.Button.String(),
		Sort:    99,
		Visible: 1,
		Status:  1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert role tree menu: %v", err)
	}
	return int(id)
}

// insertMenuRoleTreeRole inserts one role row for role-tree checked-key tests.
func insertMenuRoleTreeRole(t *testing.T, ctx context.Context, tenantID int) int {
	t.Helper()
	name := "role-tree-role"
	id, err := dao.SysRole.Ctx(ctx).Data(do.SysRole{
		Name:      name,
		Key:       name,
		Sort:      99,
		DataScope: 4,
		Status:    1,
		TenantId:  tenantID,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert role tree role: %v", err)
	}
	return int(id)
}

// insertMenuRoleTreeRoleMenu inserts one role-menu row.
func insertMenuRoleTreeRoleMenu(t *testing.T, ctx context.Context, roleID int, menuID int, tenantID int) {
	t.Helper()
	if _, err := dao.SysRoleMenu.Ctx(ctx).Data(do.SysRoleMenu{
		RoleId:   roleID,
		MenuId:   menuID,
		TenantId: tenantID,
	}).Insert(); err != nil {
		t.Fatalf("insert role tree role-menu: %v", err)
	}
}

// cleanupMenuRoleTreeRows removes role-tree test rows.
func cleanupMenuRoleTreeRows(t *testing.T, ctx context.Context, roleIDs []int, menuIDs []int) {
	t.Helper()
	if len(roleIDs) > 0 {
		if _, err := dao.SysRoleMenu.Ctx(ctx).WhereIn(dao.SysRoleMenu.Columns().RoleId, roleIDs).Delete(); err != nil {
			t.Fatalf("cleanup role tree role-menu role rows: %v", err)
		}
		if _, err := dao.SysRole.Ctx(ctx).Unscoped().WhereIn(dao.SysRole.Columns().Id, roleIDs).Delete(); err != nil {
			t.Fatalf("cleanup role tree role rows: %v", err)
		}
	}
	if len(menuIDs) > 0 {
		if _, err := dao.SysRoleMenu.Ctx(ctx).WhereIn(dao.SysRoleMenu.Columns().MenuId, menuIDs).Delete(); err != nil {
			t.Fatalf("cleanup role tree role-menu menu rows: %v", err)
		}
		if _, err := dao.SysMenu.Ctx(ctx).Unscoped().WhereIn(dao.SysMenu.Columns().Id, menuIDs).Delete(); err != nil {
			t.Fatalf("cleanup role tree menu rows: %v", err)
		}
	}
}

// menuRoleTreeContainsID reports whether one tree contains a node ID.
func menuRoleTreeContainsID(nodes []*MenuTreeNode, id int) bool {
	for _, node := range nodes {
		if node == nil {
			continue
		}
		if node.Id == id {
			return true
		}
		if menuRoleTreeContainsID(node.Children, id) {
			return true
		}
	}
	return false
}
