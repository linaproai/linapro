// This file adapts dynamic-plugin tenant host-service calls to the ordinary
// tenantcap.Service consumer contract. The dispatcher intentionally excludes
// host-internal HTTP request resolution, database query builders, membership
// write seams, and lifecycle governance services.

package wasm

import (
	"context"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/plugincap"
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
		return capabilityJSONResponse(service.Context().Current(ctx))
	case bridgehostservice.HostServiceMethodTenantCurrentInfo:
		result, err := service.Context().Info(ctx)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodTenantPlatformBypass:
		return capabilityJSONResponse(service.Context().PlatformBypass(ctx))
	case bridgehostservice.HostServiceMethodTenantBatchGet:
		var request tenantIDsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		result, err := service.Directory().BatchGet(ctx, tenantIDsFromInts(request.TenantIDs))
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodTenantDirectoryList:
		var request tenantcap.ListInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		result, err := service.Directory().List(ctx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodTenantValidateUserInTenant:
		var request userTenantRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		if response := ensureHostCallUsersVisible(ctx, hcc, bridgehostservice.HostServiceTenant, method, []int{request.UserID}); response != nil {
			return response
		}
		if err := service.Membership().Validate(ctx, request.UserID, tenantcap.TenantID(request.TenantID)); err != nil {
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
		tenants, err := service.Membership().ListByUser(ctx, request.UserID)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		return capabilityJSONResponse(tenants)
	case bridgehostservice.HostServiceMethodTenantBatchEnsureVisible:
		var request tenantIDsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		err := service.Directory().EnsureVisible(ctx, tenantIDsFromInts(request.TenantIDs))
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodTenantPluginSetEnabled:
		pluginSvc := service.Plugins()
		if pluginSvc == nil {
			return domainServiceNotScoped("tenant.plugins")
		}
		var request tenantPluginSetEnabledRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		err := pluginSvc.SetTenantPluginEnabled(ctx, plugincap.PluginID(request.PluginID), request.Enabled)
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodTenantPluginProvisionDefaults:
		pluginSvc := service.Plugins()
		if pluginSvc == nil {
			return domainServiceNotScoped("tenant.plugins")
		}
		var request tenantPluginProvisionDefaultsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		err := pluginSvc.ProvisionTenantPluginDefaults(ctx, capmodel.DomainID(request.TenantID))
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodTenantFilterContext:
		filterSvc := service.Filter()
		if filterSvc == nil {
			return domainServiceNotScoped("tenant.filter")
		}
		return capabilityJSONResponse(filterSvc.Context(ctx))
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

// intUserIDRequest carries one user identifier.
type intUserIDRequest struct {
	// UserID is the user identifier.
	UserID int `json:"userId"`
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

// tenantIDsFromInts converts transport tenant identifiers to domain IDs.
func tenantIDsFromInts(ids []int) []tenantcap.TenantID {
	out := make([]tenantcap.TenantID, 0, len(ids))
	for _, id := range ids {
		out = append(out, tenantcap.TenantID(id))
	}
	return out
}
