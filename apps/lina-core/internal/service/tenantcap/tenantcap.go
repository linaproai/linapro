// Package tenantcap implements the host-side multi-tenancy capability seam.
// The host owns only no-op defaults and delegates tenant policy to the
// registered multi-tenant plugin provider when it is enabled.
package tenantcap

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/service/bizctx"
	pkgtenantcap "lina-core/pkg/tenantcap"
)

// TenantID is the host-side tenant identifier type.
type TenantID = pkgtenantcap.TenantID

// TenantInfo is the host-side tenant projection type.
type TenantInfo = pkgtenantcap.TenantInfo

// PluginEnablementReader defines the narrow plugin state capability required by tenantcap.
type PluginEnablementReader interface {
	// IsEnabled returns whether the given plugin ID is currently enabled.
	IsEnabled(ctx context.Context, pluginID string) bool
}

// Service defines the optional tenant capability consumed by host core services
// without hard-linking them to a concrete multi-tenant plugin implementation.
type Service interface {
	// Enabled reports whether multi-tenancy is installed, enabled, and backed by
	// a registered provider. Disabled mode must behave as platform-only.
	Enabled(ctx context.Context) bool
	// Current returns the current request tenant from bizctx, defaulting to the
	// platform tenant when request context is unavailable.
	Current(ctx context.Context) TenantID
	// Apply injects tenant filtering into a model when multi-tenancy is enabled
	// and the current platform context is not allowed to bypass filtering.
	Apply(ctx context.Context, model *gdb.Model, tenantColumn string) (*gdb.Model, error)
	// PlatformBypass reports whether the current request may bypass tenant
	// filtering because it is a non-impersonated platform request with all-data
	// scope.
	PlatformBypass(ctx context.Context) bool
	// EnsureTenantVisible validates that the current user can access tenantID;
	// disabled tenancy treats every tenant check as visible for host fallback.
	EnsureTenantVisible(ctx context.Context, tenantID TenantID) error
	// ResolveTenant delegates HTTP tenant resolution to the provider when
	// enabled; disabled or missing providers resolve to platform.
	ResolveTenant(ctx context.Context, r *ghttp.Request) (*pkgtenantcap.ResolverResult, error)
	// ReadWithPlatformFallback reads tenant overrides with platform fallback
	// semantics through a caller-supplied scanner and does not cache results.
	ReadWithPlatformFallback(ctx context.Context, scanner FallbackScanner[any]) ([]any, error)
	// ApplyUserTenantScope constrains user rows by active current-tenant
	// membership and returns empty when no visible memberships remain.
	ApplyUserTenantScope(ctx context.Context, model *gdb.Model, userIDColumn string) (*gdb.Model, bool, error)
	// ListUserTenants returns active tenant memberships visible to one user.
	ListUserTenants(ctx context.Context, userID int) ([]pkgtenantcap.TenantInfo, error)
	// ApplyUserTenantFilter constrains platform user-list rows to a requested
	// tenant and validates provider-owned membership rules.
	ApplyUserTenantFilter(ctx context.Context, model *gdb.Model, userIDColumn string, tenantID TenantID) (*gdb.Model, bool, error)
	// ListUserTenantProjections returns tenant ownership labels for visible
	// users; missing providers return an empty map.
	ListUserTenantProjections(ctx context.Context, userIDs []int) (map[int]*pkgtenantcap.UserTenantProjection, error)
	// ResolveUserTenantAssignment validates requested memberships and returns a
	// host write plan for create/update operations.
	ResolveUserTenantAssignment(ctx context.Context, requested []TenantID, mode pkgtenantcap.UserTenantAssignmentMode) (*pkgtenantcap.UserTenantAssignmentPlan, error)
	// ReplaceUserTenantAssignments rewrites one user's active tenant ownership
	// rows only when the provider produced a replacement plan.
	ReplaceUserTenantAssignments(ctx context.Context, userID int, plan *pkgtenantcap.UserTenantAssignmentPlan) error
	// EnsureUsersInTenant verifies every user has active membership in the
	// tenant before tenant-local relationship writes.
	EnsureUsersInTenant(ctx context.Context, userIDs []int, tenantID TenantID) error
	// ValidateUserMembershipStartupConsistency returns startup consistency
	// failures detected by the provider without mutating host state.
	ValidateUserMembershipStartupConsistency(ctx context.Context) ([]string, error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	enablementReader PluginEnablementReader
	bizCtxSvc        bizctx.Service
}

// New creates and returns a new optional tenant capability service from explicit runtime-owned dependencies.
func New(enablementReader PluginEnablementReader, bizCtxSvc bizctx.Service) Service {
	if enablementReader == nil {
		enablementReader = noopPluginEnablementReader{}
	}
	return &serviceImpl{
		enablementReader: enablementReader,
		bizCtxSvc:        bizCtxSvc,
	}
}

// noopPluginEnablementReader reports all plugins as disabled when tenantcap is
// constructed without an explicit enablement reader.
type noopPluginEnablementReader struct{}
