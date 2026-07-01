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

// Provider defines the multi-tenancy capability implemented by plugins.
type Provider interface {
	// ResolveTenant resolves tenant identity for one HTTP request.
	ResolveTenant(ctx context.Context, r *ghttp.Request) (*tenantcap.ResolverResult, error)
	// ValidateUserInTenant verifies that a user can access a tenant.
	ValidateUserInTenant(ctx context.Context, userID int, tenantID tenantcap.TenantID) error
	// ListUserTenants returns the active tenants visible to one user.
	ListUserTenants(ctx context.Context, userID int) ([]tenantcap.TenantInfo, error)
	// SwitchTenant validates a tenant switch before token re-issue.
	SwitchTenant(ctx context.Context, userID int, target tenantcap.TenantID) error
}

// DirectoryProvider optionally exposes read-only tenant projections.
type DirectoryProvider interface {
	// Info returns one tenant projection visible in the current context.
	Info(ctx context.Context, tenantID tenantcap.TenantID) (*tenantcap.TenantInfo, error)
	// BatchGet returns visible tenant projections and opaque missing IDs.
	BatchGet(ctx context.Context, tenantIDs []tenantcap.TenantID) (*capmodel.BatchResult[*tenantcap.TenantInfo, tenantcap.TenantID], error)
	// List returns bounded tenant candidates visible to the caller.
	List(ctx context.Context, input tenantcap.ListInput) (*capmodel.PageResult[*tenantcap.TenantInfo], error)
	// EnsureVisible validates that every tenant identifier is visible.
	EnsureVisible(ctx context.Context, tenantIDs []tenantcap.TenantID) error
}

// RequestResolver defines host-internal HTTP tenant resolution. It is kept out
// of tenantcap.Service because ordinary plugin consumers must not receive
// request-object based resolver hooks through capability services.
type RequestResolver interface {
	// Available reports whether tenant resolution should use provider-backed logic.
	Available(ctx context.Context) bool
	// ResolveTenant delegates HTTP tenant resolution to the provider when enabled.
	ResolveTenant(ctx context.Context, r *ghttp.Request) (*tenantcap.ResolverResult, error)
}

// ScopeService defines host-internal tenant query-scope operations.
type ScopeService interface {
	// Available reports whether tenant-scoped database filtering can run.
	Available(ctx context.Context) bool
	// Apply injects tenant filtering into a model when multi-tenancy is enabled.
	Apply(ctx context.Context, model *gdb.Model, tenantColumn string) (*gdb.Model, error)
	// ApplyUserTenantScope constrains user rows by active current-tenant membership.
	ApplyUserTenantScope(ctx context.Context, model *gdb.Model, userIDColumn string) (*gdb.Model, bool, error)
	// ApplyUserTenantFilter constrains platform user-list rows to a requested tenant.
	ApplyUserTenantFilter(ctx context.Context, model *gdb.Model, userIDColumn string, tenantID tenantcap.TenantID) (*gdb.Model, bool, error)
}

// UserMembershipService defines host-internal user-to-tenant membership projections and mutations.
type UserMembershipService interface {
	// ListUserTenants returns the active tenants visible to one user.
	ListUserTenants(ctx context.Context, userID int) ([]tenantcap.TenantInfo, error)
	// ValidateUserInTenant verifies that userID may operate inside tenantID.
	ValidateUserInTenant(ctx context.Context, userID int, tenantID tenantcap.TenantID) error
	// ListUserTenantMemberships returns tenant ownership labels for visible users.
	ListUserTenantMemberships(ctx context.Context, userIDs []int) (map[int]*tenantcap.TenantMembershipInfo, error)
	// ResolveUserTenantAssignment validates requested memberships and returns a host write plan.
	ResolveUserTenantAssignment(ctx context.Context, requested []tenantcap.TenantID, mode tenantcap.UserTenantAssignmentMode) (*tenantcap.UserTenantAssignmentPlan, error)
	// ReplaceUserTenantAssignments rewrites one user's active tenant ownership rows.
	ReplaceUserTenantAssignments(ctx context.Context, userID int, plan *tenantcap.UserTenantAssignmentPlan) error
	// EnsureUsersInTenant verifies every user has active membership in the tenant.
	EnsureUsersInTenant(ctx context.Context, userIDs []int, tenantID tenantcap.TenantID) error
	// SwitchTenant validates a tenant switch before token re-issue.
	SwitchTenant(ctx context.Context, userID int, target tenantcap.TenantID) error
}

// HostGovernanceService groups host-internal tenant governance operations that
// do not belong to ordinary plugin-visible tenant capabilities.
type HostGovernanceService interface {
	// PlatformBypass reports whether the current context may operate at platform scope.
	PlatformBypass(ctx context.Context) bool
	// EnsureTenantVisible validates that the current user can access tenantID.
	EnsureTenantVisible(ctx context.Context, tenantID tenantcap.TenantID) error
	// ProvisionAutoEnabledTenantPlugins provisions default tenant plugins when the provider supports it.
	ProvisionAutoEnabledTenantPlugins(ctx context.Context) error
	// ValidateUserMembershipStartupConsistency returns startup consistency failures detected by the provider.
	ValidateUserMembershipStartupConsistency(ctx context.Context) ([]string, error)
}

// UserMembershipProvider optionally exposes user-to-tenant ownership behavior.
type UserMembershipProvider interface {
	// ListUserTenants returns the active tenants visible to one user.
	ListUserTenants(ctx context.Context, userID int) ([]tenantcap.TenantInfo, error)
	// ApplyUserTenantScope constrains user-owned rows to the current request tenant.
	ApplyUserTenantScope(ctx context.Context, model *gdb.Model, userIDColumn string) (*gdb.Model, bool, error)
	// ApplyUserTenantFilter constrains platform user-list rows to a requested tenant.
	ApplyUserTenantFilter(ctx context.Context, model *gdb.Model, userIDColumn string, tenantID tenantcap.TenantID) (*gdb.Model, bool, error)
	// ListUserTenantMemberships returns tenant ownership labels for visible users.
	ListUserTenantMemberships(ctx context.Context, userIDs []int) (map[int]*tenantcap.TenantMembershipInfo, error)
	// ResolveUserTenantAssignment validates requested memberships and returns a host write plan.
	ResolveUserTenantAssignment(ctx context.Context, requested []tenantcap.TenantID, mode tenantcap.UserTenantAssignmentMode) (*tenantcap.UserTenantAssignmentPlan, error)
	// ReplaceUserTenantAssignments rewrites one user's active tenant ownership rows.
	ReplaceUserTenantAssignments(ctx context.Context, userID int, plan *tenantcap.UserTenantAssignmentPlan) error
	// EnsureUsersInTenant verifies every user has active membership in the tenant.
	EnsureUsersInTenant(ctx context.Context, userIDs []int, tenantID tenantcap.TenantID) error
	// ValidateStartupConsistency returns user-membership startup consistency failures.
	ValidateStartupConsistency(ctx context.Context) ([]string, error)
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
	// Tenant exposes the tenant domain capability, including tenant plugin governance.
	Tenant tenantcap.Service
	// Users resolves host-owned user projections without exposing sys_user storage.
	Users usercap.Service
	// Plugins resolves host-owned plugin governance projections.
	Plugins plugincap.Service
}

// ProviderFactory creates one tenant provider from an explicit, typed
// construction environment during lazy capability use.
type ProviderFactory func(ctx context.Context, env ProviderEnv) (Provider, error)

// Service defines the public host tenant runtime contract returned by New.
// It names the tenantspi-owned runtime slices without exposing serviceImpl or
// reintroducing plugin table filtering into the ordinary tenant capability.
type Service interface {
	tenantcap.Service
	RequestResolver
	ScopeService
	UserMembershipService
	HostGovernanceService
}

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
	manager    *Manager
	enablement internalregistry.EnablementReader
	envFactory internalregistry.ProviderEnvFactory[ProviderEnv]
	bizCtxSvc  bizctxcap.Service
}

// Ensure serviceImpl implements the published tenant capability and retained host SPI slices.
var (
	_ Service                     = (*serviceImpl)(nil)
	_ tenantcap.Service           = (*serviceImpl)(nil)
	_ tenantcap.ContextService    = (*tenantContextService)(nil)
	_ tenantcap.DirectoryService  = (*tenantDirectoryService)(nil)
	_ tenantcap.MembershipService = (*tenantMembershipService)(nil)
	_ RequestResolver             = (*serviceImpl)(nil)
	_ ScopeService                = (*serviceImpl)(nil)
	_ UserMembershipService       = (*serviceImpl)(nil)
	_ HostGovernanceService       = (*serviceImpl)(nil)
)

// New creates an optional tenant capability service from explicit runtime-owned dependencies.
func New(
	manager *Manager,
	enablement internalregistry.EnablementReader,
	envFactory internalregistry.ProviderEnvFactory[ProviderEnv],
	bizCtxSvc bizctxcap.Service,
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
	return &serviceImpl{
		manager:    manager,
		enablement: enablement,
		envFactory: envFactory,
		bizCtxSvc:  bizCtxSvc,
	}
}

// noopEnablementReader reports all provider plugins as disabled.
type noopEnablementReader struct{}

// ProviderStatuses returns all tenant provider states.
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
