// This file verifies menu service persistence-backed helpers and role
// assignable tree flows implemented by the service methods.

package menu

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "lina-core/pkg/dbdriver"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	rolesvc "lina-core/internal/service/role"
	"lina-core/pkg/menutype"
)

// TestIsDescendantUsesInMemoryHierarchy verifies the optimized descendant
// lookup handles direct, cross-level, self, sibling, and missing-target cases.
func TestIsDescendantUsesInMemoryHierarchy(t *testing.T) {
	var (
		ctx    = context.Background()
		svc    = &serviceImpl{}
		prefix = fmt.Sprintf("menu-descendant-%d", time.Now().UnixNano())
	)

	var (
		rootID       = insertTestMenu(t, ctx, prefix+"-root", 0)
		childID      = insertTestMenu(t, ctx, prefix+"-child", rootID)
		grandchildID = insertTestMenu(t, ctx, prefix+"-grandchild", childID)
		siblingID    = insertTestMenu(t, ctx, prefix+"-sibling", rootID)
	)
	t.Cleanup(func() {
		if _, err := dao.SysMenu.Ctx(ctx).
			Unscoped().
			WhereIn(dao.SysMenu.Columns().Id, []int{rootID, childID, grandchildID, siblingID}).
			Delete(); err != nil {
			t.Fatalf("cleanup menu hierarchy: %v", err)
		}
	})

	if !svc.isDescendant(ctx, rootID, childID) {
		t.Fatal("expected child to be descendant of root")
	}
	if !svc.isDescendant(ctx, rootID, grandchildID) {
		t.Fatal("expected grandchild to be descendant of root")
	}
	if svc.isDescendant(ctx, rootID, rootID) {
		t.Fatal("expected menu to not be descendant of itself")
	}
	if svc.isDescendant(ctx, childID, siblingID) {
		t.Fatal("expected sibling to not be descendant of child")
	}
	if svc.isDescendant(ctx, rootID, grandchildID+1000000) {
		t.Fatal("expected missing target to not be reported as descendant")
	}
}

// TestGetTreeSelectFiltersThroughRoleAssignableMenus verifies menu tree
// selection excludes platform-only menu rows supplied by the role service.
func TestGetTreeSelectFiltersThroughRoleAssignableMenus(t *testing.T) {
	var (
		ctx       = context.Background()
		allowedID = insertMenuRoleTreeMenu(t, ctx, "role-tree-allowed", "system:tenant:plugin:list")
		blockedID = insertMenuRoleTreeMenu(t, ctx, "role-tree-blocked", "plugin:install")
	)
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
	var (
		ctx       = datascope.WithTenantScope(context.Background(), 62201)
		roleID    = insertMenuRoleTreeRole(t, ctx, 62201)
		allowedID = insertMenuRoleTreeMenu(t, ctx, "role-tree-checked-allowed", "system:tenant:plugin:list")
		blockedID = insertMenuRoleTreeMenu(t, ctx, "role-tree-checked-blocked", "plugin:install")
	)
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

// TestUpdateCascadesDisableToDescendants verifies disabling a parent marks all
// descendants disabled without changing their visibility.
func TestUpdateCascadesDisableToDescendants(t *testing.T) {
	var (
		ctx    = context.Background()
		prefix = fmt.Sprintf("menu-cascade-disable-%d", time.Now().UnixNano())
		role   = &menuNotifyRoleService{}
		svc    = &serviceImpl{roleSvc: role}
	)

	rootID := insertTestMenu(t, ctx, prefix+"-root", 0)
	childID := insertTestMenu(t, ctx, prefix+"-child", rootID)
	grandchildID := insertTestMenu(t, ctx, prefix+"-grandchild", childID)
	siblingID := insertTestMenu(t, ctx, prefix+"-sibling", 0)
	t.Cleanup(func() {
		cleanupTestMenus(t, ctx, rootID, childID, grandchildID, siblingID)
	})

	statusDisabled := 0
	if err := svc.Update(ctx, UpdateInput{Id: rootID, Status: &statusDisabled}); err != nil {
		t.Fatalf("disable root menu: %v", err)
	}
	if role.notified == 0 {
		t.Fatal("expected access topology notify after cascade disable")
	}

	assertMenuFlags(t, ctx, rootID, 1, 0)
	assertMenuFlags(t, ctx, childID, 1, 0)
	assertMenuFlags(t, ctx, grandchildID, 1, 0)
	assertMenuFlags(t, ctx, siblingID, 1, 1)
}

// TestUpdateCascadesHideToDescendants verifies hiding a parent hides all
// descendants without changing their status.
func TestUpdateCascadesHideToDescendants(t *testing.T) {
	var (
		ctx    = context.Background()
		prefix = fmt.Sprintf("menu-cascade-hide-%d", time.Now().UnixNano())
		svc    = &serviceImpl{roleSvc: &menuNotifyRoleService{}}
	)

	rootID := insertTestMenu(t, ctx, prefix+"-root", 0)
	childID := insertTestMenu(t, ctx, prefix+"-child", rootID)
	grandchildID := insertTestMenu(t, ctx, prefix+"-grandchild", childID)
	t.Cleanup(func() {
		cleanupTestMenus(t, ctx, rootID, childID, grandchildID)
	})

	visibleHidden := 0
	if err := svc.Update(ctx, UpdateInput{Id: rootID, Visible: &visibleHidden}); err != nil {
		t.Fatalf("hide root menu: %v", err)
	}

	assertMenuFlags(t, ctx, rootID, 0, 1)
	assertMenuFlags(t, ctx, childID, 0, 1)
	assertMenuFlags(t, ctx, grandchildID, 0, 1)
}

// TestUpdateEnableCascadesToDescendants verifies enabling a parent restores
// previously disabled descendants.
func TestUpdateEnableCascadesToDescendants(t *testing.T) {
	var (
		ctx    = context.Background()
		prefix = fmt.Sprintf("menu-cascade-enable-%d", time.Now().UnixNano())
		svc    = &serviceImpl{roleSvc: &menuNotifyRoleService{}}
	)

	rootID := insertTestMenu(t, ctx, prefix+"-root", 0)
	childID := insertTestMenu(t, ctx, prefix+"-child", rootID)
	grandchildID := insertTestMenu(t, ctx, prefix+"-grandchild", childID)
	t.Cleanup(func() {
		cleanupTestMenus(t, ctx, rootID, childID, grandchildID)
	})

	statusDisabled := 0
	statusEnabled := 1
	if err := svc.Update(ctx, UpdateInput{Id: rootID, Status: &statusDisabled}); err != nil {
		t.Fatalf("disable root menu: %v", err)
	}
	if err := svc.Update(ctx, UpdateInput{Id: rootID, Status: &statusEnabled}); err != nil {
		t.Fatalf("enable root menu: %v", err)
	}

	assertMenuFlags(t, ctx, rootID, 1, 1)
	assertMenuFlags(t, ctx, childID, 1, 1)
	assertMenuFlags(t, ctx, grandchildID, 1, 1)
}

// TestUpdateShowCascadesToDescendants verifies showing a parent restores
// previously hidden descendants.
func TestUpdateShowCascadesToDescendants(t *testing.T) {
	var (
		ctx    = context.Background()
		prefix = fmt.Sprintf("menu-cascade-show-%d", time.Now().UnixNano())
		svc    = &serviceImpl{roleSvc: &menuNotifyRoleService{}}
	)

	rootID := insertTestMenu(t, ctx, prefix+"-root", 0)
	childID := insertTestMenu(t, ctx, prefix+"-child", rootID)
	grandchildID := insertTestMenu(t, ctx, prefix+"-grandchild", childID)
	t.Cleanup(func() {
		cleanupTestMenus(t, ctx, rootID, childID, grandchildID)
	})

	visibleHidden := 0
	visibleShown := 1
	if err := svc.Update(ctx, UpdateInput{Id: rootID, Visible: &visibleHidden}); err != nil {
		t.Fatalf("hide root menu: %v", err)
	}
	if err := svc.Update(ctx, UpdateInput{Id: rootID, Visible: &visibleShown}); err != nil {
		t.Fatalf("show root menu: %v", err)
	}

	assertMenuFlags(t, ctx, rootID, 1, 1)
	assertMenuFlags(t, ctx, childID, 1, 1)
	assertMenuFlags(t, ctx, grandchildID, 1, 1)
}

// TestUpdateDisableWithoutDescendants updates only the target row.
func TestUpdateDisableWithoutDescendants(t *testing.T) {
	var (
		ctx    = context.Background()
		prefix = fmt.Sprintf("menu-cascade-leaf-%d", time.Now().UnixNano())
		svc    = &serviceImpl{roleSvc: &menuNotifyRoleService{}}
	)

	leafID := insertTestMenu(t, ctx, prefix+"-leaf", 0)
	t.Cleanup(func() {
		cleanupTestMenus(t, ctx, leafID)
	})

	statusDisabled := 0
	if err := svc.Update(ctx, UpdateInput{Id: leafID, Status: &statusDisabled}); err != nil {
		t.Fatalf("disable leaf menu: %v", err)
	}
	assertMenuFlags(t, ctx, leafID, 1, 0)
}

// menuNotifyRoleService records topology notifications for cascade update tests.
type menuNotifyRoleService struct {
	rolesvc.Service
	notified int
}

// NotifyAccessTopologyChanged records that a topology revision was requested.
func (s *menuNotifyRoleService) NotifyAccessTopologyChanged(context.Context) {
	s.notified++
}

// assertMenuFlags loads one menu and asserts its visible/status values.
func assertMenuFlags(t *testing.T, ctx context.Context, id int, visible int, status int) {
	t.Helper()
	var menu *entity.SysMenu
	if err := dao.SysMenu.Ctx(ctx).Where(do.SysMenu{Id: id}).Scan(&menu); err != nil {
		t.Fatalf("load menu %d: %v", id, err)
	}
	if menu == nil {
		t.Fatalf("menu %d not found", id)
	}
	if menu.Visible != visible || menu.Status != status {
		t.Fatalf("menu %d flags visible=%d status=%d, want visible=%d status=%d",
			id, menu.Visible, menu.Status, visible, status)
	}
}

// cleanupTestMenus hard-deletes the given menu rows after a cascade test.
func cleanupTestMenus(t *testing.T, ctx context.Context, ids ...int) {
	t.Helper()
	if len(ids) == 0 {
		return
	}
	if _, err := dao.SysMenu.Ctx(ctx).
		Unscoped().
		WhereIn(dao.SysMenu.Columns().Id, ids).
		Delete(); err != nil {
		t.Fatalf("cleanup test menus: %v", err)
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

// insertTestMenu inserts one minimal menu row for hierarchy tests.
func insertTestMenu(t *testing.T, ctx context.Context, key string, parentID int) int {
	t.Helper()

	id, err := dao.SysMenu.Ctx(ctx).Data(do.SysMenu{
		ParentId: parentID,
		MenuKey:  key,
		Name:     key,
		Type:     "M",
		Sort:     1,
		Visible:  1,
		Status:   1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert test menu %s: %v", key, err)
	}
	return int(id)
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
