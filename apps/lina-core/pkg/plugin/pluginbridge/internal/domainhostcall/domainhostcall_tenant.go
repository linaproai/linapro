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

var _ tenantcap.Service = (*tenantService)(nil)

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
