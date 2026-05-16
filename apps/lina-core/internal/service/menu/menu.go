// Package menu implements menu tree management, permission lookup, and
// plugin-aware filtering for the Lina core host service.
package menu

import (
	"context"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/role"
)

// Service defines the menu service contract.
type Service interface {
	// List queries a flat menu list with optional name, status, visibility, and
	// localization filters. Results pass through the configured plugin menu
	// filter, and localized searches compare both source and translated names.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// BuildTree builds a tree structure from a caller-provided flat menu list.
	// It has no database side effects and preserves the caller's menu visibility
	// and localization choices.
	BuildTree(list []*entity.SysMenu) []*MenuItem
	// GetById retrieves one menu by ID. Missing records return a menu not-found
	// business error.
	GetById(ctx context.Context, id int) (*entity.SysMenu, error)
	// GetParentName retrieves and localizes the parent menu name. Root parents
	// return the localized root label; lookup failures degrade to an empty name
	// for display compatibility.
	GetParentName(ctx context.Context, parentId int) string
	// Create creates a new menu after parent/name and icon uniqueness checks.
	// GoFrame fills timestamps, and successful changes notify the role service
	// so permission/access-topology caches can refresh.
	Create(ctx context.Context, in CreateInput) (int, error)
	// Update updates menu information after existence, uniqueness, and
	// descendant-move checks. Successful changes notify the role service so
	// permission/access-topology caches can refresh.
	Update(ctx context.Context, in UpdateInput) error
	// Delete deletes a menu and, when requested, descendants and role-menu
	// associations in one transaction. Successful changes notify the role
	// service so permission/access-topology caches can refresh.
	Delete(ctx context.Context, in DeleteInput) error
	// GetTreeSelect returns a localized menu tree for selection, including
	// directory, menu, and button types. Database errors are returned unchanged.
	GetTreeSelect(ctx context.Context) ([]*MenuTreeNode, error)
	// GetRoleMenuTree returns a localized menu tree plus menu IDs currently
	// assigned to the role. Role-menu association query errors are propagated.
	GetRoleMenuTree(ctx context.Context, roleId int) (*RoleMenuTreeOutput, error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	menuFilter MenuFilter
	i18nSvc    menuI18nTranslator
	roleSvc    role.Service
}

// New creates and returns a new menu service instance.
// Pass a non-nil menuFilter when menu listing must respect plugin-driven menu
// visibility; pass nil to use the default no-op filter.
func New(menuFilter MenuFilter, i18nSvc menuI18nTranslator, roleSvc role.Service) Service {
	if menuFilter == nil {
		menuFilter = noopMenuFilter{}
	}
	return &serviceImpl{
		menuFilter: menuFilter,
		i18nSvc:    i18nSvc,
		roleSvc:    roleSvc,
	}
}

// ListInput defines the supported filters for menu list queries.
type ListInput struct {
	Name      string
	Status    *int
	Visible   *int // Visible: 1=Show 0=Hide
	Localized bool // Localized controls whether list results are translated for runtime navigation surfaces.
}

// ListOutput defines output for List function.
type ListOutput struct {
	List []*entity.SysMenu // Menu list (flat)
}

// MenuItem represents a menu node in the tree structure.
type MenuItem struct {
	Id         int         `json:"id"`
	ParentId   int         `json:"parentId"`
	Name       string      `json:"name"`
	MenuKey    string      `json:"menuKey"`
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

// DeleteInput defines input for Delete function.
type DeleteInput struct {
	Id            int
	CascadeDelete bool
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

// RoleMenuTreeOutput defines output for role menu tree.
type RoleMenuTreeOutput struct {
	Menus       []*MenuTreeNode `json:"menus"`
	CheckedKeys []int           `json:"checkedKeys"`
}
