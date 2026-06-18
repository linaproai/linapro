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
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"lina-core/pkg/plugin/capability/usercap"
)

// ProviderEnv carries the explicit host services an organization provider
// adapter may use during lazy construction.
type ProviderEnv struct {
	// PluginID is the organization provider plugin being constructed.
	PluginID string
	// TenantFilter constrains provider-owned plugin tables by the current tenant.
	TenantFilter tenantspi.PluginTableFilterService
	// Users resolves host-owned user projections without exposing sys_user
	// storage to the organization provider plugin.
	Users usercap.Service
}

// ProviderRuntime defines the narrow plugin state and environment capability
// required by orgspi to use declared providers.
//
// ProviderRuntime 定义 orgspi 在延迟创建组织能力提供方时所需的最小宿主运行时入口，适用于从插件治理状态判断提供方是否可用，
// 并为指定组织插件构造受治理的宿主环境。
type ProviderRuntime interface {
	// IsProviderEnabled reports whether pluginID may serve framework provider calls.
	//
	// IsProviderEnabled 判断指定插件是否允许承接框架组织能力调用，通常用于能力服务在调用 provider 前确认插件已启用且处于可服务状态。
	IsProviderEnabled(ctx context.Context, pluginID string) bool
	// OrgProviderEnv returns typed, plugin-scoped construction inputs for one provider plugin.
	//
	// OrgProviderEnv 返回指定组织插件的类型化构造环境，通常用于 provider 工厂获取租户过滤等宿主发布能力，同时避免插件直接依赖宿主内部实现。
	OrgProviderEnv(pluginID string) ProviderEnv
}

// Provider defines the organization capability implemented by provider plugins.
//
// Provider 定义组织能力插件必须实现的完整提供方契约，适用于 linapro-org-core 等插件向宿主提供部门、岗位、组织数据权限和用户组织关系维护能力。
type Provider interface {
	// ListUserDeptAssignments returns user -> department projections for the provided users.
	//
	// ListUserDeptAssignments 批量返回用户到部门的归属投影，适用于用户列表、详情批量装配和会话投影等需要避免逐用户查询的场景。
	ListUserDeptAssignments(ctx context.Context, userIDs []int) (map[int]*orgcap.UserDeptAssignment, error)
	// GetUserDeptInfo returns one user's department projection.
	//
	// GetUserDeptInfo 返回单个用户的部门标识和名称，适用于用户详情、登录会话补充信息等单用户读取场景。
	GetUserDeptInfo(ctx context.Context, userID int) (int, string, error)
	// GetUserDeptIDs returns one user's department identifier list.
	//
	// GetUserDeptIDs 返回单个用户所属部门标识集合，适用于部门数据权限判定或组织关系读取场景。
	GetUserDeptIDs(ctx context.Context, userID int) ([]int, error)
	// ApplyUserDeptScope injects a database-side department-scope constraint for rows owned by userIDColumn.
	//
	// ApplyUserDeptScope 向宿主数据库查询注入部门数据范围约束，适用于列表、导出等需要在数据库侧过滤用户可见数据的场景。
	ApplyUserDeptScope(ctx context.Context, model *gdb.Model, userIDColumn string, currentUserID int) (*gdb.Model, bool, error)
	// BuildUserDeptScopeExists builds a database-side department-scope EXISTS subquery.
	//
	// BuildUserDeptScopeExists 构建部门数据范围 EXISTS 子查询，适用于调用方需要把组织范围与其他条件组合成复杂查询的场景。
	BuildUserDeptScopeExists(ctx context.Context, userIDColumn string, currentUserID int) (*gdb.Model, bool, error)
	// ApplyUserDeptFilter constrains user rows to a requested department subtree without materializing user IDs.
	//
	// ApplyUserDeptFilter 将用户查询限制在指定部门子树内，适用于用户列表按部门筛选且需要避免先加载大量用户标识的场景。
	ApplyUserDeptFilter(ctx context.Context, model *gdb.Model, userIDColumn string, deptID int) (*gdb.Model, bool, error)
	// ApplyUserDeptUnassignedFilter constrains user rows to records without department assignments.
	//
	// ApplyUserDeptUnassignedFilter 将用户查询限制为未分配部门的用户，适用于用户管理工作台的未分配部门筛选场景。
	ApplyUserDeptUnassignedFilter(ctx context.Context, model *gdb.Model, userIDColumn string) (*gdb.Model, bool, error)
	// GetUserPostIDs returns one user's post association list.
	//
	// GetUserPostIDs 返回单个用户关联的岗位标识集合，适用于用户详情、编辑回显和组织关系维护场景。
	GetUserPostIDs(ctx context.Context, userID int) ([]int, error)
	// BatchGetUserOrgProfiles returns stable organization profiles for provided users.
	//
	// BatchGetUserOrgProfiles 批量返回用户组织档案，适用于普通插件批量装配部门和岗位投影，不暴露 provider 内部表结构。
	BatchGetUserOrgProfiles(ctx context.Context, userIDs []int) (*capmodel.BatchResult[*orgcap.UserOrgProfile, int], error)
	// ListDeptTree returns a bounded ordinary department tree projection.
	//
	// ListDeptTree 返回有界部门树投影，适用于跨插件候选展示，不返回工作台专用扩展字段或查询构造器。
	ListDeptTree(ctx context.Context, input orgcap.DeptTreeInput) (*orgcap.DeptTreeResult, error)
	// SearchDepartments returns bounded department candidates.
	//
	// SearchDepartments 返回分页部门候选投影，适用于插件表单和筛选候选。
	SearchDepartments(ctx context.Context, input orgcap.DeptSearchInput) (*capmodel.PageResult[*orgcap.DeptProjection], error)
	// ListPostOptionsPage returns bounded post candidates.
	//
	// ListPostOptionsPage 返回分页岗位候选投影，适用于插件表单和筛选候选。
	ListPostOptionsPage(ctx context.Context, input orgcap.PostOptionsInput) (*capmodel.PageResult[*orgcap.PostOption], error)
	// EnsureDepartmentsVisible verifies all department identifiers are visible.
	//
	// EnsureDepartmentsVisible 校验部门引用可见性，任一目标不可见时整体拒绝。
	EnsureDepartmentsVisible(ctx context.Context, deptIDs []int) error
	// EnsurePostsVisible verifies all post identifiers are visible.
	//
	// EnsurePostsVisible 校验岗位引用可见性，任一目标不可见时整体拒绝。
	EnsurePostsVisible(ctx context.Context, postIDs []int) error
	// ReplaceUserAssignments rewrites one user's department and post associations.
	//
	// ReplaceUserAssignments 重写单个用户的部门和岗位关系，适用于用户创建、用户更新等需要同步组织归属的写入场景。
	ReplaceUserAssignments(ctx context.Context, userID int, deptID *int, postIDs []int) error
	// CleanupUserAssignments deletes one user's optional organization associations.
	//
	// CleanupUserAssignments 清理单个用户的可选组织关联，适用于用户删除或组织能力禁用后的关系释放场景。
	CleanupUserAssignments(ctx context.Context, userID int) error
	// UserDeptTree returns the optional department tree used by host user management.
	//
	// UserDeptTree 返回宿主用户管理需要的部门树投影，适用于部门筛选、树选择器和用户管理工作台的组织视图装配场景。
	UserDeptTree(ctx context.Context) ([]*orgcap.DeptTreeNode, error)
	// ListPostOptions returns selectable post options for one department subtree.
	//
	// ListPostOptions 返回指定部门子树下可选择的岗位选项，适用于用户创建或编辑表单中的岗位候选项装配场景。
	ListPostOptions(ctx context.Context, deptID *int) ([]*orgcap.PostOption, error)
}

// ScopeService defines host-internal organization data-scope operations.
//
// ScopeService 定义宿主内部使用的组织数据范围能力，适用于角色权限、用户列表和业务数据列表在数据库查询层接入部门范围过滤，不面向普通插件消费。
type ScopeService interface {
	// Available reports whether organization-backed data-scope filtering can run.
	//
	// Available 判断组织数据范围过滤是否可用，适用于调用方在组织插件禁用时自动降级到默认数据范围策略。
	Available(ctx context.Context) bool
	// ApplyUserDeptScope injects a database-side department-scope constraint for rows owned by userIDColumn.
	//
	// ApplyUserDeptScope 向查询模型注入当前用户可见部门范围约束，适用于列表、导出和聚合统计等需要数据库侧过滤的宿主接口。
	ApplyUserDeptScope(ctx context.Context, model *gdb.Model, userIDColumn string, currentUserID int) (*gdb.Model, bool, error)
	// BuildUserDeptScopeExists builds a database-side department-scope EXISTS subquery.
	//
	// BuildUserDeptScopeExists 构建当前用户可见部门范围 EXISTS 子查询，适用于调用方需要自行组合 OR 条件或复合权限条件的查询路径。
	BuildUserDeptScopeExists(ctx context.Context, userIDColumn string, currentUserID int) (*gdb.Model, bool, error)
	// ApplyUserDeptFilter constrains user rows to a requested department subtree without materializing user IDs.
	//
	// ApplyUserDeptFilter 将用户查询限制到指定部门子树，适用于宿主用户管理按部门筛选并保持数据库侧分页过滤的场景。
	ApplyUserDeptFilter(ctx context.Context, model *gdb.Model, userIDColumn string, deptID int) (*gdb.Model, bool, error)
	// ApplyUserDeptUnassignedFilter constrains user rows to records without department assignments.
	//
	// ApplyUserDeptUnassignedFilter 将用户查询限制为未分配部门记录，适用于宿主用户管理的未分配用户筛选场景。
	ApplyUserDeptUnassignedFilter(ctx context.Context, model *gdb.Model, userIDColumn string) (*gdb.Model, bool, error)
}

// AssignmentService defines host-internal organization assignment mutations.
//
// AssignmentService 定义宿主内部使用的组织归属写入能力，适用于用户创建、用户更新、用户删除等宿主流程维护部门和岗位关系，
// 不通过普通插件能力目录暴露。
type AssignmentService interface {
	// ReplaceUserAssignments rewrites one user's department and post associations.
	//
	// ReplaceUserAssignments 重写指定用户的部门和岗位关系，适用于宿主用户保存流程在同一业务操作中同步组织归属。
	ReplaceUserAssignments(ctx context.Context, userID int, deptID *int, postIDs []int) error
	// CleanupUserAssignments deletes one user's optional organization associations.
	//
	// CleanupUserAssignments 删除指定用户的可选组织关系，适用于宿主删除用户或回收用户组织归属的清理路径。
	CleanupUserAssignments(ctx context.Context, userID int) error
}

// ProjectionService defines host-internal organization projections used by the built-in user-management workspace.
//
// ProjectionService 定义宿主内建用户管理工作台使用的组织展示投影，适用于部门树、岗位下拉等页面适配数据装配，
// 不作为普通插件消费的通用组织领域契约。
type ProjectionService interface {
	// UserDeptTree returns the optional department tree used by host user management.
	//
	// UserDeptTree 返回用户管理工作台需要的部门树投影，适用于部门筛选器、树选择器和组织视图展示。
	UserDeptTree(ctx context.Context) ([]*orgcap.DeptTreeNode, error)
	// ListPostOptions returns selectable post options for one department subtree.
	//
	// ListPostOptions 返回指定部门子树下的岗位候选项，适用于用户创建和编辑表单中的岗位选择控件。
	ListPostOptions(ctx context.Context, deptID *int) ([]*orgcap.PostOption, error)
}

// RuntimeService is the host-owned organization adapter that combines the
// ordinary consumer surface with host-internal scope, assignment, and
// workspace-projection seams.
//
// RuntimeService 是宿主启动期持有的组织能力聚合接口，适用于在宿主内部显式注入普通消费、数据范围、归属写入和工作台投影等不同窄接口。
type RuntimeService interface {
	orgcap.Service
	ScopeService
	AssignmentService
	ProjectionService
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
	manager *Manager
	runtime ProviderRuntime
}

// Ensure serviceImpl implements Service.
var (
	_ orgcap.Service    = (*serviceImpl)(nil)
	_ ScopeService      = (*serviceImpl)(nil)
	_ AssignmentService = (*serviceImpl)(nil)
	_ ProjectionService = (*serviceImpl)(nil)
	_ RuntimeService    = (*serviceImpl)(nil)
)

// New creates an organization capability service. A nil provider runtime
// treats the capability as disabled while keeping all fallback calls safe.
func New(manager *Manager, runtime ProviderRuntime) RuntimeService {
	if manager == nil {
		manager = NewManager()
	}
	if runtime == nil {
		runtime = noopProviderRuntime{}
	}
	return &serviceImpl{manager: manager, runtime: runtime}
}

// noopProviderRuntime reports all plugins as disabled when orgspi is
// constructed without an explicit provider runtime.
type noopProviderRuntime struct{}

// ProviderStatuses returns all organization provider states.
func (m *Manager) ProviderStatuses(ctx context.Context, runtime ProviderRuntime) []capmodel.ProviderStatus {
	if m == nil || m.registry == nil {
		return nil
	}
	statuses := m.registry.Statuses(ctx, runtime)
	result := make([]capmodel.ProviderStatus, 0, len(statuses))
	for _, status := range statuses {
		result = append(result, convertProviderStatus(status))
	}
	return result
}
