// This file adapts dynamic-plugin tenant host-service calls to the ordinary
// tenantcap.Service consumer contract. The dispatcher intentionally excludes
// host-internal HTTP request resolution, database query builders, membership
// write seams, and lifecycle governance services.

package wasm

import (
	"context"

	"lina-core/pkg/plugin/capability/tenantcap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchTenantHostService routes one tenant host-service method to the same
// ordinary tenantcap.Service surface exposed to source plugins.
func dispatchTenantHostService(
	ctx context.Context,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	service := tenantServiceForHostCall(hcc)
	if service == nil {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusInternalError,
			"tenant host service is not scoped",
		)
	}

	switch method {
	case bridgehostservice.HostServiceMethodTenantAvailable:
		return capabilityJSONResponse(service.Available(ctx))
	case bridgehostservice.HostServiceMethodTenantStatus:
		return capabilityJSONResponse(service.Status(ctx))
	case bridgehostservice.HostServiceMethodTenantCurrent:
		return capabilityJSONResponse(service.Current(ctx))
	case bridgehostservice.HostServiceMethodTenantPlatformBypass:
		return capabilityJSONResponse(service.PlatformBypass(ctx))
	case bridgehostservice.HostServiceMethodTenantEnsureVisible:
		request, err := bridgehostservice.UnmarshalHostServiceCapabilityTenantRequest(payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		if err = service.EnsureTenantVisible(ctx, tenantcap.TenantID(request.TenantID)); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return capabilityJSONResponse(true)
	case bridgehostservice.HostServiceMethodTenantValidateUserInTenant:
		request, err := bridgehostservice.UnmarshalHostServiceCapabilityUserTenantRequest(payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		if response := ensureHostCallUsersVisible(ctx, hcc, bridgehostservice.HostServiceTenant, method, []int{request.UserID}); response != nil {
			return response
		}
		if err = service.ValidateUserInTenant(ctx, request.UserID, tenantcap.TenantID(request.TenantID)); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return capabilityJSONResponse(true)
	case bridgehostservice.HostServiceMethodTenantListUserTenants:
		request, err := bridgehostservice.UnmarshalHostServiceCapabilityUserRequest(payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		if response := ensureHostCallUsersVisible(ctx, hcc, bridgehostservice.HostServiceTenant, method, []int{request.UserID}); response != nil {
			return response
		}
		tenants, err := service.ListUserTenants(ctx, request.UserID)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return capabilityJSONResponse(tenants)
	case bridgehostservice.HostServiceMethodTenantValidateSwitch:
		request, err := bridgehostservice.UnmarshalHostServiceCapabilityUserTenantSwitchRequest(payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		if response := ensureHostCallUsersVisible(ctx, hcc, bridgehostservice.HostServiceTenant, method, []int{request.UserID}); response != nil {
			return response
		}
		if err = service.SwitchTenant(ctx, request.UserID, tenantcap.TenantID(request.TargetTenantID)); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return capabilityJSONResponse(true)
	default:
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			"tenant host service method not implemented: "+method,
		)
	}
}

// tenantServiceForHostCall resolves the tenant service for one host call.
func tenantServiceForHostCall(hcc *hostCallContext) tenantcap.Service {
	services := capabilityServicesForHostCall(hcc)
	if services == nil {
		return nil
	}
	return services.Tenant()
}
