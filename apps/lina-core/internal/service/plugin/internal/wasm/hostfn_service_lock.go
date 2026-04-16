// This file implements the governed distributed lock host service dispatcher.

package wasm

import (
	"context"

	"lina-core/internal/service/hostlock"
	"lina-core/pkg/pluginbridge"
)

var lockHostService = hostlock.New()

func dispatchLockHostService(
	ctx context.Context,
	hcc *hostCallContext,
	resourceRef string,
	method string,
	payload []byte,
) *pluginbridge.HostCallResponseEnvelope {
	if hcc == nil || hcc.pluginID == "" {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, "host call context not available")
	}
	if resourceRef == "" {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusCapabilityDenied, "lock host service requires one authorized logical lock name")
	}

	switch method {
	case pluginbridge.HostServiceMethodLockAcquire:
		request, err := pluginbridge.UnmarshalHostServiceLockAcquireRequest(payload)
		if err != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
		}
		output, callErr := lockHostService.Acquire(ctx, hostlock.AcquireInput{
			PluginID:    hcc.pluginID,
			ResourceRef: resourceRef,
			LeaseMillis: request.LeaseMillis,
			RequestID:   hcc.requestID,
		})
		if callErr != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, callErr.Error())
		}
		response := &pluginbridge.HostServiceLockAcquireResponse{Acquired: output.Acquired, Ticket: output.Ticket}
		if output.ExpireAt != nil {
			response.ExpireAt = output.ExpireAt.String()
		}
		return pluginbridge.NewHostCallSuccessResponse(pluginbridge.MarshalHostServiceLockAcquireResponse(response))
	case pluginbridge.HostServiceMethodLockRenew:
		request, err := pluginbridge.UnmarshalHostServiceLockRenewRequest(payload)
		if err != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
		}
		expireAt, callErr := lockHostService.Renew(ctx, hcc.pluginID, resourceRef, request.Ticket)
		if callErr != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, callErr.Error())
		}
		response := &pluginbridge.HostServiceLockRenewResponse{}
		if expireAt != nil {
			response.ExpireAt = expireAt.String()
		}
		return pluginbridge.NewHostCallSuccessResponse(pluginbridge.MarshalHostServiceLockRenewResponse(response))
	case pluginbridge.HostServiceMethodLockRelease:
		request, err := pluginbridge.UnmarshalHostServiceLockReleaseRequest(payload)
		if err != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
		}
		if callErr := lockHostService.Release(ctx, hcc.pluginID, resourceRef, request.Ticket); callErr != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, callErr.Error())
		}
		return pluginbridge.NewHostCallEmptySuccessResponse()
	default:
		return pluginbridge.NewHostCallErrorResponse(
			pluginbridge.HostCallStatusNotFound,
			"unsupported lock host service method: "+method,
		)
	}
}
