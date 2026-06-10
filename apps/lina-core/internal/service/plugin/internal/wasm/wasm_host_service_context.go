// This file adapts context-oriented host-service calls to trusted host request
// state and shared business-context capability services.

package wasm

import (
	"context"
	"strings"

	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/routecap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchBizCtxHostService routes business-context host-service calls.
func dispatchBizCtxHostService(
	ctx context.Context,
	hcc *hostCallContext,
	method string,
	_ []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	service := bizCtxServiceForHostCall(hcc)
	if service == nil {
		return domainServiceNotScoped("bizctx")
	}
	if method != bridgehostservice.HostServiceMethodBizCtxCurrent {
		return domainMethodNotFound("bizctx", method)
	}
	return capabilityJSONResponse(service.Current(ctx))
}

// bizCtxServiceForHostCall resolves the business context service for one host call.
func bizCtxServiceForHostCall(hcc *hostCallContext) bizctxcap.Service {
	services := capabilityServicesForHostCall(hcc)
	if services == nil {
		return nil
	}
	return services.BizCtx()
}

// dispatchRouteHostService routes current dynamic-route metadata reads.
func dispatchRouteHostService(
	_ context.Context,
	hcc *hostCallContext,
	method string,
	_ []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	if method != bridgehostservice.HostServiceMethodRouteMetadataGet {
		return domainMethodNotFound("route", method)
	}
	return capabilityJSONResponse(routeMetadataFromHostCall(hcc))
}

// routeMetadataFromHostCall projects trusted host-call context into route metadata.
func routeMetadataFromHostCall(hcc *hostCallContext) *routecap.DynamicRouteMetadata {
	if hcc == nil {
		return nil
	}
	return &routecap.DynamicRouteMetadata{
		PluginID:   strings.TrimSpace(hcc.pluginID),
		PublicPath: strings.TrimSpace(hcc.routePath),
		Meta: map[string]string{
			"executionSource": strings.TrimSpace(string(hcc.executionSource)),
			"requestId":       strings.TrimSpace(hcc.requestID),
		},
	}
}
