// This file implements the governed distributed cache host service dispatcher.

package wasm

import (
	"context"

	"lina-core/internal/service/kvcache"
	bridgehostcall "lina-core/pkg/pluginbridge/hostcall"
	bridgehostservice "lina-core/pkg/pluginbridge/hostservice"
)

// cacheHostService is the shared governed cache backend used by wasm host calls.
var cacheHostService = kvcache.New()

// dispatchCacheHostService routes cache host service methods to the governed cache backend.
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

	switch method {
	case bridgehostservice.HostServiceMethodCacheGet:
		request, err := bridgehostservice.UnmarshalHostServiceCacheGetRequest(payload)
		if err != nil {
			return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
		}
		item, found, callErr := cacheHostService.Get(
			ctx,
			kvcache.OwnerTypePlugin,
			kvcache.BuildCacheKey(hcc.pluginID, namespace, request.Key),
		)
		if callErr != nil {
			return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, callErr.Error())
		}
		response := &bridgehostservice.HostServiceCacheGetResponse{Found: found}
		if found {
			response.Value = buildCacheValueResponse(item)
		}
		return bridgehostcall.NewHostCallSuccessResponse(bridgehostservice.MarshalHostServiceCacheGetResponse(response))
	case bridgehostservice.HostServiceMethodCacheSet:
		request, err := bridgehostservice.UnmarshalHostServiceCacheSetRequest(payload)
		if err != nil {
			return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
		}
		item, callErr := cacheHostService.Set(
			ctx,
			kvcache.OwnerTypePlugin,
			kvcache.BuildCacheKey(hcc.pluginID, namespace, request.Key),
			request.Value,
			kvcache.TTLFromSeconds(request.ExpireSeconds),
		)
		if callErr != nil {
			return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, callErr.Error())
		}
		return bridgehostcall.NewHostCallSuccessResponse(bridgehostservice.MarshalHostServiceCacheSetResponse(&bridgehostservice.HostServiceCacheSetResponse{
			Value: buildCacheValueResponse(item),
		}))
	case bridgehostservice.HostServiceMethodCacheDelete:
		request, err := bridgehostservice.UnmarshalHostServiceCacheDeleteRequest(payload)
		if err != nil {
			return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
		}
		if callErr := cacheHostService.Delete(
			ctx,
			kvcache.OwnerTypePlugin,
			kvcache.BuildCacheKey(hcc.pluginID, namespace, request.Key),
		); callErr != nil {
			return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, callErr.Error())
		}
		return bridgehostcall.NewHostCallEmptySuccessResponse()
	case bridgehostservice.HostServiceMethodCacheIncr:
		request, err := bridgehostservice.UnmarshalHostServiceCacheIncrRequest(payload)
		if err != nil {
			return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
		}
		item, callErr := cacheHostService.Incr(
			ctx,
			kvcache.OwnerTypePlugin,
			kvcache.BuildCacheKey(hcc.pluginID, namespace, request.Key),
			request.Delta,
			kvcache.TTLFromSeconds(request.ExpireSeconds),
		)
		if callErr != nil {
			return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, callErr.Error())
		}
		return bridgehostcall.NewHostCallSuccessResponse(bridgehostservice.MarshalHostServiceCacheIncrResponse(&bridgehostservice.HostServiceCacheIncrResponse{
			Value: buildCacheValueResponse(item),
		}))
	case bridgehostservice.HostServiceMethodCacheExpire:
		request, err := bridgehostservice.UnmarshalHostServiceCacheExpireRequest(payload)
		if err != nil {
			return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
		}
		found, expireAt, callErr := cacheHostService.Expire(
			ctx,
			kvcache.OwnerTypePlugin,
			kvcache.BuildCacheKey(hcc.pluginID, namespace, request.Key),
			kvcache.TTLFromSeconds(request.ExpireSeconds),
		)
		if callErr != nil {
			return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, callErr.Error())
		}
		response := &bridgehostservice.HostServiceCacheExpireResponse{Found: found}
		if expireAt != nil {
			response.ExpireAt = expireAt.String()
		}
		return bridgehostcall.NewHostCallSuccessResponse(bridgehostservice.MarshalHostServiceCacheExpireResponse(response))
	default:
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			"unsupported cache host service method: "+method,
		)
	}
}

// buildCacheValueResponse maps one cache item into the protobuf response model.
func buildCacheValueResponse(item *kvcache.Item) *bridgehostservice.HostServiceCacheValue {
	if item == nil {
		return nil
	}

	value := &bridgehostservice.HostServiceCacheValue{
		ValueKind: int32(item.Kind),
		Value:     item.Value,
		IntValue:  item.IntValue,
	}
	if item.ExpireAt != nil {
		value.ExpireAt = item.ExpireAt.String()
	}
	return value
}
