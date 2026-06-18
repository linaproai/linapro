// This file adapts dictionary host-service calls to the shared dict capability
// service.

package wasm

import (
	"context"
	"strings"

	"lina-core/pkg/plugin/capability/capmodel"
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
	capCtx := capabilityContextForHostCall(hcc, bridgehostservice.HostServiceDict, method)
	switch method {
	case bridgehostservice.HostServiceMethodDictResolveLabels:
		var request dictResolveRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.ResolveLabels(ctx, capCtx, dictcap.ResolveInput{
			Type:         dictcap.Type(request.Type),
			Values:       dictValues(request.Values),
			IncludeLabel: request.IncludeLabel,
		})
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodDictListValues:
		var request dictListValuesRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.ListValues(ctx, capCtx, dictcap.ListValuesInput{
			Type:         dictcap.Type(request.Type),
			Status:       request.Status,
			IncludeLabel: request.IncludeLabel,
			Page: capmodel.PageRequest{
				PageNum:  request.PageNum,
				PageSize: request.PageSize,
			},
		})
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodDictEnsureValuesVisible:
		var request dictResolveRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := service.EnsureValuesVisible(ctx, capCtx, dictcap.ResolveInput{
			Type:         dictcap.Type(request.Type),
			Values:       dictValues(request.Values),
			IncludeLabel: request.IncludeLabel,
		})
		return domainCapabilityResult(true, err)
	default:
		return domainMethodNotFound("dict", method)
	}
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

// dictListValuesRequest carries dictionary candidate listing input.
type dictListValuesRequest struct {
	Type         string `json:"type"`
	Status       *int   `json:"status,omitempty"`
	IncludeLabel bool   `json:"includeLabel"`
	PageNum      int    `json:"pageNum,omitempty"`
	PageSize     int    `json:"pageSize,omitempty"`
}

// dictValues converts transport strings into typed dictionary values.
func dictValues(values []string) []dictcap.Value {
	out := make([]dictcap.Value, 0, len(values))
	for _, value := range values {
		out = append(out, dictcap.Value(strings.TrimSpace(value)))
	}
	return out
}
