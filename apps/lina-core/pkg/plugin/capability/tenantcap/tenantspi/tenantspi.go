// Package tenantspi owns source-plugin provider SPI and host-internal tenant
// seams. It may use GoFrame request and query-builder types; the parent
// tenantcap package remains the ordinary consumer contract.
package tenantspi

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	internalregistry "lina-core/pkg/plugin/capability/internal/capabilityregistry"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/usercap"
)

// ResolverName identifies one tenant resolver in the configured responsibility chain.
type ResolverName string

// Resolver resolves tenant identity from one HTTP request.
//
// Resolver 定义单个 HTTP 请求租户身份解析器，适用于按请求头、域名、路径、Token 或其他策略组成责任链来解析当前租户。
type Resolver interface {
	// Name returns the stable resolver name used by configuration.
	//
	// Name 返回解析器的稳定名称，适用于配置责任链顺序、诊断解析来源和治理检查。
	Name() ResolverName
	// Resolve attempts to resolve tenant identity for the current request.
	//
	// Resolve 尝试从当前 HTTP 请求解析租户身份，适用于租户中间件在请求进入业务处理前写入业务上下文。
	Resolve(ctx context.Context, r *ghttp.Request) (*tenantcap.ResolverResult, error)
}

// ProviderEnv carries the explicit host services a tenant provider adapter may
// use during lazy construction.
type ProviderEnv struct {
	// PluginID is the tenant provider plugin being constructed.
	PluginID string
	// BizCtx exposes the current request business context without host internals.
	BizCtx bizctxcap.Service
	// PluginLifecycle exposes governed plugin lifecycle hooks needed by tenant-owned plugin policy.
	PluginLifecycle plugincap.LifecycleService
	// Users resolves host-owned user projections without exposing sys_user storage.
	Users usercap.Service
	// Plugins resolves host-owned plugin governance projections.
	Plugins plugincap.Service
	// PluginAdmin executes governed host plugin management commands.
	PluginAdmin plugincap.AdminService
}

// ProviderRuntime defines the narrow plugin state and environment capability
// required by tenantspi to use declared providers.
//
// ProviderRuntime 定义 tenantspi 在延迟创建租户能力提供方时所需的最小宿主运行时入口，适用于判断租户插件是否可服务，并为 provider 构造受治理的宿主环境。
type ProviderRuntime interface {
	// IsProviderEnabled reports whether pluginID may serve framework provider calls.
	//
	// IsProviderEnabled 判断指定插件是否允许承接框架租户能力调用，通常用于能力服务在调用 provider 前确认插件启用状态和运行状态。
	IsProviderEnabled(ctx context.Context, pluginID string) bool
	// TenantProviderEnv returns typed, plugin-scoped construction inputs for one provider plugin.
	//
	// TenantProviderEnv 返回指定租户插件的类型化构造环境，适用于 provider 工厂获取业务上下文、插件生命周期等宿主发布能力。
	TenantProviderEnv(pluginID string) ProviderEnv
}

// Provider defines the multi-tenancy capability implemented by plugins.
//
// Provider 定义多租户能力插件必须实现的基础提供方契约，适用于 linapro-tenant-core 等插件向宿主提供租户解析、租户可见性校验和租户切换能力。
type Provider interface {
	// ResolveTenant resolves tenant identity for one HTTP request.
	//
	// ResolveTenant 从单个 HTTP 请求解析当前租户身份，适用于宿主租户中间件建立请求级租户上下文。
	ResolveTenant(ctx context.Context, r *ghttp.Request) (*tenantcap.ResolverResult, error)
	// ValidateUserInTenant verifies that a user can access a tenant.
	//
	// ValidateUserInTenant 校验指定用户是否可访问指定租户，适用于登录、Token 刷新、租户切换和跨租户访问防护。
	ValidateUserInTenant(ctx context.Context, userID int, tenantID tenantcap.TenantID) error
	// ListUserTenants returns the active tenants visible to one user.
	//
	// ListUserTenants 返回指定用户可见的活跃租户列表，适用于前端租户切换器、会话投影和权限上下文装配。
	ListUserTenants(ctx context.Context, userID int) ([]tenantcap.TenantInfo, error)
	// SwitchTenant validates a tenant switch before token re-issue.
	//
	// SwitchTenant 在重新签发 Token 前校验租户切换是否合法，适用于用户主动切换当前工作租户的流程。
	SwitchTenant(ctx context.Context, userID int, target tenantcap.TenantID) error
}

// TenantProjectionProvider optionally exposes read-only tenant projections.
//
// TenantProjectionProvider 定义租户 provider 可选实现的租户只读投影能力，适用于普通插件读取租户候选、批量租户详情和当前租户详情。
type TenantProjectionProvider interface {
	// CurrentTenantInfo returns one tenant projection visible in the current context.
	//
	// CurrentTenantInfo 返回指定租户投影，适用于当前租户详情和批量租户详情装配。
	CurrentTenantInfo(ctx context.Context, tenantID tenantcap.TenantID) (*tenantcap.TenantInfo, error)
	// BatchGetTenants returns visible tenant projections and opaque missing IDs.
	//
	// BatchGetTenants 批量返回可见租户投影，不存在、不可见和租户外目标进入缺失集合。
	BatchGetTenants(ctx context.Context, tenantIDs []tenantcap.TenantID) (*capmodel.BatchResult[*tenantcap.TenantInfo, tenantcap.TenantID], error)
	// SearchTenants returns bounded tenant candidates visible to the caller.
	//
	// SearchTenants 返回分页租户候选投影，适用于插件筛选和关系选择。
	SearchTenants(ctx context.Context, input tenantcap.SearchInput) (*capmodel.PageResult[*tenantcap.TenantInfo], error)
	// EnsureTenantsVisible validates that every tenant identifier is visible.
	//
	// EnsureTenantsVisible 批量校验租户可见性，任一目标不可见时整体拒绝。
	EnsureTenantsVisible(ctx context.Context, tenantIDs []tenantcap.TenantID) error
}

// RequestResolver defines host-internal HTTP tenant resolution. It is kept out
// of tenantcap.Service because ordinary plugin consumers must not receive
// request-object based resolver hooks through capability services.
//
// RequestResolver 定义宿主内部 HTTP 租户解析能力，适用于中间件从请求对象建立租户上下文，不通过普通插件能力目录暴露。
type RequestResolver interface {
	// Available reports whether tenant resolution should use provider-backed logic.
	//
	// Available 判断当前是否应使用 provider 支持的租户解析逻辑，适用于租户中间件在单租户或插件禁用时走平台默认路径。
	Available(ctx context.Context) bool
	// ResolveTenant delegates HTTP tenant resolution to the provider when enabled.
	//
	// ResolveTenant 将 HTTP 请求租户解析委托给可用 provider，适用于宿主请求链路在认证和权限校验前写入租户上下文。
	ResolveTenant(ctx context.Context, r *ghttp.Request) (*tenantcap.ResolverResult, error)
}

// ScopeService defines host-internal tenant query-scope operations.
//
// ScopeService 定义宿主内部租户查询范围能力，适用于数据库查询层注入租户过滤、用户租户范围和平台租户筛选，不面向普通插件消费。
type ScopeService interface {
	// Available reports whether tenant-scoped database filtering can run.
	//
	// Available 判断租户数据库过滤是否可用，适用于调用方在租户插件禁用时降级到平台租户或跳过租户范围。
	Available(ctx context.Context) bool
	// Apply injects tenant filtering into a model when multi-tenancy is enabled.
	//
	// Apply 向查询模型注入当前租户过滤条件，适用于租户隔离表的列表、详情、导出和聚合查询。
	Apply(ctx context.Context, model *gdb.Model, tenantColumn string) (*gdb.Model, error)
	// ApplyUserTenantScope constrains user rows by active current-tenant membership.
	//
	// ApplyUserTenantScope 按当前租户的有效成员关系约束用户行，适用于用户列表、授权候选和用户相关批量查询。
	ApplyUserTenantScope(ctx context.Context, model *gdb.Model, userIDColumn string) (*gdb.Model, bool, error)
	// ApplyUserTenantFilter constrains platform user-list rows to a requested tenant.
	//
	// ApplyUserTenantFilter 将平台用户列表约束到指定租户，适用于平台视角按租户筛选用户且保持数据库侧过滤分页。
	ApplyUserTenantFilter(ctx context.Context, model *gdb.Model, userIDColumn string, tenantID tenantcap.TenantID) (*gdb.Model, bool, error)
}

// UserMembershipService defines host-internal user-to-tenant membership projections and mutations.
//
// UserMembershipService 定义宿主内部用户租户成员关系投影和写入能力，适用于用户管理维护租户归属、批量校验和列表投影，不通过普通插件能力暴露。
type UserMembershipService interface {
	// ListUserTenantProjections returns tenant ownership labels for visible users.
	//
	// ListUserTenantProjections 批量返回用户租户归属投影，适用于用户列表、详情批量装配和导出中展示租户名称，避免逐用户查询。
	ListUserTenantProjections(ctx context.Context, userIDs []int) (map[int]*tenantcap.UserTenantProjection, error)
	// ResolveUserTenantAssignment validates requested memberships and returns a host write plan.
	//
	// ResolveUserTenantAssignment 校验请求的用户租户归属并生成宿主写入计划，适用于用户创建和更新前统一计算主租户与成员关系。
	ResolveUserTenantAssignment(ctx context.Context, requested []tenantcap.TenantID, mode tenantcap.UserTenantAssignmentMode) (*tenantcap.UserTenantAssignmentPlan, error)
	// ReplaceUserTenantAssignments rewrites one user's active tenant ownership rows.
	//
	// ReplaceUserTenantAssignments 重写指定用户的有效租户归属记录，适用于用户保存流程按已验证计划同步租户成员关系。
	ReplaceUserTenantAssignments(ctx context.Context, userID int, plan *tenantcap.UserTenantAssignmentPlan) error
	// EnsureUsersInTenant verifies every user has active membership in the tenant.
	//
	// EnsureUsersInTenant 校验一组用户是否都属于指定租户，适用于批量授权、批量操作和跨租户写入前的边界检查。
	EnsureUsersInTenant(ctx context.Context, userIDs []int, tenantID tenantcap.TenantID) error
}

// PluginProvisioningService defines host-internal tenant plugin provisioning.
//
// PluginProvisioningService 定义宿主内部租户插件自动供给能力，适用于启动或生命周期治理阶段为租户启用默认插件，不属于普通插件消费面。
type PluginProvisioningService interface {
	// ProvisionAutoEnabledTenantPlugins provisions default tenant plugins when the provider supports it.
	//
	// ProvisionAutoEnabledTenantPlugins 在 provider 支持时供给默认自动启用的租户插件，适用于 HTTP 启动后源码插件提供方已注册的治理流程。
	ProvisionAutoEnabledTenantPlugins(ctx context.Context) error
}

// StartupConsistencyService defines host-internal tenant startup validation.
//
// StartupConsistencyService 定义宿主内部租户启动一致性校验能力，适用于 HTTP 服务对外提供前检查用户成员关系和租户治理状态。
type StartupConsistencyService interface {
	// ValidateUserMembershipStartupConsistency returns startup consistency failures detected by the provider.
	//
	// ValidateUserMembershipStartupConsistency 返回 provider 检测到的用户成员关系启动一致性问题，适用于启动期失败前置和治理诊断。
	ValidateUserMembershipStartupConsistency(ctx context.Context) ([]string, error)
}

// UserMembershipProvider optionally exposes user-to-tenant ownership behavior.
//
// UserMembershipProvider 定义租户 provider 可选实现的用户成员关系能力，适用于支持用户租户归属、用户列表过滤和启动一致性校验的租户插件。
type UserMembershipProvider interface {
	// ListUserTenants returns the active tenants visible to one user.
	//
	// ListUserTenants 返回指定用户可见的活跃租户列表，适用于普通租户能力读取和用户上下文装配。
	ListUserTenants(ctx context.Context, userID int) ([]tenantcap.TenantInfo, error)
	// BatchListUserTenants returns active tenant memberships for visible users.
	//
	// BatchListUserTenants 批量返回用户可访问租户列表，适用于列表和批量详情装配，避免逐用户 provider 查询。
	BatchListUserTenants(ctx context.Context, userIDs []int) (map[int][]tenantcap.TenantInfo, error)
	// ApplyUserTenantScope constrains user-owned rows to the current request tenant.
	//
	// ApplyUserTenantScope 将用户相关查询限制到当前请求租户成员范围，适用于用户列表和用户关联资源查询。
	ApplyUserTenantScope(ctx context.Context, model *gdb.Model, userIDColumn string) (*gdb.Model, bool, error)
	// ApplyUserTenantFilter constrains platform user-list rows to a requested tenant.
	//
	// ApplyUserTenantFilter 将平台用户列表限制到指定租户，适用于平台管理员按租户查看或维护用户。
	ApplyUserTenantFilter(ctx context.Context, model *gdb.Model, userIDColumn string, tenantID tenantcap.TenantID) (*gdb.Model, bool, error)
	// ListUserTenantProjections returns tenant ownership labels for visible users.
	//
	// ListUserTenantProjections 批量返回用户租户归属标签，适用于用户列表、详情批量和导出投影装配。
	ListUserTenantProjections(ctx context.Context, userIDs []int) (map[int]*tenantcap.UserTenantProjection, error)
	// ResolveUserTenantAssignment validates requested memberships and returns a host write plan.
	//
	// ResolveUserTenantAssignment 校验请求成员关系并返回宿主写入计划，适用于用户创建和更新流程。
	ResolveUserTenantAssignment(ctx context.Context, requested []tenantcap.TenantID, mode tenantcap.UserTenantAssignmentMode) (*tenantcap.UserTenantAssignmentPlan, error)
	// ReplaceUserTenantAssignments rewrites one user's active tenant ownership rows.
	//
	// ReplaceUserTenantAssignments 重写指定用户的有效租户归属，适用于宿主用户保存后的成员关系同步。
	ReplaceUserTenantAssignments(ctx context.Context, userID int, plan *tenantcap.UserTenantAssignmentPlan) error
	// EnsureUsersInTenant verifies every user has active membership in the tenant.
	//
	// EnsureUsersInTenant 校验多个用户均属于指定租户，适用于批量操作和跨租户边界校验。
	EnsureUsersInTenant(ctx context.Context, userIDs []int, tenantID tenantcap.TenantID) error
	// ValidateStartupConsistency returns user-membership startup consistency failures.
	//
	// ValidateStartupConsistency 返回用户成员关系启动一致性问题，适用于租户插件在宿主启动阶段暴露治理失败原因。
	ValidateStartupConsistency(ctx context.Context) ([]string, error)
}

// PluginProvisioningProvider optionally exposes tenant-plugin provisioning behavior.
//
// PluginProvisioningProvider 定义租户 provider 可选实现的插件供给能力，适用于租户插件负责按租户策略自动启用或供给默认插件资源。
type PluginProvisioningProvider interface {
	PluginProvisioningService
}

// RuntimeService is the host-owned tenant adapter that combines the ordinary
// consumer surface with host-internal resolver, scope, membership, provisioning,
// and startup consistency seams.
//
// RuntimeService 是宿主启动期持有的租户能力聚合接口，适用于显式注入普通消费、请求解析、数据库范围、用户成员关系、插件供给和启动一致性等窄接口。
type RuntimeService interface {
	tenantcap.Service
	RequestResolver
	ScopeService
	UserMembershipService
	PluginProvisioningService
	StartupConsistencyService
}

// ProviderFactory creates one tenant provider from an explicit, typed
// construction environment during lazy capability use.
type ProviderFactory func(ctx context.Context, env ProviderEnv) (Provider, error)

// Manager owns tenant provider declarations and lazy provider instances.
type Manager struct {
	registry *internalregistry.Manager[ProviderEnv]
}

// NewManager creates an empty tenant provider manager.
func NewManager() *Manager {
	return &Manager{registry: internalregistry.NewManager[ProviderEnv]()}
}

// RegisterFactory records one plugin-provided tenant capability factory.
func (m *Manager) RegisterFactory(pluginID string, factory ProviderFactory) error {
	return m.registry.RegisterFactory(
		tenantcap.CapabilityTenantV1,
		pluginID,
		func(ctx context.Context, env ProviderEnv) (any, error) {
			return factory(ctx, env)
		},
	)
}

// serviceImpl delegates tenant calls to the active provider and returns
// platform-safe fallback values when no provider is usable.
type serviceImpl struct {
	manager   *Manager
	runtime   ProviderRuntime
	bizCtxSvc bizctxcap.Service
}

// Ensure serviceImpl implements Service.
var _ tenantcap.Service = (*serviceImpl)(nil)
var _ RequestResolver = (*serviceImpl)(nil)
var _ ScopeService = (*serviceImpl)(nil)
var _ UserMembershipService = (*serviceImpl)(nil)
var _ PluginProvisioningService = (*serviceImpl)(nil)
var _ StartupConsistencyService = (*serviceImpl)(nil)
var _ RuntimeService = (*serviceImpl)(nil)

// New creates an optional tenant capability service from explicit runtime-owned dependencies.
func New(manager *Manager, runtime ProviderRuntime, bizCtxSvc bizctxcap.Service) RuntimeService {
	if manager == nil {
		manager = NewManager()
	}
	if runtime == nil {
		runtime = noopProviderRuntime{}
	}
	return &serviceImpl{
		manager:   manager,
		runtime:   runtime,
		bizCtxSvc: bizCtxSvc,
	}
}

// noopProviderRuntime reports all plugins as disabled when tenantspi is
// constructed without an explicit provider runtime.
type noopProviderRuntime struct{}

// ProviderStatuses returns all tenant provider states.
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
