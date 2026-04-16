// Package menu implements menu tree management, permission lookup, and
// plugin-aware filtering for the Lina core host service.
package menu

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/role"
	"lina-core/pkg/logger"
)

// Service defines the menu service contract.
type Service interface {
	// List queries menu list with filters.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// BuildTree builds tree structure from flat menu list.
	BuildTree(list []*entity.SysMenu) []*MenuItem
	// GetById retrieves menu by ID.
	GetById(ctx context.Context, id int) (*entity.SysMenu, error)
	// GetParentName retrieves parent menu name.
	GetParentName(ctx context.Context, parentId int) string
	// Create creates a new menu.
	Create(ctx context.Context, in CreateInput) (int, error)
	// Update updates menu information.
	Update(ctx context.Context, in UpdateInput) error
	// Delete deletes a menu.
	Delete(ctx context.Context, in DeleteInput) error
	// GetTreeSelect returns menu tree for selection (includes all menu types: D/M/B).
	GetTreeSelect(ctx context.Context) ([]*MenuTreeNode, error)
	// GetRoleMenuTree returns menu tree with checked keys for a role.
	GetRoleMenuTree(ctx context.Context, roleId int) (*RoleMenuTreeOutput, error)
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	pluginSvc pluginsvc.Service
	roleSvc   role.Service
}

func New() Service {
	return &serviceImpl{
		pluginSvc: pluginsvc.New(),
		roleSvc:   role.New(),
	}
}

type ListInput struct {
	Name    string
	Status  *int
	Visible *int // Visible: 1=Show 0=Hide
}

// ListOutput defines output for List function.
type ListOutput struct {
	List []*entity.SysMenu // Menu list (flat)
}

// List queries menu list with filters.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	var (
		cols = dao.SysMenu.Columns()
		m    = dao.SysMenu.Ctx(ctx)
	)

	// Apply filters
	if in.Name != "" {
		m = m.WhereLike(cols.Name, "%"+in.Name+"%")
	}
	if in.Status != nil {
		m = m.Where(cols.Status, *in.Status)
	}
	if in.Visible != nil {
		m = m.Where(cols.Visible, *in.Visible)
	}

	// Query all, ordered by sort ASC
	var list []*entity.SysMenu
	err := m.OrderAsc(cols.ParentId).OrderAsc(cols.Sort).OrderAsc(cols.Id).Scan(&list)
	if err != nil {
		return nil, err
	}
	list = s.pluginSvc.FilterMenus(ctx, list)

	return &ListOutput{
		List: list,
	}, nil
}

// MenuItem represents a menu node in the tree structure.
type MenuItem struct {
	Id         int         `json:"id"`
	ParentId   int         `json:"parentId"`
	Name       string      `json:"name"`
	Path       string      `json:"path"`
	Component  string      `json:"component"`
	Perms      string      `json:"perms"`
	Icon       string      `json:"icon"`
	Type       string      `json:"type"`
	Sort       int         `json:"sort"`
	Visible    int         `json:"visible"`
	Status     int         `json:"status"`
	IsFrame    int         `json:"isFrame"`
	IsCache    int         `json:"isCache"`
	QueryParam string      `json:"queryParam"`
	Remark     string      `json:"remark"`
	CreatedAt  string      `json:"createdAt"`
	UpdatedAt  string      `json:"updatedAt"`
	Children   []*MenuItem `json:"children"`
}

// BuildTree builds tree structure from flat menu list.
func (s *serviceImpl) BuildTree(list []*entity.SysMenu) []*MenuItem {
	// Build map for quick lookup
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

	// Build tree
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
	err := dao.SysMenu.Ctx(ctx).
		Where(do.SysMenu{Id: id}).
		Scan(&menu)
	if err != nil {
		return nil, err
	}
	if menu == nil {
		return nil, gerror.New("菜单不存在")
	}
	return menu, nil
}

// GetParentName retrieves parent menu name.
func (s *serviceImpl) GetParentName(ctx context.Context, parentId int) string {
	if parentId == 0 {
		return "主类目"
	}
	parent, err := s.GetById(ctx, parentId)
	if err != nil {
		return ""
	}
	return parent.Name
}

// CreateInput defines input for Create function.
type CreateInput struct {
	ParentId   int
	Name       string
	Path       string
	Component  string
	Perms      string
	Icon       string
	Type       string
	Sort       int
	Visible    int
	Status     int
	IsFrame    int
	IsCache    int
	QueryParam string
	Remark     string
}

// Create creates a new menu.
func (s *serviceImpl) Create(ctx context.Context, in CreateInput) (int, error) {
	// Check name uniqueness under same parent
	if err := s.checkNameUnique(ctx, in.Name, in.ParentId, 0); err != nil {
		return 0, err
	}

	// Insert menu (GoFrame auto-fills created_at and updated_at)
	id, err := dao.SysMenu.Ctx(ctx).Data(do.SysMenu{
		ParentId:   in.ParentId,
		Name:       in.Name,
		Path:       in.Path,
		Component:  in.Component,
		Perms:      in.Perms,
		Icon:       in.Icon,
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

// UpdateInput defines input for Update function.
type UpdateInput struct {
	Id         int
	ParentId   *int
	Name       string
	Path       *string
	Component  *string
	Perms      *string
	Icon       *string
	Type       *string
	Sort       *int
	Visible    *int
	Status     *int
	IsFrame    *int
	IsCache    *int
	QueryParam *string
	Remark     *string
}

// Update updates menu information.
func (s *serviceImpl) Update(ctx context.Context, in UpdateInput) error {
	// Check menu exists
	menu, err := s.GetById(ctx, in.Id)
	if err != nil {
		return err
	}

	// Check name uniqueness under same parent (excluding self)
	parentId := menu.ParentId
	if in.ParentId != nil {
		parentId = *in.ParentId
	}
	if err := s.checkNameUnique(ctx, in.Name, parentId, in.Id); err != nil {
		return err
	}

	// Check not moving to self or descendant
	if in.ParentId != nil {
		if *in.ParentId == in.Id {
			return gerror.New("不能将菜单移动到自己")
		}
		// Check if new parent is a descendant
		if s.isDescendant(ctx, in.Id, *in.ParentId) {
			return gerror.New("不能将菜单移动到自己的子菜单下")
		}
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
		data.Icon = *in.Icon
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

// DeleteInput defines input for Delete function.
type DeleteInput struct {
	Id            int
	CascadeDelete bool
}

// Delete deletes a menu.
func (s *serviceImpl) Delete(ctx context.Context, in DeleteInput) error {
	// Check menu exists
	_, err := s.GetById(ctx, in.Id)
	if err != nil {
		return err
	}

	// Check children
	cols := dao.SysMenu.Columns()
	childCount, err := dao.SysMenu.Ctx(ctx).
		Where(cols.ParentId, in.Id).
		Count()
	if err != nil {
		return err
	}

	if childCount > 0 && !in.CascadeDelete {
		return gerror.New("存在子菜单，不允许删除")
	}

	// Use transaction for cascade delete
	err = dao.SysMenu.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// Collect menu IDs to delete
		menuIds := []int{in.Id}
		if in.CascadeDelete && childCount > 0 {
			// Get all descendant menu IDs
			descendants, err := s.getDescendantIds(ctx, in.Id)
			if err != nil {
				return err
			}
			menuIds = append(menuIds, descendants...)
		}

		// Delete menus
		_, err = dao.SysMenu.Ctx(ctx).
			WhereIn(cols.Id, menuIds).
			Delete()
		if err != nil {
			return err
		}

		// Delete role-menu associations
		rmCols := dao.SysRoleMenu.Columns()
		_, err = dao.SysRoleMenu.Ctx(ctx).
			WhereIn(rmCols.MenuId, menuIds).
			Delete()
		if err != nil {
			logger.Warningf(ctx, "failed to delete role-menu associations: %v", err)
		}

		return nil
	})
	if err != nil {
		return err
	}
	s.roleSvc.NotifyAccessTopologyChanged(ctx)
	return nil
}

// MenuTreeNode represents a node in the tree select.
type MenuTreeNode struct {
	Id       int             `json:"id"`
	ParentId int             `json:"parentId"`
	Label    string          `json:"label"`
	Type     string          `json:"type,omitempty"`
	Icon     string          `json:"icon,omitempty"`
	Children []*MenuTreeNode `json:"children"`
}

// GetTreeSelect returns menu tree for selection (includes all menu types: D/M/B).
func (s *serviceImpl) GetTreeSelect(ctx context.Context) ([]*MenuTreeNode, error) {
	cols := dao.SysMenu.Columns()

	// Query all menus (including button type for permission selection)
	var list []*entity.SysMenu
	err := dao.SysMenu.Ctx(ctx).
		OrderAsc(cols.ParentId).OrderAsc(cols.Sort).OrderAsc(cols.Id).
		Scan(&list)
	if err != nil {
		return nil, err
	}

	// Build map
	nodeMap := make(map[int]*MenuTreeNode)
	for _, m := range list {
		nodeMap[m.Id] = &MenuTreeNode{
			Id:       m.Id,
			ParentId: m.ParentId,
			Label:    m.Name,
			Type:     m.Type,
			Icon:     m.Icon,
			Children: make([]*MenuTreeNode, 0),
		}
	}

	// Build tree
	var roots []*MenuTreeNode
	for _, m := range list {
		node := nodeMap[m.Id]
		if parent, ok := nodeMap[m.ParentId]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			roots = append(roots, node)
		}
	}

	return roots, nil
}

// RoleMenuTreeOutput defines output for role menu tree.
type RoleMenuTreeOutput struct {
	Menus       []*MenuTreeNode `json:"menus"`
	CheckedKeys []int           `json:"checkedKeys"`
}

// GetRoleMenuTree returns menu tree with checked keys for a role.
func (s *serviceImpl) GetRoleMenuTree(ctx context.Context, roleId int) (*RoleMenuTreeOutput, error) {
	// Get menu tree
	menus, err := s.GetTreeSelect(ctx)
	if err != nil {
		return nil, err
	}

	// Get checked menu IDs for the role
	rmCols := dao.SysRoleMenu.Columns()
	var roleMenus []*entity.SysRoleMenu
	err = dao.SysRoleMenu.Ctx(ctx).
		Where(rmCols.RoleId, roleId).
		Scan(&roleMenus)
	if err != nil {
		return nil, err
	}

	checkedKeys := make([]int, 0, len(roleMenus))
	for _, rm := range roleMenus {
		checkedKeys = append(checkedKeys, rm.MenuId)
	}

	return &RoleMenuTreeOutput{
		Menus:       menus,
		CheckedKeys: checkedKeys,
	}, nil
}

// checkNameUnique checks if the menu name is unique under the same parent.
func (s *serviceImpl) checkNameUnique(ctx context.Context, name string, parentId int, excludeId int) error {
	cols := dao.SysMenu.Columns()
	m := dao.SysMenu.Ctx(ctx).
		Where(cols.Name, name).
		Where(cols.ParentId, parentId)
	if excludeId > 0 {
		m = m.WhereNot(cols.Id, excludeId)
	}
	count, err := m.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return gerror.New("同级菜单名称已存在")
	}
	return nil
}

// isDescendant checks if targetId is a descendant of parentId.
func (s *serviceImpl) isDescendant(ctx context.Context, parentId int, targetId int) bool {
	cols := dao.SysMenu.Columns()

	// Get all children of parentId recursively
	parentIds := []int{parentId}
	for len(parentIds) > 0 {
		var children []*entity.SysMenu
		err := dao.SysMenu.Ctx(ctx).
			WhereIn(cols.ParentId, parentIds).
			Scan(&children)
		if err != nil {
			return false
		}

		parentIds = make([]int, 0)
		for _, child := range children {
			if child.Id == targetId {
				return true
			}
			parentIds = append(parentIds, child.Id)
		}
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
		err := dao.SysMenu.Ctx(ctx).
			WhereIn(cols.ParentId, parentIds).
			Scan(&children)
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
