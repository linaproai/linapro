// This file implements the governed distributed cache host service dispatcher.

package wasm

import (
	"context"
	"time"

	"lina-core/pkg/plugin/capability/cachecap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchCacheHostService routes cache host service methods to the scoped domain service.
func dispatchCacheHostService(
	ctx context.Context,
	hcc *hostCallContext,
	namespace string,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	if hcc == nil || hcc.pluginID == "" {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, "host call context not available")
	}
	if namespace == "" {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusCapabilityDenied, "cache host service requires one authorized namespace")
	}
	services := capabilityServicesForHostCall(hcc)
	if services == nil {
		return domainServiceNotScoped("cache")
	}
	service := services.Cache()
	if service == nil {
		return domainServiceNotScoped("cache")
	}

	switch method {
	case bridgehostservice.HostServiceMethodCacheGet:
		request, err := bridgehostservice.UnmarshalHostServiceCacheGetRequest(payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		item, found, callErr := service.Get(ctx, namespace, request.Key)
		if callErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, callErr)
		}
		response := &bridgehostservice.HostServiceCacheGetResponse{Found: found}
		if found {
			response.Value = buildCacheValueResponse(item)
		}
		return bridgehostcall.NewHostCallSuccessResponse(bridgehostservice.MarshalHostServiceCacheGetResponse(response))
	case bridgehostservice.HostServiceMethodCacheGetMany:
		var request cacheGetManyRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		output, callErr := service.GetMany(ctx, cachecap.GetManyInput{
			Namespace: namespace,
			Keys:      request.Keys,
		})
		if callErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, callErr)
		}
		return capabilityJSONResponse(cacheGetManyResponseFromDomain(output))
	case bridgehostservice.HostServiceMethodCacheSet:
		request, err := bridgehostservice.UnmarshalHostServiceCacheSetRequest(payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		item, callErr := service.Set(ctx, namespace, request.Key, request.Value, ttlFromSeconds(request.ExpireSeconds))
		if callErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, callErr)
		}
		return bridgehostcall.NewHostCallSuccessResponse(bridgehostservice.MarshalHostServiceCacheSetResponse(&bridgehostservice.HostServiceCacheSetResponse{
			Value: buildCacheValueResponse(item),
		}))
	case bridgehostservice.HostServiceMethodCacheSetMany:
		var request cacheSetManyRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		input := cachecap.SetManyInput{Namespace: namespace, Items: make([]cachecap.SetManyItem, 0, len(request.Items))}
		for _, item := range request.Items {
			input.Items = append(input.Items, cachecap.SetManyItem{
				Key:   item.Key,
				Value: item.Value,
				TTL:   ttlFromSeconds(item.ExpireSeconds),
			})
		}
		output, callErr := service.SetMany(ctx, input)
		if callErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, callErr)
		}
		return capabilityJSONResponse(cacheSetManyResponseFromDomain(output))
	case bridgehostservice.HostServiceMethodCacheDelete:
		request, err := bridgehostservice.UnmarshalHostServiceCacheDeleteRequest(payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		if callErr := service.Delete(ctx, namespace, request.Key); callErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, callErr)
		}
		return bridgehostcall.NewHostCallEmptySuccessResponse()
	case bridgehostservice.HostServiceMethodCacheDeleteMany:
		var request cacheDeleteManyRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		if callErr := service.DeleteMany(ctx, cachecap.DeleteManyInput{Namespace: namespace, Keys: request.Keys}); callErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, callErr)
		}
		return bridgehostcall.NewHostCallEmptySuccessResponse()
	case bridgehostservice.HostServiceMethodCacheIncr:
		request, err := bridgehostservice.UnmarshalHostServiceCacheIncrRequest(payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		item, callErr := service.Incr(ctx, namespace, request.Key, request.Delta, ttlFromSeconds(request.ExpireSeconds))
		if callErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, callErr)
		}
		return bridgehostcall.NewHostCallSuccessResponse(bridgehostservice.MarshalHostServiceCacheIncrResponse(&bridgehostservice.HostServiceCacheIncrResponse{
			Value: buildCacheValueResponse(item),
		}))
	case bridgehostservice.HostServiceMethodCacheExpire:
		request, err := bridgehostservice.UnmarshalHostServiceCacheExpireRequest(payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		found, expireAt, callErr := service.Expire(ctx, namespace, request.Key, ttlFromSeconds(request.ExpireSeconds))
		if callErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, callErr)
		}
		response := &bridgehostservice.HostServiceCacheExpireResponse{Found: found}
		if expireAt != nil {
			response.ExpireAt = expireAt.UTC().Format(time.RFC3339Nano)
		}
		return bridgehostcall.NewHostCallSuccessResponse(bridgehostservice.MarshalHostServiceCacheExpireResponse(response))
	default:
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			"unsupported cache host service method: "+method,
		)
	}
}

// ttlFromSeconds converts bridge payload seconds to the domain time.Duration contract.
func ttlFromSeconds(seconds int64) time.Duration {
	if seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

// buildCacheValueResponse maps one cache item into the protobuf response model.
func buildCacheValueResponse(item *cachecap.CacheItem) *bridgehostservice.HostServiceCacheValue {
	if item == nil {
		return nil
	}

	value := &bridgehostservice.HostServiceCacheValue{
		ValueKind: int32(item.ValueKind),
		Value:     item.Value,
		IntValue:  item.IntValue,
	}
	if item.ExpireAt != nil {
		value.ExpireAt = item.ExpireAt.UTC().Format(time.RFC3339Nano)
	}
	return value
}

type cacheGetManyRequest struct {
	Keys []string `json:"keys"`
}

type cacheGetManyResponse struct {
	Items       map[string]*bridgehostservice.HostServiceCacheValue `json:"items"`
	MissingKeys []string                                             `json:"missingKeys,omitempty"`
}

type cacheSetManyRequest struct {
	Items []cacheSetManyItemRequest `json:"items"`
}

type cacheSetManyItemRequest struct {
	Key           string `json:"key"`
	Value         string `json:"value"`
	ExpireSeconds int64  `json:"expireSeconds,omitempty"`
}

type cacheSetManyResponse struct {
	Items map[string]*bridgehostservice.HostServiceCacheValue `json:"items"`
}

type cacheDeleteManyRequest struct {
	Keys []string `json:"keys"`
}

func cacheGetManyResponseFromDomain(output *cachecap.GetManyOutput) cacheGetManyResponse {
	response := cacheGetManyResponse{Items: map[string]*bridgehostservice.HostServiceCacheValue{}}
	if output == nil {
		return response
	}
	response.MissingKeys = append([]string(nil), output.MissingKeys...)
	for key, item := range output.Items {
		response.Items[key] = buildCacheValueResponse(item)
	}
	return response
}

func cacheSetManyResponseFromDomain(output *cachecap.SetManyOutput) cacheSetManyResponse {
	response := cacheSetManyResponse{Items: map[string]*bridgehostservice.HostServiceCacheValue{}}
	if output == nil {
		return response
	}
	for key, item := range output.Items {
		response.Items[key] = buildCacheValueResponse(item)
	}
	return response
}
