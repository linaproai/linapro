// tenantcap_impl.go implements optional tenant-capability delegation and
// fallback helpers. It checks source-plugin enablement before forwarding tenant
// isolation, membership, and query-scope operations, returning platform-safe
// defaults when multi-tenancy is not installed or not enabled.

package tenantspi

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/tenantcap"
)

// Available reports whether an active tenant provider is available.
func (s *serviceImpl) Available(ctx context.Context) bool {
	if s == nil {
		return false
	}
	return s.manager.registry.StatusWithProvider(ctx, tenantcap.CapabilityTenantV1, s.runtime, s.providerEnv).Available
}

// Status returns the current tenant capability activation state.
func (s *serviceImpl) Status(ctx context.Context) capmodel.CapabilityStatus {
	if s == nil {
		return convertCapabilityStatus(NewManager().registry.Status(ctx, tenantcap.CapabilityTenantV1, nil))
	}
	return convertCapabilityStatus(s.manager.registry.StatusWithProvider(ctx, tenantcap.CapabilityTenantV1, s.runtime, s.providerEnv))
}

// currentProvider returns the currently usable tenant-capability provider.
func (s *serviceImpl) currentProvider(ctx context.Context) (Provider, error) {
	if s == nil {
		return nil, nil
	}
	provider, err := s.manager.registry.ActiveProviderWithError(ctx, tenantcap.CapabilityTenantV1, s.runtime, s.providerEnv)
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
func (s *serviceImpl) Current(ctx context.Context) tenantcap.TenantID {
	if s == nil || s.bizCtxSvc == nil {
		return tenantcap.PLATFORM
	}
	current := s.bizCtxSvc.Current(ctx)
	return tenantcap.TenantID(current.TenantID)
}

// CurrentTenantInfo returns the current request tenant projection.
func (s *serviceImpl) CurrentTenantInfo(ctx context.Context) (*tenantcap.TenantInfo, error) {
	current := s.Current(ctx)
	if current <= tenantcap.PLATFORM {
		return platformTenantInfo(), nil
	}
	provider, err := s.tenantProjectionProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return &tenantcap.TenantInfo{ID: current}, nil
	}
	return provider.CurrentTenantInfo(ctx, current)
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
func (s *serviceImpl) EnsureTenantVisible(ctx context.Context, tenantID tenantcap.TenantID) error {
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
		return bizerr.NewCode(tenantcap.CodeTenantForbidden, bizerr.P("tenantId", int(tenantID)))
	}
	if s.bizCtxSvc == nil {
		return bizerr.NewCode(tenantcap.CodeTenantForbidden, bizerr.P("tenantId", int(tenantID)))
	}
	businessCtx := s.bizCtxSvc.Current(ctx)
	if businessCtx.UserID <= 0 {
		return bizerr.NewCode(tenantcap.CodeTenantForbidden, bizerr.P("tenantId", int(tenantID)))
	}
	return provider.ValidateUserInTenant(ctx, businessCtx.UserID, tenantID)
}

// ValidateUserInTenant verifies that a user can access a tenant.
func (s *serviceImpl) ValidateUserInTenant(ctx context.Context, userID int, tenantID tenantcap.TenantID) error {
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
func (s *serviceImpl) SwitchTenant(ctx context.Context, userID int, target tenantcap.TenantID) error {
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
func (s *serviceImpl) ResolveTenant(ctx context.Context, r *ghttp.Request) (*tenantcap.ResolverResult, error) {
	if r == nil {
		return &tenantcap.ResolverResult{TenantID: tenantcap.PLATFORM, Matched: true}, nil
	}
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return &tenantcap.ResolverResult{TenantID: tenantcap.PLATFORM, Matched: true}, nil
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
func (s *serviceImpl) ListUserTenants(ctx context.Context, userID int) ([]tenantcap.TenantInfo, error) {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil || userID <= 0 {
		return []tenantcap.TenantInfo{}, nil
	}
	return provider.ListUserTenants(ctx, userID)
}

// BatchGetTenants returns visible tenant projections and opaque missing IDs.
func (s *serviceImpl) BatchGetTenants(
	ctx context.Context,
	tenantIDs []tenantcap.TenantID,
) (*capmodel.BatchResult[*tenantcap.TenantInfo, tenantcap.TenantID], error) {
	result := &capmodel.BatchResult[*tenantcap.TenantInfo, tenantcap.TenantID]{
		Items:      make(map[tenantcap.TenantID]*tenantcap.TenantInfo),
		MissingIDs: make([]tenantcap.TenantID, 0),
	}
	if len(tenantIDs) == 0 {
		return result, nil
	}
	if len(tenantIDs) > tenantcap.MaxTenantBatchSize {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", tenantcap.MaxTenantBatchSize))
	}
	normalized := normalizeTenantIDs(tenantIDs)
	if len(normalized) == 0 {
		return result, nil
	}
	provider, err := s.tenantProjectionProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		for _, tenantID := range normalized {
			if tenantID == tenantcap.PLATFORM {
				result.Items[tenantID] = platformTenantInfo()
				continue
			}
			result.MissingIDs = append(result.MissingIDs, tenantID)
		}
		return result, nil
	}
	return provider.BatchGetTenants(ctx, normalized)
}

// SearchTenants returns bounded tenant candidates visible to the caller.
func (s *serviceImpl) SearchTenants(
	ctx context.Context,
	input tenantcap.SearchInput,
) (*capmodel.PageResult[*tenantcap.TenantInfo], error) {
	provider, err := s.tenantProjectionProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return &capmodel.PageResult[*tenantcap.TenantInfo]{Items: []*tenantcap.TenantInfo{}}, nil
	}
	input.Page = normalizeTenantPage(input.Page)
	return provider.SearchTenants(ctx, input)
}

// BatchListUserTenants returns active tenant memberships for visible users.
func (s *serviceImpl) BatchListUserTenants(ctx context.Context, userIDs []int) (map[int][]tenantcap.TenantInfo, error) {
	result := make(map[int][]tenantcap.TenantInfo)
	if len(userIDs) == 0 {
		return result, nil
	}
	if len(userIDs) > tenantcap.MaxUserTenantBatchSize {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", tenantcap.MaxUserTenantBatchSize))
	}
	normalized := normalizePositiveUserIDs(userIDs)
	if len(normalized) == 0 {
		return result, nil
	}
	provider, err := s.userMembershipProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return result, nil
	}
	return provider.BatchListUserTenants(ctx, normalized)
}

// EnsureTenantsVisible validates that the current user can access every tenant.
func (s *serviceImpl) EnsureTenantsVisible(ctx context.Context, tenantIDs []tenantcap.TenantID) error {
	if len(tenantIDs) == 0 {
		return nil
	}
	if len(tenantIDs) > tenantcap.MaxTenantBatchSize {
		return bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", tenantcap.MaxTenantBatchSize))
	}
	normalized := normalizeTenantIDs(tenantIDs)
	if len(normalized) == 0 {
		return nil
	}
	provider, err := s.tenantProjectionProvider(ctx)
	if err != nil {
		return err
	}
	if provider == nil {
		for _, tenantID := range normalized {
			if tenantID != tenantcap.PLATFORM {
				return bizerr.NewCode(tenantcap.CodeTenantForbidden, bizerr.P("tenantId", int(tenantID)))
			}
		}
		return nil
	}
	return provider.EnsureTenantsVisible(ctx, normalized)
}

// ApplyUserTenantFilter constrains platform user-list rows to a requested tenant.
func (s *serviceImpl) ApplyUserTenantFilter(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
	tenantID tenantcap.TenantID,
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

// tenantProjectionProvider returns the optional tenant projection capability facet.
func (s *serviceImpl) tenantProjectionProvider(ctx context.Context) (TenantProjectionProvider, error) {
	provider, err := s.currentProvider(ctx)
	if err != nil || provider == nil {
		return nil, err
	}
	projectionProvider, ok := provider.(TenantProjectionProvider)
	if !ok {
		return nil, nil
	}
	return projectionProvider, nil
}

// platformTenantInfo returns the neutral platform tenant projection.
func platformTenantInfo() *tenantcap.TenantInfo {
	return &tenantcap.TenantInfo{
		ID:     tenantcap.PLATFORM,
		Code:   "platform",
		Name:   "Platform",
		Status: "active",
	}
}

// normalizeTenantPage applies tenant search page defaults and max page size.
func normalizeTenantPage(page capmodel.PageRequest) capmodel.PageRequest {
	if page.PageNum <= 0 {
		page.PageNum = 1
	}
	pageSize := page.PageSize
	if pageSize <= 0 {
		pageSize = page.Limit
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > tenantcap.MaxTenantSearchPageSize {
		pageSize = tenantcap.MaxTenantSearchPageSize
	}
	page.PageSize = pageSize
	return page
}

// normalizeTenantIDs removes duplicates while preserving valid tenant IDs.
func normalizeTenantIDs(ids []tenantcap.TenantID) []tenantcap.TenantID {
	result := make([]tenantcap.TenantID, 0, len(ids))
	seen := make(map[tenantcap.TenantID]struct{}, len(ids))
	for _, id := range ids {
		if id < tenantcap.PLATFORM {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

// normalizePositiveUserIDs removes duplicate positive user identifiers.
func normalizePositiveUserIDs(ids []int) []int {
	result := make([]int, 0, len(ids))
	seen := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

// ListUserTenantProjections returns tenant ownership labels for visible users.
func (s *serviceImpl) ListUserTenantProjections(
	ctx context.Context,
	userIDs []int,
) (map[int]*tenantcap.UserTenantProjection, error) {
	result := make(map[int]*tenantcap.UserTenantProjection)
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
	requested []tenantcap.TenantID,
	mode tenantcap.UserTenantAssignmentMode,
) (*tenantcap.UserTenantAssignmentPlan, error) {
	provider, err := s.userMembershipProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return &tenantcap.UserTenantAssignmentPlan{PrimaryTenant: s.Current(ctx)}, nil
	}
	return provider.ResolveUserTenantAssignment(ctx, requested, mode)
}

// ReplaceUserTenantAssignments rewrites one user's active tenant ownership rows.
func (s *serviceImpl) ReplaceUserTenantAssignments(
	ctx context.Context,
	userID int,
	plan *tenantcap.UserTenantAssignmentPlan,
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
func (s *serviceImpl) EnsureUsersInTenant(ctx context.Context, userIDs []int, tenantID tenantcap.TenantID) error {
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
