// Package orgspi owns source-plugin provider SPI and host-internal
// organization seams. The parent orgcap package remains the ordinary consumer
// contract and does not expose query builders.
package orgspi

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/plugin/capability/capmodel"
	internalregistry "lina-core/pkg/plugin/capability/internal/capabilityregistry"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/usercap"
)

// Provider defines the minimal organization capability implemented by provider plugins.
type Provider interface {
	DepartmentProvider
	PostProvider
	AssignmentProvider
	ScopeProvider
}

// DepartmentProvider owns department resource operations that cannot be safely
// derived by the host organization adapter.
type DepartmentProvider interface {
	// BatchGetDepartments returns visible department projections and opaque missing IDs.
	BatchGetDepartments(ctx context.Context, deptIDs []int) (*capmodel.BatchResult[*orgcap.DeptInfo, int], error)
	// SearchDepartments returns bounded department candidates.
	SearchDepartments(ctx context.Context, input orgcap.DeptListInput) (*capmodel.PageResult[*orgcap.DeptInfo], error)
	// ListDeptTree returns a bounded ordinary department tree projection.
	ListDeptTree(ctx context.Context, input orgcap.DeptTreeInput) (*orgcap.DeptTreeResult, error)
	// CreateDepartment creates one department through the provider owner.
	CreateDepartment(ctx context.Context, input orgcap.DeptCreateInput) (int, error)
	// UpdateDepartment updates one department through the provider owner.
	UpdateDepartment(ctx context.Context, input orgcap.DeptUpdateInput) error
	// DeleteDepartment deletes one department through the provider owner.
	DeleteDepartment(ctx context.Context, deptID int) error
}

// PostProvider owns post resource operations that cannot be safely derived by
// the host organization adapter.
type PostProvider interface {
	// BatchGetPosts returns visible post projections and opaque missing IDs.
	BatchGetPosts(ctx context.Context, postIDs []int) (*capmodel.BatchResult[*orgcap.PostInfo, int], error)
	// ListPosts returns bounded post projections.
	ListPosts(ctx context.Context, input orgcap.PostListInput) (*capmodel.PageResult[*orgcap.PostInfo], error)
	// CreatePost creates one post through the provider owner.
	CreatePost(ctx context.Context, input orgcap.PostCreateInput) (int, error)
	// UpdatePost updates one post through the provider owner.
	UpdatePost(ctx context.Context, input orgcap.PostUpdateInput) error
	// DeletePost deletes one post through the provider owner.
	DeletePost(ctx context.Context, postID int) error
}

// AssignmentProvider owns user organization profile writes and batch reads.
type AssignmentProvider interface {
	// BatchGetUserOrgProfiles returns stable organization profiles for provided users.
	BatchGetUserOrgProfiles(ctx context.Context, userIDs []int) (*capmodel.BatchResult[*orgcap.UserOrgProfile, int], error)
	// ReplaceUserAssignments rewrites one user's department and post associations.
	ReplaceUserAssignments(ctx context.Context, userID int, deptID *int, postIDs []int) error
	// CleanupUserAssignments deletes one user's optional organization associations.
	CleanupUserAssignments(ctx context.Context, userID int) error
}

// ScopeProvider owns provider-side SQL predicates for organization data scope.
type ScopeProvider interface {
	// BuildUserDeptScopeExists builds a database-side department-scope EXISTS subquery.
	BuildUserDeptScopeExists(ctx context.Context, userIDColumn string, currentUserID int) (*gdb.Model, bool, error)
	// ApplyUserDeptFilter constrains user rows to a requested department subtree without materializing user IDs.
	ApplyUserDeptFilter(ctx context.Context, model *gdb.Model, userIDColumn string, deptID int) (*gdb.Model, bool, error)
	// ApplyUserDeptUnassignedFilter constrains user rows to records without department assignments.
	ApplyUserDeptUnassignedFilter(ctx context.Context, model *gdb.Model, userIDColumn string) (*gdb.Model, bool, error)
}

// ScopeService defines host-internal organization data-scope operations.
type ScopeService interface {
	// Available reports whether organization-backed data-scope filtering can run.
	Available(ctx context.Context) bool
	// ApplyUserDeptScope injects a database-side department-scope constraint for rows owned by userIDColumn.
	ApplyUserDeptScope(ctx context.Context, model *gdb.Model, userIDColumn string, currentUserID int) (*gdb.Model, bool, error)
	// BuildUserDeptScopeExists builds a database-side department-scope EXISTS subquery.
	BuildUserDeptScopeExists(ctx context.Context, userIDColumn string, currentUserID int) (*gdb.Model, bool, error)
	// ApplyUserDeptFilter constrains user rows to a requested department subtree without materializing user IDs.
	ApplyUserDeptFilter(ctx context.Context, model *gdb.Model, userIDColumn string, deptID int) (*gdb.Model, bool, error)
	// ApplyUserDeptUnassignedFilter constrains user rows to records without department assignments.
	ApplyUserDeptUnassignedFilter(ctx context.Context, model *gdb.Model, userIDColumn string) (*gdb.Model, bool, error)
}

// WorkspaceViewService defines host-internal organization projections used by the built-in user-management workspace.
type WorkspaceViewService interface {
	// UserDeptTree returns the optional department tree used by host user management.
	UserDeptTree(ctx context.Context) ([]*orgcap.DeptTreeNode, error)
	// ListPostOptions returns selectable post options for one department subtree.
	ListPostOptions(ctx context.Context, deptID *int) ([]*orgcap.PostOption, error)
}

// Service defines the host-owned organization adapter. Ordinary organization
// consumption stays on orgcap.Service; host-internal database scope and
// built-in workspace projections are reached through explicit narrow seams.
type Service interface {
	orgcap.Service
	// Scope returns host-internal organization data-scope operations.
	Scope() ScopeService
	// Workspace returns user-management workspace projections.
	Workspace() WorkspaceViewService
}

// ProviderEnv carries the explicit host services an organization provider
// adapter may use during lazy construction.
type ProviderEnv struct {
	// PluginID is the organization provider plugin being constructed.
	PluginID string
	// Tenant provides tenant context and constrains provider-owned plugin tables by the current tenant.
	Tenant tenantcap.Service
	// Users resolves host-owned user projections without exposing sys_user
	// storage to the organization provider plugin.
	Users usercap.Service
}

// ProviderFactory creates one organization provider from an explicit, typed
// construction environment during lazy capability use.
type ProviderFactory func(ctx context.Context, env ProviderEnv) (Provider, error)

// Manager owns organization provider declarations and lazy provider instances.
type Manager struct {
	registry *internalregistry.Manager[ProviderEnv]
}

// NewManager creates an empty organization provider manager.
func NewManager() *Manager {
	return &Manager{registry: internalregistry.NewManager[ProviderEnv]()}
}

// RegisterFactory records one plugin-provided organization capability factory.
func (m *Manager) RegisterFactory(pluginID string, factory ProviderFactory) error {
	return m.registry.RegisterFactory(
		orgcap.CapabilityOrgV1,
		pluginID,
		func(ctx context.Context, env ProviderEnv) (any, error) {
			return factory(ctx, env)
		},
	)
}

// serviceImpl delegates organization calls to the active provider and returns
// neutral fallback values when no provider is usable.
type serviceImpl struct {
	manager    *Manager
	enablement internalregistry.EnablementReader
	envFactory internalregistry.ProviderEnvFactory[ProviderEnv]
}

// Ensure serviceImpl implements Service.
var (
	_ orgcap.Service           = (*serviceImpl)(nil)
	_ orgcap.DepartmentService = (*departmentService)(nil)
	_ orgcap.PostService       = (*postService)(nil)
	_ orgcap.AssignmentService = (*serviceImpl)(nil)
	_ ScopeService             = (*serviceImpl)(nil)
	_ WorkspaceViewService     = (*serviceImpl)(nil)
	_ Service                  = (*serviceImpl)(nil)
)

// New creates an organization capability service. A nil enablement reader
// treats the capability as disabled while keeping all fallback calls safe.
func New(
	manager *Manager,
	enablement internalregistry.EnablementReader,
	envFactory internalregistry.ProviderEnvFactory[ProviderEnv],
) Service {
	if manager == nil {
		manager = NewManager()
	}
	if enablement == nil {
		enablement = noopEnablementReader{}
	}
	if envFactory == nil {
		envFactory = defaultProviderEnv
	}
	return &serviceImpl{manager: manager, enablement: enablement, envFactory: envFactory}
}

// noopEnablementReader reports all provider plugins as disabled.
type noopEnablementReader struct{}

// ProviderStatuses returns all organization provider states.
func (m *Manager) ProviderStatuses(ctx context.Context, enablement internalregistry.EnablementReader) []capmodel.ProviderStatus {
	if m == nil || m.registry == nil {
		return nil
	}
	statuses := m.registry.Statuses(ctx, enablement)
	result := make([]capmodel.ProviderStatus, 0, len(statuses))
	for _, status := range statuses {
		result = append(result, convertProviderStatus(status))
	}
	return result
}

// defaultProviderEnv creates a minimal provider environment when no host
// plugin runtime has been bound.
func defaultProviderEnv(_ context.Context, pluginID string) ProviderEnv {
	return ProviderEnv{PluginID: pluginID}
}
