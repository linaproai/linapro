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
	pkgtenantcap "lina-core/pkg/tenantcap"
)

// Enabled reports whether multi-tenancy is currently installed and enabled.
func (s *serviceImpl) Enabled(ctx context.Context) bool {
	if s == nil || s.enablementReader == nil {
		return false
	}
	if !s.enablementReader.IsEnabled(ctx, pkgtenantcap.ProviderPluginID) {
		return false
	}
	return pkgtenantcap.HasProvider()
}

// Current returns the current request tenant from bizctx, defaulting to platform.
func (s *serviceImpl) Current(ctx context.Context) TenantID {
	if s == nil || s.bizCtxSvc == nil {
		return pkgtenantcap.PLATFORM
	}
	businessCtx := s.bizCtxSvc.Get(ctx)
	if businessCtx == nil {
		return pkgtenantcap.PLATFORM
	}
	return TenantID(businessCtx.TenantId)
}

// Apply injects tenant filtering into a model when multi-tenancy is enabled.
func (s *serviceImpl) Apply(ctx context.Context, model *gdb.Model, tenantColumn string) (*gdb.Model, error) {
	if model == nil || !s.Enabled(ctx) || s.PlatformBypass(ctx) {
		return model, nil
	}
	return model.Where(tenantColumn, int(s.Current(ctx))), nil
}

// PlatformBypass reports whether the current request may bypass tenant filtering.
func (s *serviceImpl) PlatformBypass(ctx context.Context) bool {
	if s == nil || s.bizCtxSvc == nil {
		return false
	}
	businessCtx := s.bizCtxSvc.Get(ctx)
	if businessCtx == nil {
		return false
	}
	return businessCtx.TenantId == int(pkgtenantcap.PLATFORM) &&
		businessCtx.DataScope == 1 &&
		!businessCtx.DataScopeUnsupported &&
		!businessCtx.ActingAsTenant &&
		!businessCtx.IsImpersonation
}

// EnsureTenantVisible validates that the current user can access tenantID.
func (s *serviceImpl) EnsureTenantVisible(ctx context.Context, tenantID TenantID) error {
	if !s.Enabled(ctx) {
		return nil
	}
	if s.PlatformBypass(ctx) {
		return nil
	}
	if s.Current(ctx) != tenantID {
		return bizerr.NewCode(pkgtenantcap.CodeTenantForbidden, bizerr.P("tenantId", int(tenantID)))
	}
	provider := pkgtenantcap.CurrentProvider()
	if provider == nil {
		return nil
	}
	businessCtx := s.bizCtxSvc.Get(ctx)
	if businessCtx == nil || businessCtx.UserId <= 0 {
		return bizerr.NewCode(pkgtenantcap.CodeTenantForbidden, bizerr.P("tenantId", int(tenantID)))
	}
	return provider.ValidateUserInTenant(ctx, businessCtx.UserId, tenantID)
}

// ResolveTenant delegates HTTP tenant resolution to the provider when enabled.
func (s *serviceImpl) ResolveTenant(ctx context.Context, r *ghttp.Request) (*pkgtenantcap.ResolverResult, error) {
	if !s.Enabled(ctx) {
		return &pkgtenantcap.ResolverResult{TenantID: pkgtenantcap.PLATFORM, Matched: true}, nil
	}
	if r == nil {
		return &pkgtenantcap.ResolverResult{TenantID: pkgtenantcap.PLATFORM, Matched: true}, nil
	}
	provider := pkgtenantcap.CurrentProvider()
	if provider == nil {
		return &pkgtenantcap.ResolverResult{TenantID: pkgtenantcap.PLATFORM, Matched: true}, nil
	}
	return provider.ResolveTenant(ctx, r)
}

// userMembershipProvider returns the optional user membership capability facet.
func (s *serviceImpl) userMembershipProvider(ctx context.Context) pkgtenantcap.UserMembershipProvider {
	if !s.Enabled(ctx) {
		return nil
	}
	provider := pkgtenantcap.CurrentProvider()
	if provider == nil {
		return nil
	}
	membershipProvider, ok := provider.(pkgtenantcap.UserMembershipProvider)
	if !ok {
		return nil
	}
	return membershipProvider
}

// ApplyUserTenantScope constrains user rows by active current-tenant membership.
func (s *serviceImpl) ApplyUserTenantScope(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
) (*gdb.Model, bool, error) {
	provider := s.userMembershipProvider(ctx)
	if provider == nil {
		return model, false, nil
	}
	return provider.ApplyUserTenantScope(ctx, model, userIDColumn)
}

// ListUserTenants returns the active tenants visible to one user.
func (s *serviceImpl) ListUserTenants(ctx context.Context, userID int) ([]pkgtenantcap.TenantInfo, error) {
	provider := s.userMembershipProvider(ctx)
	if provider == nil || userID <= 0 {
		return []pkgtenantcap.TenantInfo{}, nil
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
	provider := s.userMembershipProvider(ctx)
	if provider == nil {
		return model, false, nil
	}
	return provider.ApplyUserTenantFilter(ctx, model, userIDColumn, tenantID)
}

// ListUserTenantProjections returns tenant ownership labels for visible users.
func (s *serviceImpl) ListUserTenantProjections(
	ctx context.Context,
	userIDs []int,
) (map[int]*pkgtenantcap.UserTenantProjection, error) {
	result := make(map[int]*pkgtenantcap.UserTenantProjection)
	if len(userIDs) == 0 {
		return result, nil
	}
	provider := s.userMembershipProvider(ctx)
	if provider == nil {
		return result, nil
	}
	return provider.ListUserTenantProjections(ctx, userIDs)
}

// ResolveUserTenantAssignment validates requested memberships and returns a host write plan.
func (s *serviceImpl) ResolveUserTenantAssignment(
	ctx context.Context,
	requested []TenantID,
	mode pkgtenantcap.UserTenantAssignmentMode,
) (*pkgtenantcap.UserTenantAssignmentPlan, error) {
	provider := s.userMembershipProvider(ctx)
	if provider == nil {
		return &pkgtenantcap.UserTenantAssignmentPlan{PrimaryTenant: s.Current(ctx)}, nil
	}
	return provider.ResolveUserTenantAssignment(ctx, requested, mode)
}

// ReplaceUserTenantAssignments rewrites one user's active tenant ownership rows.
func (s *serviceImpl) ReplaceUserTenantAssignments(
	ctx context.Context,
	userID int,
	plan *pkgtenantcap.UserTenantAssignmentPlan,
) error {
	provider := s.userMembershipProvider(ctx)
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
	provider := s.userMembershipProvider(ctx)
	if provider == nil {
		return nil
	}
	return provider.EnsureUsersInTenant(ctx, userIDs, tenantID)
}

// ValidateUserMembershipStartupConsistency returns startup consistency failures.
func (s *serviceImpl) ValidateUserMembershipStartupConsistency(ctx context.Context) ([]string, error) {
	provider := s.userMembershipProvider(ctx)
	if provider == nil {
		return nil, nil
	}
	return provider.ValidateStartupConsistency(ctx)
}

// IsEnabled always returns false.
func (noopPluginEnablementReader) IsEnabled(_ context.Context, _ string) bool {
	return false
}
