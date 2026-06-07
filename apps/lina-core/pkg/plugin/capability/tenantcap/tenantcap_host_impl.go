// tenantcap_impl.go implements optional tenant-capability delegation and
// fallback helpers. It checks source-plugin enablement before forwarding tenant
// isolation, membership, and query-scope operations, returning platform-safe
// defaults when multi-tenancy is not installed or not enabled.

package tenantcap

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
)

// Available reports whether an active tenant provider is available.
func (s *serviceImpl) Available(ctx context.Context) bool {
	if s == nil {
		return false
	}
	return defaultManager.StatusWithProvider(ctx, CapabilityTenantV1, s.runtime, s.providerEnv).Available
}

// Status returns the current tenant capability activation state.
func (s *serviceImpl) Status(ctx context.Context) capmodel.CapabilityStatus {
	if s == nil {
		return convertCapabilityStatus(defaultManager.Status(ctx, CapabilityTenantV1, nil))
	}
	return convertCapabilityStatus(defaultManager.StatusWithProvider(ctx, CapabilityTenantV1, s.runtime, s.providerEnv))
}

// currentProvider returns the currently usable tenant-capability provider.
func (s *serviceImpl) currentProvider(ctx context.Context) (Provider, error) {
	if s == nil {
		return nil, nil
	}
	provider, err := defaultManager.ActiveProviderWithError(ctx, CapabilityTenantV1, s.runtime, s.providerEnv)
	if err != nil || provider == nil {
		return nil, err
	}
	typedProvider, ok := provider.(Provider)
	if !ok {
		return nil, nil
	}
	return typedProvider, nil
}

// providerEnv builds lazy construction inputs for one tenant provider.
func (s *serviceImpl) providerEnv(_ context.Context, pluginID string) ProviderEnv {
	env := ProviderEnv{PluginID: pluginID}
	if s != nil && s.runtime != nil {
		env = s.runtime.TenantProviderEnv(pluginID)
	}
	if env.PluginID == "" {
		env.PluginID = pluginID
	}
	return env
}

// Current returns the current request tenant from bizctx, defaulting to platform.
func (s *serviceImpl) Current(ctx context.Context) TenantID {
	if s == nil || s.bizCtxSvc == nil {
		return PLATFORM
	}
	current := s.bizCtxSvc.Current(ctx)
	return TenantID(current.TenantID)
}

// Apply injects tenant filtering into a model when multi-tenancy is enabled.
func (s *serviceImpl) Apply(ctx context.Context, model *gdb.Model, tenantColumn string) (*gdb.Model, error) {
	if model == nil || s.PlatformBypass(ctx) {
		return model, nil
	}
	if _, err := s.currentProvider(ctx); err != nil {
		return nil, err
	}
	return model.Where(tenantColumn, int(s.Current(ctx))), nil
}

// PlatformBypass reports whether the current request may bypass tenant filtering.
func (s *serviceImpl) PlatformBypass(ctx context.Context) bool {
	if s == nil || s.bizCtxSvc == nil {
		return false
	}
	return s.bizCtxSvc.Current(ctx).PlatformBypass
}

// EnsureTenantVisible validates that the current user can access tenantID.
func (s *serviceImpl) EnsureTenantVisible(ctx context.Context, tenantID TenantID) error {
	if s.PlatformBypass(ctx) {
		return nil
	}
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return err
	}
	if provider == nil {
		return nil
	}
	if s.Current(ctx) != tenantID {
		return bizerr.NewCode(CodeTenantForbidden, bizerr.P("tenantId", int(tenantID)))
	}
	if s.bizCtxSvc == nil {
		return bizerr.NewCode(CodeTenantForbidden, bizerr.P("tenantId", int(tenantID)))
	}
	businessCtx := s.bizCtxSvc.Current(ctx)
	if businessCtx.UserID <= 0 {
		return bizerr.NewCode(CodeTenantForbidden, bizerr.P("tenantId", int(tenantID)))
	}
	return provider.ValidateUserInTenant(ctx, businessCtx.UserID, tenantID)
}

// ValidateUserInTenant verifies that a user can access a tenant.
func (s *serviceImpl) ValidateUserInTenant(ctx context.Context, userID int, tenantID TenantID) error {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return err
	}
	if provider == nil {
		return nil
	}
	return provider.ValidateUserInTenant(ctx, userID, tenantID)
}

// SwitchTenant validates a tenant switch before token re-issue.
func (s *serviceImpl) SwitchTenant(ctx context.Context, userID int, target TenantID) error {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return err
	}
	if provider == nil {
		return nil
	}
	return provider.SwitchTenant(ctx, userID, target)
}

// ResolveTenant delegates HTTP tenant resolution to the provider when enabled.
func (s *serviceImpl) ResolveTenant(ctx context.Context, r *ghttp.Request) (*ResolverResult, error) {
	if r == nil {
		return &ResolverResult{TenantID: PLATFORM, Matched: true}, nil
	}
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return &ResolverResult{TenantID: PLATFORM, Matched: true}, nil
	}
	return provider.ResolveTenant(ctx, r)
}

// userMembershipProvider returns the optional user membership capability facet.
func (s *serviceImpl) userMembershipProvider(ctx context.Context) (UserMembershipProvider, error) {
	provider, err := s.currentProvider(ctx)
	if err != nil || provider == nil {
		return nil, err
	}
	membershipProvider, ok := provider.(UserMembershipProvider)
	if !ok {
		return nil, nil
	}
	return membershipProvider, nil
}

// ApplyUserTenantScope constrains user rows by active current-tenant membership.
func (s *serviceImpl) ApplyUserTenantScope(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
) (*gdb.Model, bool, error) {
	provider, err := s.userMembershipProvider(ctx)
	if err != nil {
		return nil, false, err
	}
	if provider == nil {
		return model, false, nil
	}
	return provider.ApplyUserTenantScope(ctx, model, userIDColumn)
}

// ListUserTenants returns the active tenants visible to one user.
func (s *serviceImpl) ListUserTenants(ctx context.Context, userID int) ([]TenantInfo, error) {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil || userID <= 0 {
		return []TenantInfo{}, nil
	}
	return provider.ListUserTenants(ctx, userID)
}

// ApplyUserTenantFilter constrains platform user-list rows to a requested tenant.
func (s *serviceImpl) ApplyUserTenantFilter(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
	tenantID TenantID,
) (*gdb.Model, bool, error) {
	provider, err := s.userMembershipProvider(ctx)
	if err != nil {
		return nil, false, err
	}
	if provider == nil {
		return model, false, nil
	}
	return provider.ApplyUserTenantFilter(ctx, model, userIDColumn, tenantID)
}

// ListUserTenantProjections returns tenant ownership labels for visible users.
func (s *serviceImpl) ListUserTenantProjections(
	ctx context.Context,
	userIDs []int,
) (map[int]*UserTenantProjection, error) {
	result := make(map[int]*UserTenantProjection)
	if len(userIDs) == 0 {
		return result, nil
	}
	provider, err := s.userMembershipProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return result, nil
	}
	return provider.ListUserTenantProjections(ctx, userIDs)
}

// ResolveUserTenantAssignment validates requested memberships and returns a host write plan.
func (s *serviceImpl) ResolveUserTenantAssignment(
	ctx context.Context,
	requested []TenantID,
	mode UserTenantAssignmentMode,
) (*UserTenantAssignmentPlan, error) {
	provider, err := s.userMembershipProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return &UserTenantAssignmentPlan{PrimaryTenant: s.Current(ctx)}, nil
	}
	return provider.ResolveUserTenantAssignment(ctx, requested, mode)
}

// ReplaceUserTenantAssignments rewrites one user's active tenant ownership rows.
func (s *serviceImpl) ReplaceUserTenantAssignments(
	ctx context.Context,
	userID int,
	plan *UserTenantAssignmentPlan,
) error {
	provider, err := s.userMembershipProvider(ctx)
	if err != nil {
		return err
	}
	if provider == nil || plan == nil || !plan.ShouldReplace {
		return nil
	}
	return provider.ReplaceUserTenantAssignments(ctx, userID, plan)
}

// EnsureUsersInTenant verifies every user has active membership in the tenant.
func (s *serviceImpl) EnsureUsersInTenant(ctx context.Context, userIDs []int, tenantID TenantID) error {
	if len(userIDs) == 0 {
		return nil
	}
	provider, err := s.userMembershipProvider(ctx)
	if err != nil {
		return err
	}
	if provider == nil {
		return nil
	}
	return provider.EnsureUsersInTenant(ctx, userIDs, tenantID)
}

// ValidateUserMembershipStartupConsistency returns startup consistency failures.
func (s *serviceImpl) ValidateUserMembershipStartupConsistency(ctx context.Context) ([]string, error) {
	provider, err := s.userMembershipProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, nil
	}
	return provider.ValidateStartupConsistency(ctx)
}

// ProvisionAutoEnabledTenantPlugins provisions default tenant plugins through
// the registered provider when the linapro-tenant-core plugin exposes that optional
// startup governance facet.
func (s *serviceImpl) ProvisionAutoEnabledTenantPlugins(ctx context.Context) error {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return err
	}
	if provider == nil {
		return nil
	}
	provisioningProvider, ok := provider.(PluginProvisioningProvider)
	if !ok {
		return nil
	}
	return provisioningProvider.ProvisionAutoEnabledTenantPlugins(ctx)
}

// IsProviderEnabled always returns false.
func (noopProviderRuntime) IsProviderEnabled(_ context.Context, _ string) bool {
	return false
}

// TenantProviderEnv returns an empty typed provider environment.
func (noopProviderRuntime) TenantProviderEnv(pluginID string) ProviderEnv {
	return ProviderEnv{PluginID: pluginID}
}
