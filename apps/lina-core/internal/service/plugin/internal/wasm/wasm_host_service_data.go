// This file implements the governed structured data host service dispatcher.

package wasm

import (
	"context"

	"lina-core/internal/service/plugin/internal/datahost"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchDataHostService routes governed data service methods to the structured data host layer.
func dispatchDataHostService(
	ctx context.Context,
	hcc *hostCallContext,
	table string,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	if hcc == nil {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusInternalError,
			"host call context not available",
		)
	}
	serviceSpec := hcc.hostServiceSpec(bridgehostservice.HostServiceData)
	if serviceSpec == nil {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusCapabilityDenied,
			"data host service authorization snapshot not found",
		)
	}
	resource, err := datahost.BuildCachedAuthorizedTableContract(ctx, datahost.AuthorizedTableContractInput{
		PluginID: hcc.pluginID,
		Table:    table,
		Methods:  serviceSpec.Methods,
	})
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}

	var (
		responsePayload []byte
		execErr         error
		orgSvc          = orgServiceForHostCall(hcc)
	)
	switch method {
	case bridgehostservice.HostServiceMethodDataList:
		request, decodeErr := bridgehostservice.UnmarshalHostServiceDataListRequest(payload)
		if decodeErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, decodeErr)
		}
		response, callErr := datahost.ExecuteList(
			ctx,
			hcc.pluginID,
			table,
			hcc.executionSource,
			hcc.identity,
			orgSvc,
			resource,
			request,
		)
		if callErr != nil {
			execErr = callErr
			break
		}
		responsePayload = bridgehostservice.MarshalHostServiceDataListResponse(response)
	case bridgehostservice.HostServiceMethodDataGet:
		request, decodeErr := bridgehostservice.UnmarshalHostServiceDataGetRequest(payload)
		if decodeErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, decodeErr)
		}
		response, callErr := datahost.ExecuteGet(
			ctx,
			hcc.pluginID,
			table,
			hcc.executionSource,
			hcc.identity,
			orgSvc,
			resource,
			request,
		)
		if callErr != nil {
			execErr = callErr
			break
		}
		responsePayload = bridgehostservice.MarshalHostServiceDataGetResponse(response)
	case bridgehostservice.HostServiceMethodDataCreate:
		request, decodeErr := bridgehostservice.UnmarshalHostServiceDataMutationRequest(payload)
		if decodeErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, decodeErr)
		}
		response, callErr := datahost.ExecuteCreate(
			ctx,
			hcc.pluginID,
			table,
			hcc.executionSource,
			hcc.identity,
			orgSvc,
			resource,
			request,
		)
		if callErr != nil {
			execErr = callErr
			break
		}
		responsePayload = bridgehostservice.MarshalHostServiceDataMutationResponse(response)
	case bridgehostservice.HostServiceMethodDataUpdate:
		request, decodeErr := bridgehostservice.UnmarshalHostServiceDataMutationRequest(payload)
		if decodeErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, decodeErr)
		}
		response, callErr := datahost.ExecuteUpdate(
			ctx,
			hcc.pluginID,
			table,
			hcc.executionSource,
			hcc.identity,
			orgSvc,
			resource,
			request,
		)
		if callErr != nil {
			execErr = callErr
			break
		}
		responsePayload = bridgehostservice.MarshalHostServiceDataMutationResponse(response)
	case bridgehostservice.HostServiceMethodDataDelete:
		request, decodeErr := bridgehostservice.UnmarshalHostServiceDataMutationRequest(payload)
		if decodeErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, decodeErr)
		}
		response, callErr := datahost.ExecuteDelete(
			ctx,
			hcc.pluginID,
			table,
			hcc.executionSource,
			hcc.identity,
			orgSvc,
			resource,
			request,
		)
		if callErr != nil {
			execErr = callErr
			break
		}
		responsePayload = bridgehostservice.MarshalHostServiceDataMutationResponse(response)
	case bridgehostservice.HostServiceMethodDataTransaction:
		request, decodeErr := bridgehostservice.UnmarshalHostServiceDataTransactionRequest(payload)
		if decodeErr != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, decodeErr)
		}
		response, callErr := datahost.ExecuteTransaction(
			ctx,
			hcc.pluginID,
			table,
			hcc.executionSource,
			hcc.identity,
			orgSvc,
			resource,
			request,
		)
		if callErr != nil {
			execErr = callErr
			break
		}
		responsePayload = bridgehostservice.MarshalHostServiceDataTransactionResponse(response)
	default:
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			"unsupported data host service method: "+method,
		)
	}
	if execErr != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, execErr)
	}
	return bridgehostcall.NewHostCallSuccessResponse(responsePayload)
}
