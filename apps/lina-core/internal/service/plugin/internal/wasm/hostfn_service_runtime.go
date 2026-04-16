// This file implements the runtime host service backed by existing host log
// and plugin-scoped state handlers plus lightweight runtime info methods.

package wasm

import (
	"context"
	"os"

	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/guid"

	"lina-core/pkg/pluginbridge"
)

func dispatchRuntimeHostService(
	ctx context.Context,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *pluginbridge.HostCallResponseEnvelope {
	switch method {
	case pluginbridge.HostServiceMethodRuntimeLogWrite:
		return handleHostLog(ctx, hcc, payload)
	case pluginbridge.HostServiceMethodRuntimeStateGet:
		return handleHostStateGet(ctx, hcc, payload)
	case pluginbridge.HostServiceMethodRuntimeStateSet:
		return handleHostStateSet(ctx, hcc, payload)
	case pluginbridge.HostServiceMethodRuntimeStateDelete:
		return handleHostStateDelete(ctx, hcc, payload)
	case pluginbridge.HostServiceMethodRuntimeInfoNow:
		return buildRuntimeInfoValueResponse(gtime.Now().String())
	case pluginbridge.HostServiceMethodRuntimeInfoUUID:
		return buildRuntimeInfoValueResponse(guid.S())
	case pluginbridge.HostServiceMethodRuntimeInfoNode:
		nodeName, err := os.Hostname()
		if err != nil {
			return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
		}
		return buildRuntimeInfoValueResponse(nodeName)
	default:
		return pluginbridge.NewHostCallErrorResponse(
			pluginbridge.HostCallStatusNotFound,
			"unsupported runtime host service method: "+method,
		)
	}
}

func buildRuntimeInfoValueResponse(value string) *pluginbridge.HostCallResponseEnvelope {
	payload := pluginbridge.MarshalHostServiceValueResponse(&pluginbridge.HostServiceValueResponse{
		Value: value,
	})
	return pluginbridge.NewHostCallSuccessResponse(payload)
}
