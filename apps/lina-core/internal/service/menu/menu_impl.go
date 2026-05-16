// This file contains menu query, tree-building, mutation, and role-tree
// operations that update access-topology state after successful changes.

package menu

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/bizerr"
)

// List queries menu list with filters.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	var (
		cols             = dao.SysMenu.Columns()
		m                = dao.SysMenu.Ctx(ctx)
		nameKeyword      = strings.TrimSpace(in.Name)
		normalizedSearch = strings.ToLower(nameKeyword)
	)
	if nameKeyword != "" && !in.Localized {
		m = m.WhereLike(cols.Name, "%"+nameKeyword+"%")
	}
	if in.Status != nil {
		m = m.Where(cols.Status, *in.Status)
	}
	if in.Visible != nil {
		m = m.Where(cols.Visible, *in.Visible)
	}
	var list []*entity.SysMenu
	err := m.OrderAsc(cols.ParentId).OrderAsc(cols.Sort).OrderAsc(cols.Id).Scan(&list)
	if err != nil {
		return nil, err
	}
	list = s.menuFilter.FilterMenus(ctx, list)
	if in.Localized {
		localizedList := make([]*entity.SysMenu, 0, len(list))
		for _, menu := range list {
			rawName := strings.ToLower(menu.Name)
			s.localizeMenuEntity(ctx, menu)
			if nameKeyword == "" || strings.Contains(rawName, normalizedSearch) || strings.Contains(strings.ToLower(menu.Name), normalizedSearch) {
				localizedList = append(localizedList, menu)
			}
		}
		list = localizedList
	}
	return &ListOutput{List: list}, nil
}

// BuildTree builds tree structure from flat menu list.
func (s *serviceImpl) BuildTree(list []*entity.SysMenu) []*MenuItem {
	nodeMap := make(map[int]*MenuItem)
	for _, m := range list {
		createdAt := ""
		if m.CreatedAt != nil {
			createdAt = m.CreatedAt.String()
		}
		updatedAt := ""
		if m.UpdatedAt != nil {
			updatedAt = m.UpdatedAt.String()
		}
		nodeMap[m.Id] = &MenuItem{
			Id:         m.Id,
			ParentId:   m.ParentId,
			Name:       m.Name,
			MenuKey:    m.MenuKey,
			Path:       m.Path,
			Component:  m.Component,
			Perms:      m.Perms,
			Icon:       m.Icon,
			Type:       m.Type,
			Sort:       m.Sort,
			Visible:    m.Visible,
			Status:     m.Status,
			IsFrame:    m.IsFrame,
			IsCache:    m.IsCache,
			QueryParam: m.QueryParam,
			Remark:     m.Remark,
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
			Children:   make([]*MenuItem, 0),
		}
	}
	var roots []*MenuItem
	for _, m := range list {
		node := nodeMap[m.Id]
		if parent, ok := nodeMap[m.ParentId]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			roots = append(roots, node)
		}
	}
	return roots
}

// GetById retrieves menu by ID.
func (s *serviceImpl) GetById(ctx context.Context, id int) (*entity.SysMenu, error) {
	var menu *entity.SysMenu
	err := dao.SysMenu.Ctx(ctx).Where(do.SysMenu{Id: id}).Scan(&menu)
	if err != nil {
		return nil, err
	}
	if menu == nil {
		return nil, bizerr.NewCode(CodeMenuNotFound)
	}
	return menu, nil
}

// GetParentName retrieves parent menu name.
func (s *serviceImpl) GetParentName(ctx context.Context, parentId int) string {
	if parentId == 0 {
		return s.i18nSvc.Translate(ctx, "menu.root.title", "Main Category")
	}
	parent, err := s.GetById(ctx, parentId)
	if err != nil {
		return ""
	}
	s.localizeMenuEntity(ctx, parent)
	return parent.Name
}

// Create creates a new menu.
func (s *serviceImpl) Create(ctx context.Context, in CreateInput) (int, error) {
	if err := s.checkNameUnique(ctx, in.Name, in.ParentId, 0); err != nil {
		return 0, err
	}
	if err := s.checkIconUnique(ctx, in.Type, in.Icon, 0); err != nil {
		return 0, err
	}
	id, err := dao.SysMenu.Ctx(ctx).Data(do.SysMenu{
		ParentId:   in.ParentId,
		Name:       in.Name,
		Path:       in.Path,
		Component:  in.Component,
		Perms:      in.Perms,
		Icon:       normalizeMenuIcon(in.Icon),
		Type:       in.Type,
		Sort:       in.Sort,
		Visible:    in.Visible,
		Status:     in.Status,
		IsFrame:    in.IsFrame,
		IsCache:    in.IsCache,
		QueryParam: in.QueryParam,
		Remark:     in.Remark,
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}
	s.roleSvc.NotifyAccessTopologyChanged(ctx)
	return int(id), nil
}

// Update updates menu information.
func (s *serviceImpl) Update(ctx context.Context, in UpdateInput) error {
	menu, err := s.GetById(ctx, in.Id)
	if err != nil {
		return err
	}
	parentId := menu.ParentId
	if in.ParentId != nil {
		parentId = *in.ParentId
	}
	if err := s.checkNameUnique(ctx, in.Name, parentId, in.Id); err != nil {
		return err
	}
	if in.ParentId != nil {
		if *in.ParentId == in.Id {
			return bizerr.NewCode(CodeMenuMoveToSelfDenied)
		}
		if s.isDescendant(ctx, in.Id, *in.ParentId) {
			return bizerr.NewCode(CodeMenuMoveToDescendantDenied)
		}
	}
	menuType := menu.Type
	if in.Type != nil {
		menuType = *in.Type
	}
	menuIcon := menu.Icon
	if in.Icon != nil {
		menuIcon = *in.Icon
	}
	if err := s.checkIconUnique(ctx, menuType, menuIcon, in.Id); err != nil {
		return err
	}
	data := do.SysMenu{}
	if in.ParentId != nil {
		data.ParentId = *in.ParentId
	}
	if in.Name != "" {
		data.Name = in.Name
	}
	if in.Path != nil {
		data.Path = *in.Path
	}
	if in.Component != nil {
		data.Component = *in.Component
	}
	if in.Perms != nil {
		data.Perms = *in.Perms
	}
	if in.Icon != nil {
		data.Icon = normalizeMenuIcon(*in.Icon)
	}
	if in.Type != nil {
		data.Type = *in.Type
	}
	if in.Sort != nil {
		data.Sort = *in.Sort
	}
	if in.Visible != nil {
		data.Visible = *in.Visible
	}
	if in.Status != nil {
		data.Status = *in.Status
	}
	if in.IsFrame != nil {
		data.IsFrame = *in.IsFrame
	}
	if in.IsCache != nil {
		data.IsCache = *in.IsCache
	}
	if in.QueryParam != nil {
		data.QueryParam = *in.QueryParam
	}
	if in.Remark != nil {
		data.Remark = *in.Remark
	}
	_, err = dao.SysMenu.Ctx(ctx).Where(do.SysMenu{Id: in.Id}).Data(data).Update()
	if err != nil {
		return err
	}
	s.roleSvc.NotifyAccessTopologyChanged(ctx)
	return nil
}

// Delete deletes a menu.
func (s *serviceImpl) Delete(ctx context.Context, in DeleteInput) error {
	if _, err := s.GetById(ctx, in.Id); err != nil {
		return err
	}
	cols := dao.SysMenu.Columns()
	childCount, err := dao.SysMenu.Ctx(ctx).Where(cols.ParentId, in.Id).Count()
	if err != nil {
		return err
	}
	if childCount > 0 && !in.CascadeDelete {
		return bizerr.NewCode(CodeMenuHasChildrenDeleteDenied)
	}
	err = dao.SysMenu.Ctx(ctx).Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		menuIds := []int{in.Id}
		if in.CascadeDelete && childCount > 0 {
			descendants, err := s.getDescendantIds(ctx, in.Id)
			if err != nil {
				return err
			}
			menuIds = append(menuIds, descendants...)
		}
		if _, err := dao.SysMenu.Ctx(ctx).WhereIn(cols.Id, menuIds).Delete(); err != nil {
			return err
		}
		rmCols := dao.SysRoleMenu.Columns()
		_, err := dao.SysRoleMenu.Ctx(ctx).WhereIn(rmCols.MenuId, menuIds).Delete()
		return err
	})
	if err != nil {
		return err
	}
	s.roleSvc.NotifyAccessTopologyChanged(ctx)
	return nil
}

// GetTreeSelect returns menu tree for selection (includes all menu types: D/M/B).
func (s *serviceImpl) GetTreeSelect(ctx context.Context) ([]*MenuTreeNode, error) {
	cols := dao.SysMenu.Columns()
	var list []*entity.SysMenu
	err := dao.SysMenu.Ctx(ctx).OrderAsc(cols.ParentId).OrderAsc(cols.Sort).OrderAsc(cols.Id).Scan(&list)
	if err != nil {
		return nil, err
	}
	s.localizeMenuEntities(ctx, list)
	return s.buildPermissionTreeNodes(ctx, list), nil
}

// GetRoleMenuTree returns menu tree with checked keys for a role.
func (s *serviceImpl) GetRoleMenuTree(ctx context.Context, roleId int) (*RoleMenuTreeOutput, error) {
	menus, err := s.GetTreeSelect(ctx)
	if err != nil {
		return nil, err
	}
	rmCols := dao.SysRoleMenu.Columns()
	var roleMenus []*entity.SysRoleMenu
	err = dao.SysRoleMenu.Ctx(ctx).Where(rmCols.RoleId, roleId).Scan(&roleMenus)
	if err != nil {
		return nil, err
	}
	checkedKeys := make([]int, 0, len(roleMenus))
	for _, rm := range roleMenus {
		checkedKeys = append(checkedKeys, rm.MenuId)
	}
	return &RoleMenuTreeOutput{Menus: menus, CheckedKeys: checkedKeys}, nil
}

// checkNameUnique checks if the menu name is unique under the same parent.
func (s *serviceImpl) checkNameUnique(ctx context.Context, name string, parentId int, excludeId int) error {
	cols := dao.SysMenu.Columns()
	m := dao.SysMenu.Ctx(ctx).Where(cols.Name, name).Where(cols.ParentId, parentId)
	if excludeId > 0 {
		m = m.WhereNot(cols.Id, excludeId)
	}
	count, err := m.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return bizerr.NewCode(CodeMenuNameExists)
	}
	return nil
}

// isDescendant checks if targetId is a descendant of parentId.
func (s *serviceImpl) isDescendant(ctx context.Context, parentId int, targetId int) bool {
	if parentId == targetId {
		return false
	}
	cols := dao.SysMenu.Columns()
	var menus []*entity.SysMenu
	err := dao.SysMenu.Ctx(ctx).Fields(cols.Id, cols.ParentId).Scan(&menus)
	if err != nil {
		return false
	}
	childrenByParent := make(map[int][]int, len(menus))
	for _, menu := range menus {
		childrenByParent[menu.ParentId] = append(childrenByParent[menu.ParentId], menu.Id)
	}
	queue := append([]int(nil), childrenByParent[parentId]...)
	visited := make(map[int]struct{}, len(menus))
	for len(queue) > 0 {
		currentId := queue[0]
		queue = queue[1:]
		if _, ok := visited[currentId]; ok {
			continue
		}
		visited[currentId] = struct{}{}
		if currentId == targetId {
			return true
		}
		queue = append(queue, childrenByParent[currentId]...)
	}
	return false
}

// getDescendantIds returns all descendant menu IDs.
func (s *serviceImpl) getDescendantIds(ctx context.Context, menuId int) ([]int, error) {
	cols := dao.SysMenu.Columns()
	var result []int
	parentIds := []int{menuId}
	for len(parentIds) > 0 {
		var children []*entity.SysMenu
		err := dao.SysMenu.Ctx(ctx).WhereIn(cols.ParentId, parentIds).Scan(&children)
		if err != nil {
			return nil, err
		}
		parentIds = make([]int, 0)
		for _, child := range children {
			result = append(result, child.Id)
			parentIds = append(parentIds, child.Id)
		}
	}
	return result, nil
}
