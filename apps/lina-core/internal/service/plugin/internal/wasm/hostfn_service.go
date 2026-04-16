// This file implements the structured host service dispatcher used by the
// Wasm runtime host_call entrypoint.

package wasm

import (
	"context"
	"fmt"

	"lina-core/pkg/pluginbridge"
)

// handleHostServiceInvoke validates and dispatches one structured host service invocation.
func handleHostServiceInvoke(
	ctx context.Context,
	hcc *hostCallContext,
	reqBytes []byte,
) *pluginbridge.HostCallResponseEnvelope {
	request, err := pluginbridge.UnmarshalHostServiceRequestEnvelope(reqBytes)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}

	requiredCapability := pluginbridge.RequiredCapabilityForHostServiceMethod(request.Service, request.Method)
	if requiredCapability == "" {
		return pluginbridge.NewHostCallErrorResponse(
			pluginbridge.HostCallStatusNotFound,
			fmt.Sprintf("unsupported host service method: %s.%s", request.Service, request.Method),
		)
	}
	if !hcc.hasCapability(requiredCapability) {
		return pluginbridge.NewHostCallErrorResponse(
			pluginbridge.HostCallStatusCapabilityDenied,
			fmt.Sprintf("plugin %s lacks capability %s", hcc.pluginID, requiredCapability),
		)
	}
	if !hcc.hasHostServiceAccess(request.Service, request.Method, request.ResourceRef, request.Table) {
		return pluginbridge.NewHostCallErrorResponse(
			pluginbridge.HostCallStatusCapabilityDenied,
			fmt.Sprintf(
				"plugin %s is not authorized for host service %s.%s resource=%s table=%s",
				hcc.pluginID,
				request.Service,
				request.Method,
				request.ResourceRef,
				request.Table,
			),
		)
	}

	switch request.Service {
	case pluginbridge.HostServiceRuntime:
		return dispatchRuntimeHostService(ctx, hcc, request.Method, request.Payload)
	case pluginbridge.HostServiceStorage:
		return dispatchStorageHostService(ctx, hcc, request.ResourceRef, request.Method, request.Payload)
	case pluginbridge.HostServiceNetwork:
		return dispatchNetworkHostService(ctx, hcc, request.ResourceRef, request.Method, request.Payload)
	case pluginbridge.HostServiceData:
		return dispatchDataHostService(ctx, hcc, request.Table, request.Method, request.Payload)
	case pluginbridge.HostServiceCache:
		return dispatchCacheHostService(ctx, hcc, request.ResourceRef, request.Method, request.Payload)
	case pluginbridge.HostServiceLock:
		return dispatchLockHostService(ctx, hcc, request.ResourceRef, request.Method, request.Payload)
	case pluginbridge.HostServiceNotify:
		return dispatchNotifyHostService(ctx, hcc, request.ResourceRef, request.Method, request.Payload)
	default:
		return pluginbridge.NewHostCallErrorResponse(
			pluginbridge.HostCallStatusNotFound,
			fmt.Sprintf("host service not implemented yet: %s", request.Service),
		)
	}
}
