// Package tenantcap publishes the stable multi-tenancy capability contract
// shared between the host and source plugins so tenant ownership and
// membership policy can remain plugin-owned.
package tenantcap

import (
	"context"
	"strconv"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ProviderPluginID is the official source-plugin identifier that provides the
// multi-tenancy capability.
const ProviderPluginID = "multi-tenant"

// TenantID identifies one tenant in the pooled tenancy model.
type TenantID int

// Tenant identity constants define platform and all-tenant sentinel values.
const (
	// PlatformTenantID is the platform tenant used by single-tenant mode and platform defaults.
	PlatformTenantID TenantID = 0
	// AllTenantsID is the explicit all-tenant cache invalidation sentinel.
	AllTenantsID TenantID = -1
)

// Backward-compatible exported aliases keep call sites concise.
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

// UserTenantProjection describes the host-facing tenant ownership projection
// for one user row.
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

// UserTenantAssignmentMode identifies the host operation that requested
// membership assignment planning.
type UserTenantAssignmentMode string

// User tenant assignment modes.
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
type Resolver interface {
	// Name returns the stable resolver name used by configuration.
	Name() ResolverName
	// Resolve attempts to resolve tenant identity for the current request.
	Resolve(ctx context.Context, r *ghttp.Request) (*ResolverResult, error)
}

// Provider defines the stable multi-tenancy capability implemented by plugins.
type Provider interface {
	// ResolveTenant resolves tenant identity for one HTTP request.
	ResolveTenant(ctx context.Context, r *ghttp.Request) (*ResolverResult, error)
	// ValidateUserInTenant verifies that a user can access a tenant.
	ValidateUserInTenant(ctx context.Context, userID int, tenantID TenantID) error
	// ListUserTenants returns the active tenants visible to one user.
	ListUserTenants(ctx context.Context, userID int) ([]TenantInfo, error)
	// SwitchTenant validates a tenant switch before token re-issue.
	SwitchTenant(ctx context.Context, userID int, target TenantID) error
}

// UserMembershipProvider optionally exposes user-to-tenant ownership behavior.
// Providers implement this facet when plugin-owned membership data must
// participate in host user, role, notification, and startup-governance flows.
type UserMembershipProvider interface {
	// ListUserTenants returns the active tenants visible to one user.
	ListUserTenants(ctx context.Context, userID int) ([]TenantInfo, error)
	// ApplyUserTenantScope constrains user-owned rows to the current request tenant.
	ApplyUserTenantScope(ctx context.Context, model *gdb.Model, userIDColumn string) (*gdb.Model, bool, error)
	// ApplyUserTenantFilter constrains platform user-list rows to a requested tenant.
	ApplyUserTenantFilter(ctx context.Context, model *gdb.Model, userIDColumn string, tenantID TenantID) (*gdb.Model, bool, error)
	// ListUserTenantProjections returns tenant ownership labels for visible users.
	ListUserTenantProjections(ctx context.Context, userIDs []int) (map[int]*UserTenantProjection, error)
	// ResolveUserTenantAssignment validates requested memberships and returns a host write plan.
	ResolveUserTenantAssignment(ctx context.Context, requested []TenantID, mode UserTenantAssignmentMode) (*UserTenantAssignmentPlan, error)
	// ReplaceUserTenantAssignments rewrites one user's active tenant ownership rows.
	ReplaceUserTenantAssignments(ctx context.Context, userID int, plan *UserTenantAssignmentPlan) error
	// EnsureUsersInTenant verifies every user has active membership in the tenant.
	EnsureUsersInTenant(ctx context.Context, userIDs []int, tenantID TenantID) error
	// ValidateStartupConsistency returns user-membership startup consistency failures.
	ValidateStartupConsistency(ctx context.Context) ([]string, error)
}

var (
	providerMu sync.RWMutex
	provider   Provider
)

// RegisterProvider publishes the current tenant-capability provider.
func RegisterProvider(p Provider) {
	providerMu.Lock()
	defer providerMu.Unlock()

	provider = p
}

// CurrentProvider returns the registered tenant-capability provider.
func CurrentProvider() Provider {
	providerMu.RLock()
	defer providerMu.RUnlock()

	return provider
}

// HasProvider reports whether one tenant-capability provider is registered.
func HasProvider() bool {
	return CurrentProvider() != nil
}

// CacheKey builds the canonical tenant-aware cache key for runtime caches.
func CacheKey(tenant TenantID, scope string, key string) string {
	return "tenant=" + strconv.Itoa(int(tenant)) +
		":scope=" + strings.TrimSpace(scope) +
		":key=" + strings.TrimSpace(key)
}
