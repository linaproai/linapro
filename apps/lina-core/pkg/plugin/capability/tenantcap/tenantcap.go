// Package tenantcap owns the stable tenant capability contract exposed through
// capability. Provider SPI, HTTP request resolution, and database query-scope
// seams live in tenantspi so ordinary consumers do not see host-only types.
package tenantcap

import (
	"context"
	"strconv"
	"strings"

	"lina-core/pkg/plugin/capability/capmodel"
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

const (
	// MaxTenantBatchSize is the maximum tenant identifiers accepted by batch tenant reads.
	MaxTenantBatchSize = 200
	// MaxTenantSearchPageSize is the maximum tenant candidate page size.
	MaxTenantSearchPageSize = 200
	// MaxUserTenantBatchSize is the maximum user count accepted by batch membership reads.
	MaxUserTenantBatchSize = 200
)

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

// SearchInput describes bounded tenant candidate search.
type SearchInput struct {
	Keyword string               // Keyword matches stable tenant code or name.
	Code    string               // Code filters by tenant code fragment.
	Name    string               // Name filters by tenant name fragment.
	Status  string               // Status optionally filters by tenant lifecycle status.
	Page    capmodel.PageRequest // Page constrains page number, page size and optional limit.
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
	// CurrentTenantInfo returns the current request tenant projection.
	//
	// CurrentTenantInfo 返回当前请求租户投影，适用于插件上下文展示和租户感知业务分支；租户能力不可用时返回平台租户中性投影。
	CurrentTenantInfo(ctx context.Context) (*TenantInfo, error)
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
	// BatchGetTenants returns visible tenant projections and opaque missing IDs.
	//
	// BatchGetTenants 批量返回当前 actor 可见租户投影，不存在、不可见和租户外目标统一进入缺失集合。
	BatchGetTenants(ctx context.Context, tenantIDs []TenantID) (*capmodel.BatchResult[*TenantInfo, TenantID], error)
	// SearchTenants returns bounded tenant candidates visible to the caller.
	//
	// SearchTenants 返回分页租户候选投影，适用于插件筛选和关系选择；provider 缺失时返回空页。
	SearchTenants(ctx context.Context, input SearchInput) (*capmodel.PageResult[*TenantInfo], error)
	// BatchListUserTenants returns active tenant memberships for visible users.
	//
	// BatchListUserTenants 批量返回用户可访问租户列表，适用于列表和批量详情装配，避免逐用户查询 provider。
	BatchListUserTenants(ctx context.Context, userIDs []int) (map[int][]TenantInfo, error)
	// EnsureTenantsVisible validates that the current user can access every tenant.
	//
	// EnsureTenantsVisible 批量校验当前用户可访问指定租户，任一目标不可见时整体拒绝。
	EnsureTenantsVisible(ctx context.Context, tenantIDs []TenantID) error
	// SwitchTenant validates a tenant switch before token re-issue.
	//
	// SwitchTenant 校验指定用户切换到目标租户是否合法，适用于租户切换接口在重新签发令牌前执行准入检查。
	SwitchTenant(ctx context.Context, userID int, target TenantID) error
}

// CacheKey builds the canonical tenant-aware cache key for runtime caches.
func CacheKey(tenant TenantID, scope string, key string) string {
	return "tenant=" + strconv.Itoa(int(tenant)) +
		":scope=" + strings.TrimSpace(scope) +
		":key=" + strings.TrimSpace(key)
}
