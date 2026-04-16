// This file implements the governed structured data host service dispatcher.

package wasm

import (
	"context"

	"lina-core/internal/service/plugin/internal/datahost"
	"lina-core/pkg/pluginbridge"
)

func dispatchDataHostService(
	ctx context.Context,
	hcc *hostCallContext,
	table string,
	method string,
	payload []byte,
) *pluginbridge.HostCallResponseEnvelope {
	if hcc == nil {
		return pluginbridge.NewHostCallErrorResponse(
			pluginbridge.HostCallStatusInternalError,
			"host call context not available",
		)
	}
	serviceSpec := hcc.hostServiceSpec(pluginbridge.HostServiceData)
	if serviceSpec == nil {
		return pluginbridge.NewHostCallErrorResponse(
			pluginbridge.HostCallStatusCapabilityDenied,
			"data host service authorization snapshot not found",
		)
	}
	resource, err := datahost.BuildAuthorizedTableContract(ctx, table, serviceSpec.Methods)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}

	var (
		responsePayload []byte
		execErr         error
	)
	switch method {
	case pluginbridge.HostServiceMethodDataList:
		request, decodeErr := pluginbridge.UnmarshalHostServiceDataListRequest(payload)
		if decodeErr != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, decodeErr.Error())
		}
		response, callErr := datahost.ExecuteList(
			ctx,
			hcc.pluginID,
			table,
			hcc.executionSource,
			hcc.identity,
			resource,
			request,
		)
		if callErr != nil {
			execErr = callErr
			break
		}
		responsePayload = pluginbridge.MarshalHostServiceDataListResponse(response)
	case pluginbridge.HostServiceMethodDataGet:
		request, decodeErr := pluginbridge.UnmarshalHostServiceDataGetRequest(payload)
		if decodeErr != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, decodeErr.Error())
		}
		response, callErr := datahost.ExecuteGet(
			ctx,
			hcc.pluginID,
			table,
			hcc.executionSource,
			hcc.identity,
			resource,
			request,
		)
		if callErr != nil {
			execErr = callErr
			break
		}
		responsePayload = pluginbridge.MarshalHostServiceDataGetResponse(response)
	case pluginbridge.HostServiceMethodDataCreate:
		request, decodeErr := pluginbridge.UnmarshalHostServiceDataMutationRequest(payload)
		if decodeErr != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, decodeErr.Error())
		}
		response, callErr := datahost.ExecuteCreate(
			ctx,
			hcc.pluginID,
			table,
			hcc.executionSource,
			hcc.identity,
			resource,
			request,
		)
		if callErr != nil {
			execErr = callErr
			break
		}
		responsePayload = pluginbridge.MarshalHostServiceDataMutationResponse(response)
	case pluginbridge.HostServiceMethodDataUpdate:
		request, decodeErr := pluginbridge.UnmarshalHostServiceDataMutationRequest(payload)
		if decodeErr != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, decodeErr.Error())
		}
		response, callErr := datahost.ExecuteUpdate(
			ctx,
			hcc.pluginID,
			table,
			hcc.executionSource,
			hcc.identity,
			resource,
			request,
		)
		if callErr != nil {
			execErr = callErr
			break
		}
		responsePayload = pluginbridge.MarshalHostServiceDataMutationResponse(response)
	case pluginbridge.HostServiceMethodDataDelete:
		request, decodeErr := pluginbridge.UnmarshalHostServiceDataMutationRequest(payload)
		if decodeErr != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, decodeErr.Error())
		}
		response, callErr := datahost.ExecuteDelete(
			ctx,
			hcc.pluginID,
			table,
			hcc.executionSource,
			hcc.identity,
			resource,
			request,
		)
		if callErr != nil {
			execErr = callErr
			break
		}
		responsePayload = pluginbridge.MarshalHostServiceDataMutationResponse(response)
	case pluginbridge.HostServiceMethodDataTransaction:
		request, decodeErr := pluginbridge.UnmarshalHostServiceDataTransactionRequest(payload)
		if decodeErr != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, decodeErr.Error())
		}
		response, callErr := datahost.ExecuteTransaction(
			ctx,
			hcc.pluginID,
			table,
			hcc.executionSource,
			hcc.identity,
			resource,
			request,
		)
		if callErr != nil {
			execErr = callErr
			break
		}
		responsePayload = pluginbridge.MarshalHostServiceDataTransactionResponse(response)
	default:
		return pluginbridge.NewHostCallErrorResponse(
			pluginbridge.HostCallStatusNotFound,
			"unsupported data host service method: "+method,
		)
	}
	if execErr != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, execErr.Error())
	}
	return pluginbridge.NewHostCallSuccessResponse(responsePayload)
}
