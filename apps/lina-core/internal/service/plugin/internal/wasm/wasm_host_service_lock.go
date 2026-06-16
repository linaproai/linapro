// This file implements the governed distributed lock host service dispatcher.

package wasm

import (
	"context"
	"time"

	"lina-core/pkg/plugin/capability/lockcap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchLockHostService routes lock host service methods to the scoped domain service.
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
	services := capabilityServicesForHostCall(hcc)
	if services == nil {
		return domainServiceNotScoped("lock")
	}
	service := services.Lock()
	if service == nil {
		return domainServiceNotScoped("lock")
	}

	switch method {
	case bridgehostservice.HostServiceMethodLockAcquire:
		request, err := bridgehostservice.UnmarshalHostServiceLockAcquireRequest(payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		output, callErr := service.Acquire(ctx, lockcap.AcquireInput{
			Name:  resourceRef,
			Lease: time.Duration(request.LeaseMillis) * time.Millisecond,
		})
		if callErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, callErr)
		}
		response := &bridgehostservice.HostServiceLockAcquireResponse{Acquired: output.Acquired, Ticket: output.Ticket}
		if output.ExpireAt != nil {
			response.ExpireAt = output.ExpireAt.UTC().Format(time.RFC3339Nano)
		}
		return bridgehostcall.NewHostCallSuccessResponse(bridgehostservice.MarshalHostServiceLockAcquireResponse(response))
	case bridgehostservice.HostServiceMethodLockRenew:
		request, err := bridgehostservice.UnmarshalHostServiceLockRenewRequest(payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		output, callErr := service.Renew(ctx, lockcap.RenewInput{Name: resourceRef, Ticket: request.Ticket})
		if callErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, callErr)
		}
		response := &bridgehostservice.HostServiceLockRenewResponse{}
		if output != nil && output.ExpireAt != nil {
			response.ExpireAt = output.ExpireAt.UTC().Format(time.RFC3339Nano)
		}
		return bridgehostcall.NewHostCallSuccessResponse(bridgehostservice.MarshalHostServiceLockRenewResponse(response))
	case bridgehostservice.HostServiceMethodLockRelease:
		request, err := bridgehostservice.UnmarshalHostServiceLockReleaseRequest(payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		if callErr := service.Release(ctx, lockcap.ReleaseInput{Name: resourceRef, Ticket: request.Ticket}); callErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, callErr)
		}
		return bridgehostcall.NewHostCallEmptySuccessResponse()
	default:
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			"unsupported lock host service method: "+method,
		)
	}
}
