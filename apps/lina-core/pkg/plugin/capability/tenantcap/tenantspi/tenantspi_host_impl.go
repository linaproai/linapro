// This file implements tenant provider delegation, provider status conversion,
// and host fallback helpers. It checks source-plugin enablement before
// forwarding tenant isolation, membership, and query-scope operations, returning
// platform-safe defaults when multi-tenancy is not installed or not enabled.

package tenantspi

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	internalregistry "lina-core/pkg/plugin/capability/internal/capabilityregistry"
	"lina-core/pkg/plugin/capability/tenantcap"
)

// Available reports whether an active tenant provider is available.
func (s *serviceImpl) Available(ctx context.Context) bool {
	if s == nil {
		return false
	}
	return s.manager.registry.StatusWithProvider(ctx, tenantcap.CapabilityTenantV1, s.enablement, s.providerEnv).Available
}

// Status returns the current tenant capability activation state.
func (s *serviceImpl) Status(ctx context.Context) capmodel.CapabilityStatus {
	if s == nil {
		return convertCapabilityStatus(NewManager().registry.Status(ctx, tenantcap.CapabilityTenantV1, nil))
	}
	return convertCapabilityStatus(s.manager.registry.StatusWithProvider(ctx, tenantcap.CapabilityTenantV1, s.enablement, s.providerEnv))
}

// convertCapabilityStatus copies internal capability state into public DTOs.
func convertCapabilityStatus(status internalregistry.CapabilityStatus) capmodel.CapabilityStatus {
	providers := make([]capmodel.ProviderStatus, 0, len(status.Providers))
	for _, provider := range status.Providers {
		providers = append(providers, convertProviderStatus(provider))
	}
	return capmodel.CapabilityStatus{
		CapabilityID:   status.CapabilityID,
		Available:      status.Available,
		ActiveProvider: status.ActiveProvider,
		Reason:         status.Reason,
		Providers:      providers,
	}
}

// convertProviderStatus copies one internal provider state into a public DTO.
func convertProviderStatus(status internalregistry.ProviderStatus) capmodel.ProviderStatus {
	return capmodel.ProviderStatus{
		CapabilityID: status.CapabilityID,
		PluginID:     status.PluginID,
		Active:       status.Active,
		Conflict:     status.Conflict,
		Reason:       status.Reason,
	}
}

// Context returns current-tenant context operations.
func (s *serviceImpl) Context() tenantcap.ContextService {
	return tenantContextService{root: s}
}

// Directory returns tenant directory operations.
func (s *serviceImpl) Directory() tenantcap.DirectoryService {
	return tenantDirectoryService{root: s}
}

// Membership returns user-to-tenant membership operations.
func (s *serviceImpl) Membership() tenantcap.MembershipService {
	return tenantMembershipService{root: s}
}

// Plugins returns tenant-plugin governance operations when the host directory injects them.
func (s *serviceImpl) Plugins() tenantcap.PluginService {
	return nil
}

// Filter returns plugin-visible tenant filter context reads.
func (s *serviceImpl) Filter() tenantcap.FilterService {
	if s == nil {
		return nil
	}
	return tenantFilterContextService{root: s}
}

// tenantContextService delegates tenant context reads to the root runtime service.
type tenantContextService struct {
	root *serviceImpl
}

// tenantDirectoryService delegates tenant directory operations to the root runtime service.
type tenantDirectoryService struct {
	root *serviceImpl
}

// tenantMembershipService delegates tenant membership operations to the root runtime service.
type tenantMembershipService struct {
	root *serviceImpl
}

// tenantFilterContextService delegates tenant filter context reads to the root runtime service.
type tenantFilterContextService struct {
	root *serviceImpl
}

// currentProvider returns the currently usable tenant-capability provider.
func (s *serviceImpl) currentProvider(ctx context.Context) (Provider, error) {
	if s == nil {
		return nil, nil
	}
	provider, err := s.manager.registry.ActiveProviderWithError(ctx, tenantcap.CapabilityTenantV1, s.enablement, s.providerEnv)
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
func (s *serviceImpl) providerEnv(ctx context.Context, pluginID string) ProviderEnv {
	env := defaultProviderEnv(ctx, pluginID)
	if s != nil && s.envFactory != nil {
		env = s.envFactory(ctx, pluginID)
	}
	if env.PluginID == "" {
		env.PluginID = pluginID
	}
	return env
}

// Current returns the current request tenant.
func (s tenantContextService) Current(ctx context.Context) tenantcap.TenantID {
	return s.root.Current(ctx)
}

// Current returns the current request tenant from bizctx, defaulting to platform.
func (s *serviceImpl) Current(ctx context.Context) tenantcap.TenantID {
	if s == nil || s.bizCtxSvc == nil {
		return tenantcap.PLATFORM
	}
	current := s.bizCtxSvc.Current(ctx)
	return tenantcap.TenantID(current.TenantID)
}

// Info returns the current request tenant projection.
func (s tenantContextService) Info(ctx context.Context) (*tenantcap.TenantInfo, error) {
	return s.root.currentTenantInfo(ctx)
}

// currentTenantInfo returns the current request tenant projection.
func (s *serviceImpl) currentTenantInfo(ctx context.Context) (*tenantcap.TenantInfo, error) {
	current := s.Current(ctx)
	if current <= tenantcap.PLATFORM {
		return platformTenantInfo(), nil
	}
	provider, err := s.tenantDirectoryProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return &tenantcap.TenantInfo{ID: current}, nil
	}
	return provider.Info(ctx, current)
}

// PlatformBypass reports whether the current request may bypass tenant filtering.
func (s tenantContextService) PlatformBypass(ctx context.Context) bool {
	return s.root.PlatformBypass(ctx)
}

// Context returns plugin-visible tenant and audit metadata.
func (s tenantFilterContextService) Context(ctx context.Context) tenantcap.TenantFilterContext {
	if s.root == nil || s.root.bizCtxSvc == nil {
		return tenantcap.TenantFilterContext{}
	}
	return tenantFilterContextFromCurrent(s.root.bizCtxSvc.Current(ctx))
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

// EnsureVisible validates that the current user can access tenant identifiers.
func (s tenantDirectoryService) EnsureVisible(ctx context.Context, tenantIDs []tenantcap.TenantID) error {
	return s.root.ensureTenantsVisible(ctx, tenantIDs)
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

// ensureTenantsVisible validates that the current user can access tenant identifiers.
func (s *serviceImpl) ensureTenantsVisible(ctx context.Context, tenantIDs []tenantcap.TenantID) error {
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
	provider, err := s.tenantDirectoryProvider(ctx)
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
	return provider.EnsureVisible(ctx, normalized)
}

// Validate verifies that a user can access a tenant.
func (s tenantMembershipService) Validate(ctx context.Context, userID int, tenantID tenantcap.TenantID) error {
	return s.root.ValidateUserInTenant(ctx, userID, tenantID)
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

// ListByUser returns active tenant memberships visible to one user.
func (s tenantMembershipService) ListByUser(ctx context.Context, userID int) ([]tenantcap.TenantInfo, error) {
	return s.root.ListUserTenants(ctx, userID)
}

// batchGetTenants returns visible tenant projections and opaque missing IDs.
func (s *serviceImpl) batchGetTenants(
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
	provider, err := s.tenantDirectoryProvider(ctx)
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
	return provider.BatchGet(ctx, normalized)
}

// Get returns one visible tenant projection.
func (s tenantDirectoryService) Get(ctx context.Context, tenantID tenantcap.TenantID) (*tenantcap.TenantInfo, error) {
	result, err := s.root.batchGetTenants(ctx, []tenantcap.TenantID{tenantID})
	if err != nil {
		return nil, err
	}
	if result == nil || result.Items[tenantID] == nil {
		return nil, bizerr.NewCode(tenantcap.CodeTenantForbidden, bizerr.P("tenantId", int(tenantID)))
	}
	return result.Items[tenantID], nil
}

// BatchGet returns visible tenant projections and opaque missing IDs.
func (s tenantDirectoryService) BatchGet(
	ctx context.Context,
	tenantIDs []tenantcap.TenantID,
) (*capmodel.BatchResult[*tenantcap.TenantInfo, tenantcap.TenantID], error) {
	return s.root.batchGetTenants(ctx, tenantIDs)
}

// searchTenants returns bounded tenant candidates visible to the caller.
func (s *serviceImpl) searchTenants(
	ctx context.Context,
	input tenantcap.ListInput,
) (*capmodel.PageResult[*tenantcap.TenantInfo], error) {
	provider, err := s.tenantDirectoryProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return &capmodel.PageResult[*tenantcap.TenantInfo]{Items: []*tenantcap.TenantInfo{}}, nil
	}
	input.Page = normalizeTenantPage(input.Page)
	return provider.List(ctx, input)
}

// List returns bounded tenant candidates visible to the caller.
func (s tenantDirectoryService) List(
	ctx context.Context,
	input tenantcap.ListInput,
) (*capmodel.PageResult[*tenantcap.TenantInfo], error) {
	return s.root.searchTenants(ctx, input)
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

// tenantDirectoryProvider returns the optional tenant projection capability facet.
func (s *serviceImpl) tenantDirectoryProvider(ctx context.Context) (DirectoryProvider, error) {
	provider, err := s.currentProvider(ctx)
	if err != nil || provider == nil {
		return nil, err
	}
	projectionProvider, ok := provider.(DirectoryProvider)
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

// ListUserTenantMemberships returns tenant ownership labels for visible users.
func (s *serviceImpl) ListUserTenantMemberships(
	ctx context.Context,
	userIDs []int,
) (map[int]*tenantcap.TenantMembershipInfo, error) {
	result := make(map[int]*tenantcap.TenantMembershipInfo)
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
	return provider.ListUserTenantMemberships(ctx, userIDs)
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
	provisioningProvider, ok := provider.(interface {
		ProvisionAutoEnabledTenantPlugins(ctx context.Context) error
	})
	if !ok {
		return nil
	}
	return provisioningProvider.ProvisionAutoEnabledTenantPlugins(ctx)
}

// IsProviderEnabled always returns false.
func (noopEnablementReader) IsProviderEnabled(_ context.Context, _ string) bool {
	return false
}
