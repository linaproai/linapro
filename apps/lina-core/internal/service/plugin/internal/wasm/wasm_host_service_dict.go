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
	"lina-core/pkg/statusflag"
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
	switch method {
	case bridgehostservice.HostServiceMethodDictRefresh:
		var request dictTypeRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := service.Refresh(ctx, dictcap.Type(strings.TrimSpace(request.Type)))
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodDictTypeGet:
		typeService := service.Type()
		if typeService == nil {
			return domainServiceNotScoped("dict.type")
		}
		var request dictIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := typeService.Get(ctx, request.ID)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodDictTypeBatchGet:
		typeService := service.Type()
		if typeService == nil {
			return domainServiceNotScoped("dict.type")
		}
		var request dictIDsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := typeService.BatchGet(ctx, append([]int(nil), request.IDs...))
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodDictTypeList:
		typeService := service.Type()
		if typeService == nil {
			return domainServiceNotScoped("dict.type")
		}
		var request dictcap.ListTypesInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := typeService.List(ctx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodDictTypeEnsureVisible:
		typeService := service.Type()
		if typeService == nil {
			return domainServiceNotScoped("dict.type")
		}
		var request dictIDsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := typeService.EnsureVisible(ctx, append([]int(nil), request.IDs...))
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodDictTypeEnsureKeysVisible:
		typeService := service.Type()
		if typeService == nil {
			return domainServiceNotScoped("dict.type")
		}
		var request dictTypeKeysRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := typeService.EnsureKeysVisible(ctx, dictTypes(request.Keys))
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodDictTypeCreate:
		typeService := service.Type()
		if typeService == nil {
			return domainServiceNotScoped("dict.type")
		}
		var request dictcap.CreateTypeInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := typeService.Create(ctx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodDictTypeUpdate:
		typeService := service.Type()
		if typeService == nil {
			return domainServiceNotScoped("dict.type")
		}
		var request dictcap.UpdateTypeInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := typeService.Update(ctx, request)
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodDictTypeDelete:
		typeService := service.Type()
		if typeService == nil {
			return domainServiceNotScoped("dict.type")
		}
		var request dictIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := typeService.Delete(ctx, request.ID)
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodDictValueGet:
		valueService := service.Value()
		if valueService == nil {
			return domainServiceNotScoped("dict.value")
		}
		var request dictIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := valueService.Get(ctx, request.ID)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodDictValueBatchGet:
		valueService := service.Value()
		if valueService == nil {
			return domainServiceNotScoped("dict.value")
		}
		var request dictcap.BatchGetValuesInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := valueService.BatchGet(ctx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodDictValueResolveLabels:
		valueService := service.Value()
		if valueService == nil {
			return domainServiceNotScoped("dict.value")
		}
		var request dictResolveRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := valueService.ResolveLabels(ctx, dictcap.ResolveInput{
			Type:         dictcap.Type(request.Type),
			Values:       dictValues(request.Values),
			IncludeLabel: request.IncludeLabel,
		})
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodDictListValues:
		valueService := service.Value()
		if valueService == nil {
			return domainServiceNotScoped("dict.value")
		}
		var request dictListValuesRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := valueService.List(ctx, dictcap.ListValuesInput{
			Type:         dictcap.Type(request.Type),
			Status:       dictStatusFlag(request.Status),
			IncludeLabel: request.IncludeLabel,
			Page: capmodel.PageRequest{
				PageNum:  request.PageNum,
				PageSize: request.PageSize,
			},
		})
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodDictValueEnsureVisible:
		valueService := service.Value()
		if valueService == nil {
			return domainServiceNotScoped("dict.value")
		}
		var request dictIDsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := valueService.EnsureVisible(ctx, append([]int(nil), request.IDs...))
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodDictValueEnsureValuesVisible:
		valueService := service.Value()
		if valueService == nil {
			return domainServiceNotScoped("dict.value")
		}
		var request dictResolveRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := valueService.EnsureValuesVisible(ctx, dictcap.ResolveInput{
			Type:         dictcap.Type(request.Type),
			Values:       dictValues(request.Values),
			IncludeLabel: request.IncludeLabel,
		})
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodDictValueCreate:
		valueService := service.Value()
		if valueService == nil {
			return domainServiceNotScoped("dict.value")
		}
		var request dictcap.CreateValueInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := valueService.Create(ctx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodDictValueUpdate:
		valueService := service.Value()
		if valueService == nil {
			return domainServiceNotScoped("dict.value")
		}
		var request dictcap.UpdateValueInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := valueService.Update(ctx, request)
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodDictValueDelete:
		valueService := service.Value()
		if valueService == nil {
			return domainServiceNotScoped("dict.value")
		}
		var request dictIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := valueService.Delete(ctx, request.ID)
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodDictValueDeleteByType:
		valueService := service.Value()
		if valueService == nil {
			return domainServiceNotScoped("dict.value")
		}
		var request dictTypeRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := valueService.DeleteByType(ctx, dictcap.Type(strings.TrimSpace(request.Type)))
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

type dictIDRequest struct {
	ID int `json:"id"`
}

type dictIDsRequest struct {
	IDs []int `json:"ids"`
}

type dictTypeRequest struct {
	Type string `json:"type"`
}

type dictTypeKeysRequest struct {
	Keys []string `json:"keys"`
}

// dictValues converts transport strings into typed dictionary values.
func dictValues(values []string) []dictcap.Value {
	out := make([]dictcap.Value, 0, len(values))
	for _, value := range values {
		out = append(out, dictcap.Value(strings.TrimSpace(value)))
	}
	return out
}

// dictTypes converts transport strings into typed dictionary type keys.
func dictTypes(values []string) []dictcap.Type {
	out := make([]dictcap.Type, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			out = append(out, dictcap.Type(trimmed))
		}
	}
	return out
}

// dictStatusFlag converts the wire optional integer status into the shared capability status type.
func dictStatusFlag(status *int) *statusflag.Enabled {
	if status == nil {
		return nil
	}
	value := statusflag.Enabled(*status)
	return &value
}
