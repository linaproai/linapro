// Package tenantcap owns the stable tenant capability contract exposed through
// capability. The host owns request-context fallback and filtering behavior;
// provider plugins declare factories while the consumer service lazily creates
// concrete multi-tenant policy when it is used.
package tenantcap

import (
	"context"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	internalregistry "lina-core/pkg/plugin/capability/internal/capabilityregistry"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/usercap"
)

const (
	// CapabilityTenantV1 identifies the versioned tenant framework capability.
	CapabilityTenantV1 = "framework.tenant.v1"
	// ProviderPluginID is the official source-plugin identifier that provides tenant capability.
	ProviderPluginID = "linapro-tenant-core"
)

// TenantID identifies one tenant in the pooled tenancy model.
type TenantID int

const (
	// PlatformTenantID is the platform tenant used by single-tenant mode and platform defaults.
	PlatformTenantID TenantID = 0
	// AllTenantsID is the explicit all-tenant cache invalidation sentinel.
	AllTenantsID TenantID = -1
)

const (
	// PLATFORM is the platform tenant used by single-tenant mode and platform defaults.
	PLATFORM = PlatformTenantID
	// ALL_TENANTS is the explicit all-tenant cache invalidation sentinel.
	ALL_TENANTS = AllTenantsID
)

// ResolverName identifies one tenant resolver in the configured responsibility chain.
type ResolverName string

// TenantInfo describes one host-facing tenant projection.
type TenantInfo struct {
	ID     TenantID // ID is the numeric tenant identifier.
	Code   string   // Code is the stable tenant code.
	Name   string   // Name is the tenant display name.
	Status string   // Status is the provider-owned tenant lifecycle status.
}

// UserTenantProjection describes the host-facing tenant ownership projection for one user row.
type UserTenantProjection struct {
	TenantIDs   []TenantID // TenantIDs are active tenant identifiers.
	TenantNames []string   // TenantNames are active tenant display names.
}

// UserTenantAssignmentPlan carries a validated replacement plan for one user.
type UserTenantAssignmentPlan struct {
	TenantIDs     []TenantID // TenantIDs are the active tenant memberships to persist.
	ShouldReplace bool       // ShouldReplace reports whether the provider should rewrite memberships.
	PrimaryTenant TenantID   // PrimaryTenant is the host sys_user tenant_id value.
}

// UserTenantAssignmentMode identifies the host operation requesting assignment planning.
type UserTenantAssignmentMode string

const (
	// UserTenantAssignmentCreate plans memberships for user creation.
	UserTenantAssignmentCreate UserTenantAssignmentMode = "create"
	// UserTenantAssignmentUpdate plans memberships for user update.
	UserTenantAssignmentUpdate UserTenantAssignmentMode = "update"
)

// ResolverResult is one resolver outcome for the responsibility-chain dispatcher.
type ResolverResult struct {
	TenantID        TenantID // TenantID is the resolved tenant.
	Matched         bool     // Matched reports whether this resolver produced a tenant decision.
	ActingAsTenant  bool     // ActingAsTenant marks platform impersonation of a tenant.
	ActingUserID    int      // ActingUserID is the real user when impersonation is active.
	IsImpersonation bool     // IsImpersonation marks an impersonation token or override.
}

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
	Resolve(ctx context.Context, r *ghttp.Request) (*ResolverResult, error)
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
// required by tenantcap to use declared providers.
//
// ProviderRuntime 定义 tenantcap 在延迟创建租户能力提供方时所需的最小宿主运行时入口，适用于判断租户插件是否可服务，并为 provider 构造受治理的宿主环境。
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
	ResolveTenant(ctx context.Context, r *ghttp.Request) (*ResolverResult, error)
	// ValidateUserInTenant verifies that a user can access a tenant.
	//
	// ValidateUserInTenant 校验指定用户是否可访问指定租户，适用于登录、Token 刷新、租户切换和跨租户访问防护。
	ValidateUserInTenant(ctx context.Context, userID int, tenantID TenantID) error
	// ListUserTenants returns the active tenants visible to one user.
	//
	// ListUserTenants 返回指定用户可见的活跃租户列表，适用于前端租户切换器、会话投影和权限上下文装配。
	ListUserTenants(ctx context.Context, userID int) ([]TenantInfo, error)
	// SwitchTenant validates a tenant switch before token re-issue.
	//
	// SwitchTenant 在重新签发 Token 前校验租户切换是否合法，适用于用户主动切换当前工作租户的流程。
	SwitchTenant(ctx context.Context, userID int, target TenantID) error
}

// Service defines the optional tenant capability consumed by host core services
// and plugins without hard-linking them to a concrete provider implementation.
//
// Service 定义宿主核心服务和普通插件可消费的租户能力，适用于读取当前租户、判断平台绕过、校验租户可见性和列出用户可访问租户。
type Service interface {
	// Available reports whether an active tenant provider is available.
	//
	// Available 判断当前是否存在可用租户能力提供方，适用于调用方决定启用多租户逻辑或降级到平台租户。
	Available(ctx context.Context) bool
	// Status returns the current tenant capability activation state.
	//
	// Status 返回租户能力激活状态，适用于运行时诊断、治理检查和插件能力状态展示。
	Status(ctx context.Context) capmodel.CapabilityStatus
	// Current returns the current request tenant, defaulting to platform when context is unavailable.
	//
	// Current 返回当前请求租户，适用于业务查询、缓存键和权限判断；当上下文缺失或租户能力不可用时返回平台租户。
	Current(ctx context.Context) TenantID
	// PlatformBypass reports whether the current request may bypass tenant filtering.
	//
	// PlatformBypass 判断当前请求是否允许绕过租户过滤，适用于平台管理员、启动治理和平台级数据读取路径。
	PlatformBypass(ctx context.Context) bool
	// EnsureTenantVisible validates that the current user can access tenantID.
	//
	// EnsureTenantVisible 校验当前用户是否可访问指定租户，适用于写入、查询参数校验和跨租户资源访问防护。
	EnsureTenantVisible(ctx context.Context, tenantID TenantID) error
	// ValidateUserInTenant verifies that a user can access a tenant.
	//
	// ValidateUserInTenant 校验指定用户是否可访问指定租户，适用于认证、租户切换和后台治理流程。
	ValidateUserInTenant(ctx context.Context, userID int, tenantID TenantID) error
	// ListUserTenants returns active tenant memberships visible to one user.
	//
	// ListUserTenants 返回指定用户可见的活跃租户列表，适用于登录响应、租户切换候选和用户上下文展示。
	ListUserTenants(ctx context.Context, userID int) ([]TenantInfo, error)
	// SwitchTenant validates a tenant switch before token re-issue.
	//
	// SwitchTenant 校验指定用户切换到目标租户是否合法，适用于租户切换接口在重新签发令牌前执行准入检查。
	SwitchTenant(ctx context.Context, userID int, target TenantID) error
}

// RequestResolver defines host-internal HTTP tenant resolution. It is kept out
// of Service because ordinary plugin consumers must not receive *ghttp.Request
// based resolver hooks through the capability services.
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
	ResolveTenant(ctx context.Context, r *ghttp.Request) (*ResolverResult, error)
}

// ScopeService defines host-internal tenant query-scope operations. It carries
// *gdb.Model mutators and therefore must not be exposed through
// capability.Services.Tenant.
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
	ApplyUserTenantFilter(ctx context.Context, model *gdb.Model, userIDColumn string, tenantID TenantID) (*gdb.Model, bool, error)
}

// UserMembershipService defines host-internal user-to-tenant membership
// projections and mutations. Ordinary plugins receive read-only tenant DTOs
// through Service instead of this write-capable seam.
//
// UserMembershipService 定义宿主内部用户租户成员关系投影和写入能力，适用于用户管理维护租户归属、批量校验和列表投影，不通过普通插件能力暴露。
type UserMembershipService interface {
	// ListUserTenantProjections returns tenant ownership labels for visible users.
	//
	// ListUserTenantProjections 批量返回用户租户归属投影，适用于用户列表、详情批量装配和导出中展示租户名称，避免逐用户查询。
	ListUserTenantProjections(ctx context.Context, userIDs []int) (map[int]*UserTenantProjection, error)
	// ResolveUserTenantAssignment validates requested memberships and returns a host write plan.
	//
	// ResolveUserTenantAssignment 校验请求的用户租户归属并生成宿主写入计划，适用于用户创建和更新前统一计算主租户与成员关系。
	ResolveUserTenantAssignment(ctx context.Context, requested []TenantID, mode UserTenantAssignmentMode) (*UserTenantAssignmentPlan, error)
	// ReplaceUserTenantAssignments rewrites one user's active tenant ownership rows.
	//
	// ReplaceUserTenantAssignments 重写指定用户的有效租户归属记录，适用于用户保存流程按已验证计划同步租户成员关系。
	ReplaceUserTenantAssignments(ctx context.Context, userID int, plan *UserTenantAssignmentPlan) error
	// EnsureUsersInTenant verifies every user has active membership in the tenant.
	//
	// EnsureUsersInTenant 校验一组用户是否都属于指定租户，适用于批量授权、批量操作和跨租户写入前的边界检查。
	EnsureUsersInTenant(ctx context.Context, userIDs []int, tenantID TenantID) error
}

// PluginProvisioningService defines host-internal tenant plugin provisioning.
// It is startup/lifecycle governance and is not part of ordinary plugin
// consumption.
//
// PluginProvisioningService 定义宿主内部租户插件自动供给能力，适用于启动或生命周期治理阶段为租户启用默认插件，不属于普通插件消费面。
type PluginProvisioningService interface {
	// ProvisionAutoEnabledTenantPlugins provisions default tenant plugins when the provider supports it.
	//
	// ProvisionAutoEnabledTenantPlugins 在 provider 支持时供给默认自动启用的租户插件，适用于 HTTP 启动后源码插件提供方已注册的治理流程。
	ProvisionAutoEnabledTenantPlugins(ctx context.Context) error
}

// StartupConsistencyService defines host-internal tenant startup validation.
// It is only used by the plugin service before HTTP startup completes.
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
	ListUserTenants(ctx context.Context, userID int) ([]TenantInfo, error)
	// ApplyUserTenantScope constrains user-owned rows to the current request tenant.
	//
	// ApplyUserTenantScope 将用户相关查询限制到当前请求租户成员范围，适用于用户列表和用户关联资源查询。
	ApplyUserTenantScope(ctx context.Context, model *gdb.Model, userIDColumn string) (*gdb.Model, bool, error)
	// ApplyUserTenantFilter constrains platform user-list rows to a requested tenant.
	//
	// ApplyUserTenantFilter 将平台用户列表限制到指定租户，适用于平台管理员按租户查看或维护用户。
	ApplyUserTenantFilter(ctx context.Context, model *gdb.Model, userIDColumn string, tenantID TenantID) (*gdb.Model, bool, error)
	// ListUserTenantProjections returns tenant ownership labels for visible users.
	//
	// ListUserTenantProjections 批量返回用户租户归属标签，适用于用户列表、详情批量和导出投影装配。
	ListUserTenantProjections(ctx context.Context, userIDs []int) (map[int]*UserTenantProjection, error)
	// ResolveUserTenantAssignment validates requested memberships and returns a host write plan.
	//
	// ResolveUserTenantAssignment 校验请求成员关系并返回宿主写入计划，适用于用户创建和更新流程。
	ResolveUserTenantAssignment(ctx context.Context, requested []TenantID, mode UserTenantAssignmentMode) (*UserTenantAssignmentPlan, error)
	// ReplaceUserTenantAssignments rewrites one user's active tenant ownership rows.
	//
	// ReplaceUserTenantAssignments 重写指定用户的有效租户归属，适用于宿主用户保存后的成员关系同步。
	ReplaceUserTenantAssignments(ctx context.Context, userID int, plan *UserTenantAssignmentPlan) error
	// EnsureUsersInTenant verifies every user has active membership in the tenant.
	//
	// EnsureUsersInTenant 校验多个用户均属于指定租户，适用于批量操作和跨租户边界校验。
	EnsureUsersInTenant(ctx context.Context, userIDs []int, tenantID TenantID) error
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
	Service
	RequestResolver
	ScopeService
	UserMembershipService
	PluginProvisioningService
	StartupConsistencyService
}

// ProviderFactory creates one tenant provider from an explicit, typed
// construction environment during lazy capability use.
type ProviderFactory func(ctx context.Context, env ProviderEnv) (Provider, error)

// serviceImpl delegates tenant calls to the active provider and returns
// platform-safe fallback values when no provider is usable.
type serviceImpl struct {
	runtime   ProviderRuntime
	bizCtxSvc bizctxcap.Service
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)
var _ RequestResolver = (*serviceImpl)(nil)
var _ ScopeService = (*serviceImpl)(nil)
var _ UserMembershipService = (*serviceImpl)(nil)
var _ PluginProvisioningService = (*serviceImpl)(nil)
var _ StartupConsistencyService = (*serviceImpl)(nil)
var _ RuntimeService = (*serviceImpl)(nil)

// New creates an optional tenant capability service from explicit runtime-owned dependencies.
func New(runtime ProviderRuntime, bizCtxSvc bizctxcap.Service) RuntimeService {
	if runtime == nil {
		runtime = noopProviderRuntime{}
	}
	return &serviceImpl{
		runtime:   runtime,
		bizCtxSvc: bizCtxSvc,
	}
}

// Provide declares one plugin-provided tenant capability factory.
func Provide(pluginID string, factory ProviderFactory) error {
	return registerFactory(pluginID, factory)
}

// CacheKey builds the canonical tenant-aware cache key for runtime caches.
func CacheKey(tenant TenantID, scope string, key string) string {
	return "tenant=" + strconv.Itoa(int(tenant)) +
		":scope=" + strings.TrimSpace(scope) +
		":key=" + strings.TrimSpace(key)
}

// noopProviderRuntime reports all plugins as disabled when tenantcap is
// constructed without an explicit provider runtime.
type noopProviderRuntime struct{}

var defaultManager = internalregistry.NewManager[ProviderEnv]()
