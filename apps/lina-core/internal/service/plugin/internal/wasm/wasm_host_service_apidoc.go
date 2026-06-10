// This file adapts API-documentation host-service calls to the shared apidoc
// capability service.

package wasm

import (
	"context"

	"lina-core/pkg/plugin/capability/apidoccap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchAPIDocHostService routes API-documentation domain host-service calls.
func dispatchAPIDocHostService(
	ctx context.Context,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	service := apiDocServiceForHostCall(hcc)
	if service == nil {
		return domainServiceNotScoped("apidoc")
	}
	switch method {
	case bridgehostservice.HostServiceMethodAPIDocResolveRouteText:
		var request apiDocRouteTextRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		return capabilityJSONResponse(service.ResolveRouteText(ctx, request.toInput()))
	case bridgehostservice.HostServiceMethodAPIDocResolveRouteTexts:
		var request apiDocRouteTextsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		inputs := make([]apidoccap.RouteTextInput, 0, len(request.Inputs))
		for _, input := range request.Inputs {
			inputs = append(inputs, input.toInput())
		}
		return capabilityJSONResponse(service.ResolveRouteTexts(ctx, inputs))
	case bridgehostservice.HostServiceMethodAPIDocFindRouteTitleOperationKeys:
		var request keywordRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		return capabilityJSONResponse(service.FindRouteTitleOperationKeys(ctx, request.Keyword))
	default:
		return domainMethodNotFound("apidoc", method)
	}
}

// apiDocServiceForHostCall resolves the API documentation service for one host call.
func apiDocServiceForHostCall(hcc *hostCallContext) apidoccap.Service {
	services := capabilityServicesForHostCall(hcc)
	if services == nil {
		return nil
	}
	return services.APIDoc()
}

// keywordRequest carries one keyword search term.
type keywordRequest struct {
	Keyword string `json:"keyword"`
}

// apiDocRouteTextRequest carries one API route text resolution request.
type apiDocRouteTextRequest struct {
	OperationKey    string `json:"operationKey"`
	Method          string `json:"method"`
	Path            string `json:"path"`
	FallbackTitle   string `json:"fallbackTitle"`
	FallbackSummary string `json:"fallbackSummary"`
}

// apiDocRouteTextsRequest carries batch API route text resolution requests.
type apiDocRouteTextsRequest struct {
	Inputs []apiDocRouteTextRequest `json:"inputs"`
}

// toInput converts the transport request into the API documentation capability input.
func (r apiDocRouteTextRequest) toInput() apidoccap.RouteTextInput {
	return apidoccap.RouteTextInput{
		OperationKey:    r.OperationKey,
		Method:          r.Method,
		Path:            r.Path,
		FallbackTitle:   r.FallbackTitle,
		FallbackSummary: r.FallbackSummary,
	}
}
