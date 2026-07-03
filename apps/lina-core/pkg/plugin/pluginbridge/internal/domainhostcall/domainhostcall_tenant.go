// This file implements guest-side tenant capability reads that cross the
// pluginbridge host-service transport. Status and current-scope reads follow
// source-plugin fallback semantics by returning zero values when transport is
// unavailable.

package domainhostcall

import (
	"context"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// tenantService adapts tenant capability reads to host services.
type tenantService struct{ baseService }

// tenantContextService adapts tenant context reads to host services.
type tenantContextService struct{ baseService }

// tenantDirectoryService adapts tenant directory operations to host services.
type tenantDirectoryService struct{ baseService }

// tenantMembershipService adapts tenant membership operations to host services.
type tenantMembershipService struct{ baseService }

// tenantPluginService adapts tenant-plugin governance operations to host services.
type tenantPluginService struct{ baseService }

// tenantFilterService adapts tenant-filter context reads to host services.
type tenantFilterService struct{ baseService }

// Tenant creates the tenant capability guest client.
func Tenant(invoker Invoker) tenantcap.Service {
	return tenantService{baseService: newBaseService(invoker)}
}

// Status returns the current tenant capability activation state.
func (s tenantService) Status(_ context.Context) capmodel.CapabilityStatus {
	var status capmodel.CapabilityStatus
	if err := s.call(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantStatus,
		nil,
		&status,
	); err != nil {
		return capmodel.CapabilityStatus{}
	}
	return status
}

// Available reports whether the tenant capability has an active provider.
func (s tenantService) Available(_ context.Context) bool {
	var available bool
	if err := s.call(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantAvailable,
		nil,
		&available,
	); err != nil {
		return false
	}
	return available
}

// Context returns current-tenant context operations.
func (s tenantService) Context() tenantcap.ContextService {
	return tenantContextService{baseService: s.baseService}
}

// Directory returns tenant directory operations.
func (s tenantService) Directory() tenantcap.DirectoryService {
	return tenantDirectoryService{baseService: s.baseService}
}

// Membership returns user-to-tenant membership operations.
func (s tenantService) Membership() tenantcap.MembershipService {
	return tenantMembershipService{baseService: s.baseService}
}

// Plugins returns tenant-plugin governance operations.
func (s tenantService) Plugins() tenantcap.PluginService {
	return tenantPluginService{baseService: s.baseService}
}

// Filter returns tenant filter context reads.
func (s tenantService) Filter() tenantcap.FilterService {
	return tenantFilterService{baseService: s.baseService}
}

// Current returns the current request tenant.
func (s tenantContextService) Current(_ context.Context) tenantcap.TenantID {
	var tenantID tenantcap.TenantID
	if err := s.call(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantCurrent,
		nil,
		&tenantID,
	); err != nil {
		return tenantcap.PlatformTenantID
	}
	return tenantID
}

// Info returns the current request tenant projection.
func (s tenantContextService) Info(_ context.Context) (*tenantcap.TenantInfo, error) {
	out := &tenantcap.TenantInfo{}
	err := s.call(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantCurrentInfo,
		nil,
		out,
	)
	return out, err
}

// PlatformBypass reports whether the current request may bypass tenant filtering.
func (s tenantContextService) PlatformBypass(_ context.Context) bool {
	var bypass bool
	if err := s.call(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantPlatformBypass,
		nil,
		&bypass,
	); err != nil {
		return false
	}
	return bypass
}

// Get returns one visible tenant projection.
func (s tenantDirectoryService) Get(ctx context.Context, tenantID tenantcap.TenantID) (*tenantcap.TenantInfo, error) {
	result, err := s.BatchGet(ctx, []tenantcap.TenantID{tenantID})
	if err != nil {
		return nil, err
	}
	if result == nil || result.Items[tenantID] == nil {
		return nil, nil
	}
	return result.Items[tenantID], nil
}

// EnsureVisible validates that the current user can access tenant identifiers.
func (s tenantDirectoryService) EnsureVisible(_ context.Context, tenantIDs []tenantcap.TenantID) error {
	return s.callJSONRequest(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantBatchEnsureVisible,
		tenantIDsRequest{TenantIDs: tenantIDsToInts(tenantIDs)},
		nil,
	)
}

// BatchGet returns visible tenant projections and opaque missing IDs.
func (s tenantDirectoryService) BatchGet(_ context.Context, tenantIDs []tenantcap.TenantID) (*capmodel.BatchResult[*tenantcap.TenantInfo, tenantcap.TenantID], error) {
	out := &capmodel.BatchResult[*tenantcap.TenantInfo, tenantcap.TenantID]{Items: map[tenantcap.TenantID]*tenantcap.TenantInfo{}}
	err := s.callJSONRequest(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantBatchGet,
		tenantIDsRequest{TenantIDs: tenantIDsToInts(tenantIDs)},
		out,
	)
	return out, err
}

// List returns bounded tenant candidates visible to the caller.
func (s tenantDirectoryService) List(_ context.Context, input tenantcap.ListInput) (*capmodel.PageResult[*tenantcap.TenantInfo], error) {
	out := &capmodel.PageResult[*tenantcap.TenantInfo]{Items: []*tenantcap.TenantInfo{}}
	err := s.callJSONRequest(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantDirectoryList,
		input,
		out,
	)
	return out, err
}

// Validate verifies that a user can access tenantID.
func (s tenantMembershipService) Validate(_ context.Context, userID int, tenantID tenantcap.TenantID) error {
	return s.callJSONRequest(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantValidateUserInTenant,
		userTenantRequest{UserID: userID, TenantID: int(tenantID)},
		nil,
	)
}

// ListByUser returns active tenants visible to one user.
func (s tenantMembershipService) ListByUser(_ context.Context, userID int) ([]tenantcap.TenantInfo, error) {
	var tenants []tenantcap.TenantInfo
	err := s.callJSONRequest(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantListUserTenants,
		intUserIDRequest{UserID: userID},
		&tenants,
	)
	return tenants, err
}

// SetTenantPluginEnabled updates one tenant plugin enablement row.
func (s tenantPluginService) SetTenantPluginEnabled(_ context.Context, pluginID plugincap.PluginID, enabled bool) error {
	return s.callJSONRequest(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantPluginSetEnabled,
		tenantPluginSetEnabledRequest{PluginID: string(pluginID), Enabled: enabled},
		nil,
	)
}

// ProvisionTenantPluginDefaults creates missing default plugin rows for one tenant.
func (s tenantPluginService) ProvisionTenantPluginDefaults(_ context.Context, tenantID capmodel.DomainID) error {
	return s.callJSONRequest(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantPluginProvisionDefaults,
		tenantPluginProvisionDefaultsRequest{TenantID: string(tenantID)},
		nil,
	)
}

// Context returns plugin-visible tenant and audit metadata.
func (s tenantFilterService) Context(context.Context) tenantcap.TenantFilterContext {
	var out tenantcap.TenantFilterContext
	if err := s.call(protocol.HostServiceTenant, protocol.HostServiceMethodTenantFilterContext, nil, &out); err != nil {
		return tenantcap.TenantFilterContext{}
	}
	return out
}

// tenantIDsRequest carries multiple tenant identifiers.
type tenantIDsRequest struct {
	// TenantIDs are the tenant identifiers.
	TenantIDs []int `json:"tenantIds"`
}

// intUserIDRequest carries one integer user identifier.
type intUserIDRequest struct {
	// UserID is the user identifier.
	UserID int `json:"userId"`
}

// userTenantRequest carries one user and tenant pair.
type userTenantRequest struct {
	// UserID is the user identifier.
	UserID int `json:"userId"`
	// TenantID is the tenant identifier.
	TenantID int `json:"tenantId"`
}

// tenantPluginSetEnabledRequest carries a tenant-plugin enablement update.
type tenantPluginSetEnabledRequest struct {
	// PluginID is the plugin identifier.
	PluginID string `json:"pluginId"`
	// Enabled is the requested tenant plugin enablement state.
	Enabled bool `json:"enabled"`
}

// tenantPluginProvisionDefaultsRequest carries one tenant default-provisioning target.
type tenantPluginProvisionDefaultsRequest struct {
	// TenantID is the tenant identifier.
	TenantID string `json:"tenantId"`
}

// tenantIDsToInts converts tenant IDs to transport integers.
func tenantIDsToInts(ids []tenantcap.TenantID) []int {
	out := make([]int, 0, len(ids))
	for _, id := range ids {
		out = append(out, int(id))
	}
	return out
}

var (
	_ tenantcap.Service       = (*tenantService)(nil)
	_ tenantcap.PluginService = (*tenantPluginService)(nil)
	_ tenantcap.FilterService = (*tenantFilterService)(nil)
)
