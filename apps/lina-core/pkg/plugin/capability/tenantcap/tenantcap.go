// Package tenantcap owns the stable tenant capability contract exposed through
// capability. Provider SPI, HTTP request resolution, and database query-scope
// seams live in tenantspi so ordinary consumers do not see host-only types.
package tenantcap

import (
	"context"
	"strconv"
	"strings"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/plugincap"
)

// Service defines the optional tenant capability consumed by host core services
// and plugins without hard-linking them to a concrete provider implementation.
type Service interface {
	// Available reports whether an active tenant provider is available.
	Available(ctx context.Context) bool
	// Status returns the current tenant capability activation state.
	Status(ctx context.Context) capmodel.CapabilityStatus
	// Context returns current-tenant context operations.
	Context() ContextService
	// Directory returns tenant directory operations.
	Directory() DirectoryService
	// Membership returns user-to-tenant membership operations.
	Membership() MembershipService
	// Plugins returns tenant-plugin governance operations.
	Plugins() PluginService
	// Filter returns plugin-visible tenant filter context operations.
	Filter() FilterService
}

// ContextService defines current-tenant context reads.
type ContextService interface {
	// Current returns the current request tenant, defaulting to platform when context is unavailable.
	Current(ctx context.Context) TenantID
	// Info returns the current request tenant information.
	Info(ctx context.Context) (*TenantInfo, error)
	// PlatformBypass reports whether the current request may bypass tenant filtering.
	PlatformBypass(ctx context.Context) bool
}

// DirectoryService defines plugin-visible tenant directory operations.
type DirectoryService interface {
	// Get returns one visible tenant info record.
	Get(ctx context.Context, tenantID TenantID) (*TenantInfo, error)
	// BatchGet returns visible tenant info records and opaque missing IDs.
	BatchGet(ctx context.Context, tenantIDs []TenantID) (*capmodel.BatchResult[*TenantInfo, TenantID], error)
	// List returns bounded tenant candidates visible to the caller.
	List(ctx context.Context, input ListInput) (*capmodel.PageResult[*TenantInfo], error)
	// EnsureVisible validates that the current user can access tenant identifiers.
	EnsureVisible(ctx context.Context, tenantIDs []TenantID) error
}

// MembershipService defines plugin-visible user-to-tenant membership operations.
type MembershipService interface {
	// ListByUser returns active tenant memberships visible to one user.
	ListByUser(ctx context.Context, userID int) ([]TenantInfo, error)
	// Validate verifies that a user can access a tenant.
	Validate(ctx context.Context, userID int, tenantID TenantID) error
}

// PluginService defines tenant-plugin governance operations under the tenant domain.
type PluginService interface {
	// SetTenantPluginEnabled updates one tenant plugin enablement row after caller and tenant policy checks.
	SetTenantPluginEnabled(ctx context.Context, pluginID plugincap.PluginID, enabled bool) error
	// ProvisionTenantPluginDefaults creates missing default plugin rows for one tenant.
	ProvisionTenantPluginDefaults(ctx context.Context, tenantID capmodel.DomainID) error
}

// FilterService defines plugin-visible tenant filter context reads.
type FilterService interface {
	// Context returns tenant, actor, impersonation, and platform-bypass metadata for the current request.
	Context(ctx context.Context) TenantFilterContext
}

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
)

// TenantInfo describes one host-facing tenant information record.
type TenantInfo struct {
	ID     TenantID // ID is the numeric tenant identifier.
	Code   string   // Code is the stable tenant code.
	Name   string   // Name is the tenant display name.
	Status string   // Status is the provider-owned tenant lifecycle status.
}

// TenantMembershipInfo describes the host-facing tenant membership for one user row.
type TenantMembershipInfo struct {
	TenantIDs   []TenantID // TenantIDs are active tenant identifiers.
	TenantNames []string   // TenantNames are active tenant display names.
}

// TenantFilterContext carries plugin-visible tenant and audit identity metadata.
type TenantFilterContext struct {
	UserID             int  // UserID is the authenticated user bound to the current request.
	TenantID           int  // TenantID is the current request tenant.
	ActingUserID       int  // ActingUserID is the real actor to persist in audit records.
	OnBehalfOfTenantID int  // OnBehalfOfTenantID is set only when the request acts on behalf of a tenant.
	ActingAsTenant     bool // ActingAsTenant reports whether the request acts through a tenant view.
	IsImpersonation    bool // IsImpersonation marks platform impersonation.
	PlatformBypass     bool // PlatformBypass reports whether the request runs in platform scope.
}

// ListInput describes bounded tenant candidate listing.
type ListInput struct {
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

// CacheKey builds the canonical tenant-aware cache key for runtime caches.
func CacheKey(tenant TenantID, scope string, key string) string {
	return "tenant=" + strconv.Itoa(int(tenant)) +
		":scope=" + strings.TrimSpace(scope) +
		":key=" + strings.TrimSpace(key)
}
