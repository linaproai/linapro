// This file implements the structured host service dispatcher used by the
// Wasm runtime host_call entrypoint.

package wasm

import (
	"context"
	"fmt"

	bridgecontract "lina-core/pkg/pluginbridge/contract"
	bridgehostcall "lina-core/pkg/pluginbridge/hostcall"
	bridgehostservice "lina-core/pkg/pluginbridge/hostservice"
)

// handleHostServiceInvoke validates capability and authorization state before
// dispatching one structured host service invocation.
func handleHostServiceInvoke(
	ctx context.Context,
	hcc *hostCallContext,
	reqBytes []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceRequestEnvelope(reqBytes)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}

	requiredCapability := bridgehostservice.RequiredCapabilityForHostServiceMethod(request.Service, request.Method)
	if requiredCapability == "" {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			fmt.Sprintf("unsupported host service method: %s.%s", request.Service, request.Method),
		)
	}
	if !hcc.hasCapability(requiredCapability) {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusCapabilityDenied,
			fmt.Sprintf("plugin %s lacks capability %s", hcc.pluginID, requiredCapability),
		)
	}
	if !hcc.hasHostServiceAccess(request.Service, request.Method, request.ResourceRef, request.Table) {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusCapabilityDenied,
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
	if bridgecontract.NormalizeExecutionSource(hcc.executionSource) == bridgecontract.ExecutionSourceCronDiscovery &&
		request.Service != bridgehostservice.HostServiceCron {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusCapabilityDenied,
			fmt.Sprintf("cron discovery execution does not allow host service %s", request.Service),
		)
	}

	switch request.Service {
	case bridgehostservice.HostServiceRuntime:
		return dispatchRuntimeHostService(ctx, hcc, request.Method, request.Payload)
	case bridgehostservice.HostServiceCron:
		return dispatchCronHostService(ctx, hcc, request.Method, request.Payload)
	case bridgehostservice.HostServiceStorage:
		return dispatchStorageHostService(ctx, hcc, request.ResourceRef, request.Method, request.Payload)
	case bridgehostservice.HostServiceNetwork:
		return dispatchNetworkHostService(ctx, hcc, request.ResourceRef, request.Method, request.Payload)
	case bridgehostservice.HostServiceData:
		return dispatchDataHostService(ctx, hcc, request.Table, request.Method, request.Payload)
	case bridgehostservice.HostServiceCache:
		return dispatchCacheHostService(ctx, hcc, request.ResourceRef, request.Method, request.Payload)
	case bridgehostservice.HostServiceLock:
		return dispatchLockHostService(ctx, hcc, request.ResourceRef, request.Method, request.Payload)
	case bridgehostservice.HostServiceNotify:
		return dispatchNotifyHostService(ctx, hcc, request.ResourceRef, request.Method, request.Payload)
	case bridgehostservice.HostServiceConfig:
		return dispatchConfigHostService(ctx, hcc, request.Method, request.Payload)
	default:
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			fmt.Sprintf("host service not implemented yet: %s", request.Service),
		)
	}
}
