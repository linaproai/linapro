// Package menu implements the menu service contract, stable host menu
// metadata, plugin-aware filtering, and default construction for Lina core.
package menu

import (
	"context"

	"lina-core/internal/model/entity"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/role"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
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
	// descendant-move checks. When status or visible is written, all
	// descendant menus receive the same field value in the same transaction
	// (including enable/show). Successful changes notify the role service so
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

// menuFilter defines the narrow dependency required by the menu service to hide
// menus that should not be exposed for the current host state.
type menuFilter interface {
	// FilterMenus returns only the menus that should remain visible.
	FilterMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	menuFilter menuFilter
	i18nSvc    i18nsvc.Service
	roleSvc    role.Service
	tenantSvc  tenantspi.Service
}

// New creates and returns a new menu service instance.
// Pass a non-nil menuFilter when menu listing must respect plugin-driven menu
// visibility; pass nil to use the default no-op filter.
func New(menuFilter menuFilter, i18nSvc i18nsvc.Service, roleSvc role.Service, tenantSvc tenantspi.Service) Service {
	if menuFilter == nil {
		menuFilter = noopMenuFilter{}
	}
	return &serviceImpl{
		menuFilter: menuFilter,
		i18nSvc:    i18nSvc,
		roleSvc:    roleSvc,
		tenantSvc:  tenantSvc,
	}
}

// noopMenuFilter leaves the menu list unchanged when no external filter is injected.
type noopMenuFilter struct{}

// FilterMenus returns the original menu list unchanged.
func (noopMenuFilter) FilterMenus(_ context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	return menus
}

// Stable host catalog menu keys.
const (
	// Dashboard is the stable key of the workspace catalog.
	Dashboard = "dashboard"
	// IAM is the stable key of the identity-and-access catalog.
	IAM = "iam"
	// Org is the stable key of the organization catalog.
	Org = "org"
	// Setting is the stable key of the system-settings catalog.
	Setting = "setting"
	// Content is the stable key of the content catalog.
	Content = "content"
	// Monitor is the stable key of the monitoring catalog.
	Monitor = "monitor"
	// Scheduler is the stable key of the scheduled-job catalog.
	Scheduler = "scheduler"
	// Extension is the stable key of the extension-governance catalog.
	Extension = "extension"
	// Platform is the stable key of the platform-administration catalog.
	Platform = "platform"
	// Developer is the stable key of the developer-support catalog.
	Developer = "developer"
)

var stableCatalogKeys = map[string]struct{}{
	Dashboard: {},
	IAM:       {},
	Org:       {},
	Setting:   {},
	Content:   {},
	Monitor:   {},
	Scheduler: {},
	Extension: {},
	Platform:  {},
	Developer: {},
}

// IsStableCatalogKey reports whether the given menu key belongs to one
// host-owned top-level catalog.
func IsStableCatalogKey(menuKey string) bool {
	_, ok := stableCatalogKeys[menuKey]
	return ok
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
	CreatedAt  *int64      `json:"createdAt"`
	UpdatedAt  *int64      `json:"updatedAt"`
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
