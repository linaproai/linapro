// This file implements the governed distributed lock host service dispatcher.

package wasm

import (
	"context"

	"lina-core/internal/service/hostlock"
	bridgehostcall "lina-core/pkg/pluginbridge/hostcall"
	bridgehostservice "lina-core/pkg/pluginbridge/hostservice"
)

// lockHostService is the shared governed lock backend used by wasm host calls.
var lockHostService = hostlock.New()

// ConfigureLockHostService replaces the governed lock backend used by wasm
// host calls. The service must be non-nil.
func ConfigureLockHostService(service hostlock.Service) {
	if service == nil {
		panic("wasm lock host service requires a non-nil lock service")
	}
	lockHostService = service
}

// dispatchLockHostService routes lock host service methods to the governed lock backend.
func dispatchLockHostService(
	ctx context.Context,
	hcc *hostCallContext,
	resourceRef string,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	if hcc == nil || hcc.pluginID == "" {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, "host call context not available")
	}
	if resourceRef == "" {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusCapabilityDenied, "lock host service requires one authorized logical lock name")
	}
	tenantID := int32(0)
	if hcc.identity != nil {
		tenantID = hcc.identity.TenantId
	}
	normalizedTenantID := hostlock.TenantIDFromIdentity(tenantID)

	switch method {
	case bridgehostservice.HostServiceMethodLockAcquire:
		request, err := bridgehostservice.UnmarshalHostServiceLockAcquireRequest(payload)
		if err != nil {
			return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
		}
		output, callErr := lockHostService.Acquire(ctx, hostlock.AcquireInput{
			PluginID:    hcc.pluginID,
			TenantID:    normalizedTenantID,
			ResourceRef: resourceRef,
			LeaseMillis: request.LeaseMillis,
			RequestID:   hcc.requestID,
		})
		if callErr != nil {
			return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, callErr.Error())
		}
		response := &bridgehostservice.HostServiceLockAcquireResponse{Acquired: output.Acquired, Ticket: output.Ticket}
		if output.ExpireAt != nil {
			response.ExpireAt = output.ExpireAt.String()
		}
		return bridgehostcall.NewHostCallSuccessResponse(bridgehostservice.MarshalHostServiceLockAcquireResponse(response))
	case bridgehostservice.HostServiceMethodLockRenew:
		request, err := bridgehostservice.UnmarshalHostServiceLockRenewRequest(payload)
		if err != nil {
			return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
		}
		expireAt, callErr := lockHostService.Renew(ctx, hcc.pluginID, normalizedTenantID, resourceRef, request.Ticket)
		if callErr != nil {
			return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, callErr.Error())
		}
		response := &bridgehostservice.HostServiceLockRenewResponse{}
		if expireAt != nil {
			response.ExpireAt = expireAt.String()
		}
		return bridgehostcall.NewHostCallSuccessResponse(bridgehostservice.MarshalHostServiceLockRenewResponse(response))
	case bridgehostservice.HostServiceMethodLockRelease:
		request, err := bridgehostservice.UnmarshalHostServiceLockReleaseRequest(payload)
		if err != nil {
			return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
		}
		if callErr := lockHostService.Release(ctx, hcc.pluginID, normalizedTenantID, resourceRef, request.Ticket); callErr != nil {
			return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, callErr.Error())
		}
		return bridgehostcall.NewHostCallEmptySuccessResponse()
	default:
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			"unsupported lock host service method: "+method,
		)
	}
}
