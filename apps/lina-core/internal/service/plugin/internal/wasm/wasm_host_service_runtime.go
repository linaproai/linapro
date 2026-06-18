// This file implements the runtime host service backed by existing host log
// and plugin-scoped state handlers plus lightweight runtime info methods.

package wasm

import (
	"context"
	"os"
	"strconv"

	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/guid"

	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchRuntimeHostService routes runtime host service methods to log, state,
// and lightweight runtime info handlers.
func dispatchRuntimeHostService(
	ctx context.Context,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	switch method {
	case bridgehostservice.HostServiceMethodRuntimeLogWrite:
		return handleHostLog(ctx, hcc, payload)
	case bridgehostservice.HostServiceMethodRuntimeStateGet:
		return handleHostStateGet(ctx, hcc, payload)
	case bridgehostservice.HostServiceMethodRuntimeStateGetMany:
		return handleHostStateGetMany(ctx, hcc, payload)
	case bridgehostservice.HostServiceMethodRuntimeStateSet:
		return handleHostStateSet(ctx, hcc, payload)
	case bridgehostservice.HostServiceMethodRuntimeStateSetMany:
		return handleHostStateSetMany(ctx, hcc, payload)
	case bridgehostservice.HostServiceMethodRuntimeStateDelete:
		return handleHostStateDelete(ctx, hcc, payload)
	case bridgehostservice.HostServiceMethodRuntimeStateDeleteMany:
		return handleHostStateDeleteMany(ctx, hcc, payload)
	case bridgehostservice.HostServiceMethodRuntimeInfoNow:
		return buildRuntimeInfoValueResponse(strconv.FormatInt(gtime.Now().Time.UnixMilli(), 10))
	case bridgehostservice.HostServiceMethodRuntimeInfoUUID:
		return buildRuntimeInfoValueResponse(guid.S())
	case bridgehostservice.HostServiceMethodRuntimeInfoNode:
		nodeName, err := os.Hostname()
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
		}
		return buildRuntimeInfoValueResponse(nodeName)
	default:
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			"unsupported runtime host service method: "+method,
		)
	}
}

// buildRuntimeInfoValueResponse wraps one scalar runtime info value in a success envelope.
func buildRuntimeInfoValueResponse(value string) *bridgehostcall.HostCallResponseEnvelope {
	payload := bridgehostservice.MarshalHostServiceValueResponse(&bridgehostservice.HostServiceValueResponse{
		Value: value,
	})
	return bridgehostcall.NewHostCallSuccessResponse(payload)
}
