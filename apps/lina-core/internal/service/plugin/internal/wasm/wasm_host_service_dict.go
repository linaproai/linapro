// This file adapts dictionary host-service calls to the shared dict capability
// service.

package wasm

import (
	"context"
	"strings"

	"lina-core/pkg/plugin/capability/dictcap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchDictHostService routes dictionary-domain host-service calls.
func dispatchDictHostService(
	ctx context.Context,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	service := dictServiceForHostCall(hcc)
	if service == nil {
		return domainServiceNotScoped("dict")
	}
	if method != bridgehostservice.HostServiceMethodDictResolveLabels {
		return domainMethodNotFound("dict", method)
	}
	var request dictResolveRequest
	if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
		return invalidCapabilityRequest(err)
	}
	capCtx := capabilityContextForHostCall(hcc, bridgehostservice.HostServiceDict, method)
	result, err := service.ResolveLabels(ctx, capCtx, dictcap.ResolveInput{
		Type:         dictcap.Type(request.Type),
		Values:       dictValues(request.Values),
		IncludeLabel: request.IncludeLabel,
	})
	return domainCapabilityResult(result, err)
}

// dictServiceForHostCall resolves the dictionary service for one host call.
func dictServiceForHostCall(hcc *hostCallContext) dictcap.Service {
	services := capabilityServicesForHostCall(hcc)
	if services == nil {
		return nil
	}
	return services.Dict()
}

// dictResolveRequest carries dictionary label resolution input.
type dictResolveRequest struct {
	Type         string   `json:"type"`
	Values       []string `json:"values"`
	IncludeLabel bool     `json:"includeLabel"`
}

// dictValues converts transport strings into typed dictionary values.
func dictValues(values []string) []dictcap.Value {
	out := make([]dictcap.Value, 0, len(values))
	for _, value := range values {
		if normalized := strings.TrimSpace(value); normalized != "" {
			out = append(out, dictcap.Value(normalized))
		}
	}
	return out
}
