// This file implements guest-side tenant capability reads that cross the
// pluginbridge host-service transport. The wrapper is compiled for ordinary Go
// tests and wasip1 guests; only the lower-level transport selects the real host
// import or the unsupported stub.

package guest

import (
	"context"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// TenantService exposes guest-side tenant capability reads.
type TenantService interface {
	// Status returns the current tenant capability activation state. ctx is
	// accepted for parity with source-plugin capability services; the current
	// guest transport cannot cancel an in-flight WASI host call. The zero value is
	// returned with an error when transport or response decoding fails.
	Status(ctx context.Context) (capmodel.CapabilityStatus, error)
	// Available reports whether the tenant capability has an active provider. It
	// returns false with the Status error when the host call cannot complete or
	// returns an invalid response.
	Available(ctx context.Context) (bool, error)
	// Current returns the current request tenant.
	Current(ctx context.Context) (tenantcap.TenantID, error)
	// PlatformBypass reports whether the current request may bypass tenant filtering.
	PlatformBypass(ctx context.Context) (bool, error)
	// EnsureTenantVisible validates that the current user can access tenantID.
	EnsureTenantVisible(ctx context.Context, tenantID tenantcap.TenantID) error
	// ValidateUserInTenant verifies that a user can access tenantID.
	ValidateUserInTenant(ctx context.Context, userID int, tenantID tenantcap.TenantID) error
	// ListUserTenants returns active tenants visible to one user.
	ListUserTenants(ctx context.Context, userID int) ([]tenantcap.TenantInfo, error)
	// SwitchTenant validates a tenant switch before token re-issue.
	SwitchTenant(ctx context.Context, userID int, target tenantcap.TenantID) error
}

var _ TenantService = (*tenantService)(nil)

// tenantService implements guest tenant capability reads.
type tenantService struct{}

// Status returns the current tenant capability activation state.
func (tenantService) Status(_ context.Context) (capmodel.CapabilityStatus, error) {
	var status capmodel.CapabilityStatus
	err := invokeCapabilityJSON(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantStatus,
		nil,
		&status,
	)
	return status, err
}

// Available reports whether the tenant capability has an active provider.
func (tenantService) Available(_ context.Context) (bool, error) {
	var available bool
	err := invokeCapabilityJSON(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantAvailable,
		nil,
		&available,
	)
	return available, err
}

// Current returns the current request tenant.
func (tenantService) Current(_ context.Context) (tenantcap.TenantID, error) {
	var tenantID tenantcap.TenantID
	err := invokeCapabilityJSON(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantCurrent,
		nil,
		&tenantID,
	)
	return tenantID, err
}

// PlatformBypass reports whether the current request may bypass tenant filtering.
func (tenantService) PlatformBypass(_ context.Context) (bool, error) {
	var bypass bool
	err := invokeCapabilityJSON(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantPlatformBypass,
		nil,
		&bypass,
	)
	return bypass, err
}

// EnsureTenantVisible validates that the current user can access tenantID.
func (tenantService) EnsureTenantVisible(_ context.Context, tenantID tenantcap.TenantID) error {
	return invokeCapabilityJSON(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantEnsureVisible,
		protocol.MarshalHostServiceCapabilityTenantRequest(
			&protocol.HostServiceCapabilityTenantRequest{TenantID: int(tenantID)},
		),
		nil,
	)
}

// ValidateUserInTenant verifies that a user can access tenantID.
func (tenantService) ValidateUserInTenant(_ context.Context, userID int, tenantID tenantcap.TenantID) error {
	return invokeCapabilityJSON(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantValidateUserInTenant,
		protocol.MarshalHostServiceCapabilityUserTenantRequest(
			&protocol.HostServiceCapabilityUserTenantRequest{UserID: userID, TenantID: int(tenantID)},
		),
		nil,
	)
}

// ListUserTenants returns active tenants visible to one user.
func (tenantService) ListUserTenants(_ context.Context, userID int) ([]tenantcap.TenantInfo, error) {
	var tenants []tenantcap.TenantInfo
	err := invokeCapabilityJSON(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantListUserTenants,
		protocol.MarshalHostServiceCapabilityUserRequest(
			&protocol.HostServiceCapabilityUserRequest{UserID: userID},
		),
		&tenants,
	)
	return tenants, err
}

// SwitchTenant validates a tenant switch before token re-issue.
func (tenantService) SwitchTenant(_ context.Context, userID int, target tenantcap.TenantID) error {
	return invokeCapabilityJSON(
		protocol.HostServiceTenant,
		protocol.HostServiceMethodTenantValidateSwitch,
		protocol.MarshalHostServiceCapabilityUserTenantSwitchRequest(
			&protocol.HostServiceCapabilityUserTenantSwitchRequest{
				UserID:         userID,
				TargetTenantID: int(target),
			},
		),
		nil,
	)
}
