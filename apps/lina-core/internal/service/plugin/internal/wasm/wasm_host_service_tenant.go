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
	case bridgehostservice.HostServiceMethodTenantCurrentInfo:
		result, err := service.CurrentTenantInfo(ctx)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodTenantPlatformBypass:
		return capabilityJSONResponse(service.PlatformBypass(ctx))
	case bridgehostservice.HostServiceMethodTenantEnsureVisible:
		var request tenantIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		if err := service.EnsureTenantVisible(ctx, tenantcap.TenantID(request.TenantID)); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return capabilityJSONResponse(true)
	case bridgehostservice.HostServiceMethodTenantBatchGet:
		var request tenantIDsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		result, err := service.BatchGetTenants(ctx, tenantIDsFromInts(request.TenantIDs))
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodTenantSearch:
		var request tenantcap.SearchInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		result, err := service.SearchTenants(ctx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodTenantValidateUserInTenant:
		var request userTenantRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		if response := ensureHostCallUsersVisible(ctx, hcc, bridgehostservice.HostServiceTenant, method, []int{request.UserID}); response != nil {
			return response
		}
		if err := service.ValidateUserInTenant(ctx, request.UserID, tenantcap.TenantID(request.TenantID)); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return capabilityJSONResponse(true)
	case bridgehostservice.HostServiceMethodTenantListUserTenants:
		var request intUserIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
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
	case bridgehostservice.HostServiceMethodTenantBatchListUserTenants:
		var request intUserIDsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		if response := ensureHostCallUsersVisible(ctx, hcc, bridgehostservice.HostServiceTenant, method, request.UserIDs); response != nil {
			return response
		}
		result, err := service.BatchListUserTenants(ctx, request.UserIDs)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodTenantBatchEnsureVisible:
		var request tenantIDsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		err := service.EnsureTenantsVisible(ctx, tenantIDsFromInts(request.TenantIDs))
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodTenantValidateSwitch:
		var request tenantSwitchRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		if response := ensureHostCallUsersVisible(ctx, hcc, bridgehostservice.HostServiceTenant, method, []int{request.UserID}); response != nil {
			return response
		}
		if err := service.SwitchTenant(ctx, request.UserID, tenantcap.TenantID(request.TargetTenantID)); err != nil {
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

// tenantIDsFromInts converts transport tenant identifiers to domain IDs.
func tenantIDsFromInts(ids []int) []tenantcap.TenantID {
	out := make([]tenantcap.TenantID, 0, len(ids))
	for _, id := range ids {
		out = append(out, tenantcap.TenantID(id))
	}
	return out
}
