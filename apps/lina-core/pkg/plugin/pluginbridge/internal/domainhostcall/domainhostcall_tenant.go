// This file implements guest-side tenant capability reads that cross the
// pluginbridge host-service transport. Status and current-scope reads follow
// source-plugin fallback semantics by returning zero values when transport is
// unavailable.

package domainhostcall

import (
	"context"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// tenantService adapts tenant capability reads to host services.
type tenantService struct{ baseService }

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

// Current returns the current request tenant.
func (s tenantService) Current(_ context.Context) tenantcap.TenantID {
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

// CurrentTenantInfo returns the current request tenant projection.
func (s tenantService) CurrentTenantInfo(_ context.Context) (*tenantcap.TenantInfo, error) {
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
func (s tenantService) PlatformBypass(_ context.Context) bool {
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

// EnsureTenantVisible validates that the current user can access tenantID.
func (s tenantService) EnsureTenantVisible(_ context.Context, tenantID tenantcap.TenantID) error {
	return s.callJSONRequest(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantEnsureVisible,
		tenantIDRequest{TenantID: int(tenantID)},
		nil,
	)
}

// BatchGetTenants returns visible tenant projections and opaque missing IDs.
func (s tenantService) BatchGetTenants(_ context.Context, tenantIDs []tenantcap.TenantID) (*capmodel.BatchResult[*tenantcap.TenantInfo, tenantcap.TenantID], error) {
	out := &capmodel.BatchResult[*tenantcap.TenantInfo, tenantcap.TenantID]{Items: map[tenantcap.TenantID]*tenantcap.TenantInfo{}}
	err := s.callJSONRequest(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantBatchGet,
		tenantIDsRequest{TenantIDs: tenantIDsToInts(tenantIDs)},
		out,
	)
	return out, err
}

// SearchTenants returns bounded tenant candidates visible to the caller.
func (s tenantService) SearchTenants(_ context.Context, input tenantcap.SearchInput) (*capmodel.PageResult[*tenantcap.TenantInfo], error) {
	out := &capmodel.PageResult[*tenantcap.TenantInfo]{Items: []*tenantcap.TenantInfo{}}
	err := s.callJSONRequest(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantSearch,
		input,
		out,
	)
	return out, err
}

// ValidateUserInTenant verifies that a user can access tenantID.
func (s tenantService) ValidateUserInTenant(_ context.Context, userID int, tenantID tenantcap.TenantID) error {
	return s.callJSONRequest(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantValidateUserInTenant,
		userTenantRequest{UserID: userID, TenantID: int(tenantID)},
		nil,
	)
}

// ListUserTenants returns active tenants visible to one user.
func (s tenantService) ListUserTenants(_ context.Context, userID int) ([]tenantcap.TenantInfo, error) {
	var tenants []tenantcap.TenantInfo
	err := s.callJSONRequest(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantListUserTenants,
		intUserIDRequest{UserID: userID},
		&tenants,
	)
	return tenants, err
}

// BatchListUserTenants returns active tenant memberships for visible users.
func (s tenantService) BatchListUserTenants(_ context.Context, userIDs []int) (map[int][]tenantcap.TenantInfo, error) {
	out := make(map[int][]tenantcap.TenantInfo)
	err := s.callJSONRequest(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantBatchListUserTenants,
		intUserIDsRequest{UserIDs: userIDs},
		&out,
	)
	return out, err
}

// EnsureTenantsVisible validates that the current user can access every tenant.
func (s tenantService) EnsureTenantsVisible(_ context.Context, tenantIDs []tenantcap.TenantID) error {
	return s.callJSONRequest(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantBatchEnsureVisible,
		tenantIDsRequest{TenantIDs: tenantIDsToInts(tenantIDs)},
		nil,
	)
}

// SwitchTenant validates a tenant switch before token re-issue.
func (s tenantService) SwitchTenant(_ context.Context, userID int, target tenantcap.TenantID) error {
	return s.callJSONRequest(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantValidateSwitch,
		tenantSwitchRequest{UserID: userID, TargetTenantID: int(target)},
		nil,
	)
}

// tenantIDRequest carries one tenant identifier.
type tenantIDRequest struct {
	// TenantID is the tenant identifier.
	TenantID int `json:"tenantId"`
}

// tenantIDsRequest carries multiple tenant identifiers.
type tenantIDsRequest struct {
	// TenantIDs are the tenant identifiers.
	TenantIDs []int `json:"tenantIds"`
}

// userTenantRequest carries one user and tenant pair.
type userTenantRequest struct {
	// UserID is the user identifier.
	UserID int `json:"userId"`
	// TenantID is the tenant identifier.
	TenantID int `json:"tenantId"`
}

// tenantSwitchRequest carries one tenant switch check.
type tenantSwitchRequest struct {
	// UserID is the user identifier.
	UserID int `json:"userId"`
	// TargetTenantID is the requested tenant identifier.
	TargetTenantID int `json:"targetTenantId"`
}

// tenantIDsToInts converts tenant IDs to transport integers.
func tenantIDsToInts(ids []tenantcap.TenantID) []int {
	out := make([]int, 0, len(ids))
	for _, id := range ids {
		out = append(out, int(id))
	}
	return out
}

var _ tenantcap.Service = (*tenantService)(nil)
