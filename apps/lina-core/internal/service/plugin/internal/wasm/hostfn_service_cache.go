// This file implements the governed distributed cache host service dispatcher.

package wasm

import (
	"context"

	"lina-core/internal/service/kvcache"
	"lina-core/pkg/pluginbridge"
)

var cacheHostService = kvcache.New()

func dispatchCacheHostService(
	ctx context.Context,
	hcc *hostCallContext,
	namespace string,
	method string,
	payload []byte,
) *pluginbridge.HostCallResponseEnvelope {
	if hcc == nil || hcc.pluginID == "" {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, "host call context not available")
	}
	if namespace == "" {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusCapabilityDenied, "cache host service requires one authorized namespace")
	}

	switch method {
	case pluginbridge.HostServiceMethodCacheGet:
		request, err := pluginbridge.UnmarshalHostServiceCacheGetRequest(payload)
		if err != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
		}
		item, found, callErr := cacheHostService.Get(ctx, kvcache.OwnerTypePlugin, hcc.pluginID, namespace, request.Key)
		if callErr != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, callErr.Error())
		}
		response := &pluginbridge.HostServiceCacheGetResponse{Found: found}
		if found {
			response.Value = buildCacheValueResponse(item)
		}
		return pluginbridge.NewHostCallSuccessResponse(pluginbridge.MarshalHostServiceCacheGetResponse(response))
	case pluginbridge.HostServiceMethodCacheSet:
		request, err := pluginbridge.UnmarshalHostServiceCacheSetRequest(payload)
		if err != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
		}
		item, callErr := cacheHostService.Set(ctx, kvcache.OwnerTypePlugin, hcc.pluginID, namespace, request.Key, request.Value, request.ExpireSeconds)
		if callErr != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, callErr.Error())
		}
		return pluginbridge.NewHostCallSuccessResponse(pluginbridge.MarshalHostServiceCacheSetResponse(&pluginbridge.HostServiceCacheSetResponse{
			Value: buildCacheValueResponse(item),
		}))
	case pluginbridge.HostServiceMethodCacheDelete:
		request, err := pluginbridge.UnmarshalHostServiceCacheDeleteRequest(payload)
		if err != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
		}
		if callErr := cacheHostService.Delete(ctx, kvcache.OwnerTypePlugin, hcc.pluginID, namespace, request.Key); callErr != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, callErr.Error())
		}
		return pluginbridge.NewHostCallEmptySuccessResponse()
	case pluginbridge.HostServiceMethodCacheIncr:
		request, err := pluginbridge.UnmarshalHostServiceCacheIncrRequest(payload)
		if err != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
		}
		item, callErr := cacheHostService.Incr(ctx, kvcache.OwnerTypePlugin, hcc.pluginID, namespace, request.Key, request.Delta, request.ExpireSeconds)
		if callErr != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, callErr.Error())
		}
		return pluginbridge.NewHostCallSuccessResponse(pluginbridge.MarshalHostServiceCacheIncrResponse(&pluginbridge.HostServiceCacheIncrResponse{
			Value: buildCacheValueResponse(item),
		}))
	case pluginbridge.HostServiceMethodCacheExpire:
		request, err := pluginbridge.UnmarshalHostServiceCacheExpireRequest(payload)
		if err != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
		}
		found, expireAt, callErr := cacheHostService.Expire(ctx, kvcache.OwnerTypePlugin, hcc.pluginID, namespace, request.Key, request.ExpireSeconds)
		if callErr != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, callErr.Error())
		}
		response := &pluginbridge.HostServiceCacheExpireResponse{Found: found}
		if expireAt != nil {
			response.ExpireAt = expireAt.String()
		}
		return pluginbridge.NewHostCallSuccessResponse(pluginbridge.MarshalHostServiceCacheExpireResponse(response))
	default:
		return pluginbridge.NewHostCallErrorResponse(
			pluginbridge.HostCallStatusNotFound,
			"unsupported cache host service method: "+method,
		)
	}
}

func buildCacheValueResponse(item *kvcache.Item) *pluginbridge.HostServiceCacheValue {
	if item == nil {
		return nil
	}

	value := &pluginbridge.HostServiceCacheValue{
		ValueKind: int32(item.ValueKind),
		Value:     item.Value,
		IntValue:  item.IntValue,
	}
	if item.ExpireAt != nil {
		value.ExpireAt = item.ExpireAt.String()
	}
	return value
}
