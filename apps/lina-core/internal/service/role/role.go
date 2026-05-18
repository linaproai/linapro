// Package role implements role management, permission lookup, and shared access
// context caching for the Lina core host service.
package role

import (
	"context"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	orgcapsvc "lina-core/internal/service/orgcap"
	tenantcapsvc "lina-core/internal/service/tenantcap"
)

const (
	// builtinAdminRoleKey identifies the protected built-in administrator role.
	builtinAdminRoleKey = "admin"
	// builtinAdminRoleNameI18n is the runtime i18n key for the built-in administrator role.
	builtinAdminRoleNameI18n = "role.builtin.admin.name"
	// builtinUserRoleKey identifies the protected built-in standard user role.
	builtinUserRoleKey = "user"
	// builtinUserRoleNameI18n is the runtime i18n key for the built-in standard user role.
	builtinUserRoleNameI18n = "role.builtin.user.name"
)

// Role data-scope values stored in sys_role.data_scope.
const (
	// roleDataScopeAll grants access to all governed records across tenant boundaries.
	roleDataScopeAll = 1
	// roleDataScopeTenant grants access to all governed records in the current tenant.
	roleDataScopeTenant = 2
	// roleDataScopeDept grants access to records owned by users in the current department scope.
	roleDataScopeDept = 3
	// roleDataScopeSelf grants access only to the current user's own records.
	roleDataScopeSelf = 4
)

// PermissionMenuFilter defines the narrow dependency required by the role
// service to hide plugin-owned permission menus that are not currently active.
type PermissionMenuFilter interface {
	// FilterPermissionMenus returns only the permission menus that should remain
	// effective for the current host and plugin state.
	FilterPermissionMenus(ctx context.Context, menus []*entity.SysMenu) []*entity.SysMenu
}

// OrganizationCapabilityState defines the narrow organization-capability
// dependency role needs to validate organization-dependent data-scope values.
type OrganizationCapabilityState interface {
	// Enabled reports whether organization capability is currently usable.
	Enabled(ctx context.Context) bool
}

// pluginEnablementState defines the plugin-status reader used to derive
// organization capability state in production controllers.
type pluginEnablementState interface {
	// IsEnabled returns whether the given plugin ID is currently enabled.
	IsEnabled(ctx context.Context, pluginID string) bool
}

// RoleQueryService defines read-only role management operations.
type RoleQueryService interface {
	// List queries role list with pagination.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// GetById retrieves role by ID.
	GetById(ctx context.Context, id int) (*entity.SysRole, error)
	// GetDetail retrieves role detail with menu IDs.
	GetDetail(ctx context.Context, id int) (*GetDetailOutput, error)
	// GetOptions returns role options for dropdown.
	GetOptions(ctx context.Context) ([]*OptionItem, error)
	// DisplayName returns the read-only display name for one role, localizing
	// protected built-in roles while preserving custom role names.
	DisplayName(ctx context.Context, role *entity.SysRole) string
}

// RoleMutationService defines role create, update, delete, and status operations.
type RoleMutationService interface {
	// Create creates a new role.
	Create(ctx context.Context, in CreateInput) (int, error)
	// Update updates role information.
	Update(ctx context.Context, in UpdateInput) error
	// Delete deletes a role.
	Delete(ctx context.Context, id int) error
	// BatchDelete deletes multiple roles atomically.
	BatchDelete(ctx context.Context, ids []int) error
	// UpdateStatus updates role status.
	UpdateStatus(ctx context.Context, id int, status int) error
}

// RoleUserAssignmentService defines role-to-user assignment operations.
type RoleUserAssignmentService interface {
	// GetUsers queries users assigned to a role.
	GetUsers(ctx context.Context, in GetUsersInput) (*GetUsersOutput, error)
	// AssignUsers assigns users to a role.
	AssignUsers(ctx context.Context, roleId int, userIds []int) error
	// UnassignUser removes user from a role.
	UnassignUser(ctx context.Context, roleId int, userId int) error
	// UnassignUsers removes multiple users from a role.
	UnassignUsers(ctx context.Context, roleId int, userIds []int) error
}

// UserRoleLookupService defines read-only user role lookup operations.
type UserRoleLookupService interface {
	// GetUserRoleIds returns role IDs for a user.
	GetUserRoleIds(ctx context.Context, userId int) ([]int, error)
	// GetUserRoles returns role entities for a user.
	GetUserRoles(ctx context.Context, userId int) ([]*entity.SysRole, error)
	// GetUserRoleNames returns role names for a user.
	GetUserRoleNames(ctx context.Context, userId int) ([]string, error)
}

// UserPermissionLookupService defines role-derived menu and permission lookup operations.
type UserPermissionLookupService interface {
	// GetUserMenuIds returns menu IDs accessible by a user through their roles.
	GetUserMenuIds(ctx context.Context, userId int) ([]int, error)
	// GetUserPermissions returns effective menu and button permission strings for a user.
	GetUserPermissions(ctx context.Context, userId int) ([]string, error)
	// IsSuperAdmin checks whether the user is the built-in admin account.
	IsSuperAdmin(ctx context.Context, userId int) bool
}

// RoleAccessSnapshotService defines token access snapshot and topology invalidation operations.
type RoleAccessSnapshotService interface {
	// PrimeTokenAccessContext preloads the access context cache for one freshly issued login token.
	PrimeTokenAccessContext(
		ctx context.Context,
		tokenID string,
		userID int,
	) (*UserAccessContext, error)
	// InvalidateTokenAccessContext removes the cached access context bound to one token.
	InvalidateTokenAccessContext(ctx context.Context, tokenID string)
	// InvalidateUserAccessContexts removes all cached access contexts bound to one user.
	InvalidateUserAccessContexts(ctx context.Context, userID int)
	// MarkAccessTopologyChanged bumps the shared permission topology revision and clears local token caches.
	MarkAccessTopologyChanged(ctx context.Context) error
	// NotifyAccessTopologyChanged best-effort refreshes the shared permission topology revision.
	NotifyAccessTopologyChanged(ctx context.Context)
	// SyncAccessTopologyRevision synchronizes the process-local permission
	// topology revision and evicts stale token snapshots after cross-node changes.
	SyncAccessTopologyRevision(ctx context.Context) error
	// GetUserAccessContext loads the user's roles, menus, and permissions with token-aware caching when available.
	GetUserAccessContext(ctx context.Context, userId int) (*UserAccessContext, error)
	// GetUserDataScopeSnapshot returns the user's effective role data-scope from the cached access snapshot.
	GetUserDataScopeSnapshot(ctx context.Context, userId int) (*datascope.AccessSnapshot, error)
	// SetDataScopeService wires the shared data-scope service used by role user operations.
	SetDataScopeService(scopeSvc datascope.Service)
}

// Service defines the full role service contract by composing feature-scoped contracts.
type Service interface {
	RoleQueryService
	RoleMutationService
	RoleUserAssignmentService
	UserRoleLookupService
	UserPermissionLookupService
	RoleAccessSnapshotService
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc          bizctx.Service
	configSvc          config.Service
	i18nSvc            roleI18nTranslator
	permissionFilter   PermissionMenuFilter
	orgCapabilityState OrganizationCapabilityState
	orgCapSvc          orgcapsvc.Service
	tenantSvc          tenantcapsvc.Service
	accessRevisionCtrl accessRevisionController
	scopeSvc           datascope.Service
}

// New creates and returns a new role service from explicit runtime-owned dependencies.
func New(permissionFilter PermissionMenuFilter, bizCtxSvc bizctx.Service, configSvc config.Service, i18nSvc roleI18nTranslator, orgCapabilityState OrganizationCapabilityState, orgCapSvc orgcapsvc.Service, tenantSvc tenantcapsvc.Service) Service {
	if permissionFilter == nil {
		permissionFilter = noopPermissionMenuFilter{}
	}
	if orgCapabilityState == nil {
		orgCapabilityState = organizationCapabilityStateFromPermissionFilter(permissionFilter)
	}
	svc := &serviceImpl{
		bizCtxSvc:          bizCtxSvc,
		configSvc:          configSvc,
		i18nSvc:            i18nSvc,
		permissionFilter:   permissionFilter,
		orgCapabilityState: orgCapabilityState,
		orgCapSvc:          orgCapSvc,
		tenantSvc:          tenantSvc,
		accessRevisionCtrl: newCacheCoordAccessRevisionController(
			configSvc.IsClusterEnabled(context.Background()),
		),
	}
	return svc
}

// roleI18nTranslator defines the narrow translation capability role needs.
type roleI18nTranslator interface {
	// Translate returns one runtime translation key with caller-provided fallback text.
	Translate(ctx context.Context, key string, fallback string) string
}

// noopPermissionMenuFilter keeps permission menus unchanged when no external
// plugin-aware filter is injected into the role service.
type noopPermissionMenuFilter struct{}

// pluginBackedOrganizationCapabilityState derives organization capability from
// both plugin enablement state and the registered orgcap provider.
type pluginBackedOrganizationCapabilityState struct {
	pluginState pluginEnablementState
}

// ListInput defines filters and pagination for role list queries.
type ListInput struct {
	Name   string
	Key    string
	Status *int
	Page   int
	Size   int
}

// ListOutput defines the paged role list result.
type ListOutput struct {
	List  []*RoleItem // Role list
	Total int         // Total count
}

// RoleItem represents a role in the list response.
type RoleItem struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Key       string `json:"key"`
	Sort      int    `json:"sort"`
	DataScope int    `json:"dataScope"`
	Status    int    `json:"status"`
	Remark    string `json:"remark"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// GetDetailOutput defines output for GetDetail function.
type GetDetailOutput struct {
	Role    *entity.SysRole
	MenuIds []int
}

// CreateInput defines input for Create function.
type CreateInput struct {
	Name      string
	Key       string
	Sort      int
	DataScope int
	Status    int
	Remark    string
	MenuIds   []int
}

// UpdateInput defines input for Update function.
type UpdateInput struct {
	Id        int
	Name      string
	Key       string
	Sort      *int
	DataScope *int
	Status    *int
	Remark    *string
	MenuIds   []int
}

// roleOwnership describes the tenant boundary attached to one role and its
// relation rows.
type roleOwnership struct {
	TenantID int
}

// OptionItem represents a role option.
type OptionItem struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

// RoleUserItem represents a user assigned to a role.
type RoleUserItem struct {
	Id        int    `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Status    int    `json:"status"`
	CreatedAt string `json:"createdAt"`
}

// GetUsersInput defines input for GetUsers function.
type GetUsersInput struct {
	RoleId   int
	Username string
	Phone    string
	Status   *int
	Page     int
	Size     int
}

// GetUsersOutput defines output for GetUsers function.
type GetUsersOutput struct {
	List  []*RoleUserItem
	Total int
}
