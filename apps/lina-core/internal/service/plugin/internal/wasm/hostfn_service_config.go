// This file implements the read-only config host service for dynamic plugins.

package wasm

import (
	"context"
	"strconv"

	"github.com/gogf/gf/v2/encoding/gjson"

	bridgehostcall "lina-core/pkg/pluginbridge/hostcall"
	bridgehostservice "lina-core/pkg/pluginbridge/hostservice"
	configsvc "lina-core/pkg/pluginservice/config"
	"lina-core/pkg/pluginservice/contract"
)

// configHostService is the shared read-only configuration adapter used by wasm
// host calls.
var configHostService = configsvc.New()

// ConfigureConfigHostService replaces the read-only configuration adapter used
// by wasm host calls. The service must be non-nil.
func ConfigureConfigHostService(service contract.ConfigService) {
	if service == nil {
		panic("wasm config host service requires a non-nil config adapter")
	}
	configHostService = service
}

// dispatchConfigHostService routes config host service methods to the generic
// read-only plugin configuration service.
func dispatchConfigHostService(
	ctx context.Context,
	_ *hostCallContext,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceConfigKeyRequest(payload)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, err.Error())
	}

	switch method {
	case bridgehostservice.HostServiceMethodConfigGet:
		return handleConfigGet(ctx, configHostService, request.Key)
	case bridgehostservice.HostServiceMethodConfigExists:
		return handleConfigExists(ctx, configHostService, request.Key)
	case bridgehostservice.HostServiceMethodConfigString:
		return handleConfigString(ctx, configHostService, request.Key)
	case bridgehostservice.HostServiceMethodConfigBool:
		return handleConfigBool(ctx, configHostService, request.Key)
	case bridgehostservice.HostServiceMethodConfigInt:
		return handleConfigInt(ctx, configHostService, request.Key)
	case bridgehostservice.HostServiceMethodConfigDuration:
		return handleConfigDuration(ctx, configHostService, request.Key)
	default:
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			"unsupported config host service method: "+method,
		)
	}
}

// handleConfigGet reads one raw configuration value and returns its JSON representation.
func handleConfigGet(ctx context.Context, reader contract.ConfigService, key string) *bridgehostcall.HostCallResponseEnvelope {
	found, err := reader.Exists(ctx, key)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	if !found {
		return configValueResponse("", false)
	}

	value, err := reader.Get(ctx, key)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	encoded, err := gjson.Encode(value.Val())
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	return configValueResponse(string(encoded), true)
}

// handleConfigExists reports whether one configuration key exists.
func handleConfigExists(ctx context.Context, reader contract.ConfigService, key string) *bridgehostcall.HostCallResponseEnvelope {
	found, err := reader.Exists(ctx, key)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	return configValueResponse("", found)
}

// handleConfigString reads one configuration value as a string.
func handleConfigString(ctx context.Context, reader contract.ConfigService, key string) *bridgehostcall.HostCallResponseEnvelope {
	found, err := reader.Exists(ctx, key)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	if !found {
		return configValueResponse("", false)
	}
	value, err := reader.String(ctx, key, "")
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	return configValueResponse(value, true)
}

// handleConfigBool reads one configuration value as a bool string.
func handleConfigBool(ctx context.Context, reader contract.ConfigService, key string) *bridgehostcall.HostCallResponseEnvelope {
	found, err := reader.Exists(ctx, key)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	if !found {
		return configValueResponse("", false)
	}
	value, err := reader.Bool(ctx, key, false)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	return configValueResponse(strconv.FormatBool(value), true)
}

// handleConfigInt reads one configuration value as an int string.
func handleConfigInt(ctx context.Context, reader contract.ConfigService, key string) *bridgehostcall.HostCallResponseEnvelope {
	found, err := reader.Exists(ctx, key)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	if !found {
		return configValueResponse("", false)
	}
	value, err := reader.Int(ctx, key, 0)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	return configValueResponse(strconv.Itoa(value), true)
}

// handleConfigDuration reads one configuration value as a duration string.
func handleConfigDuration(ctx context.Context, reader contract.ConfigService, key string) *bridgehostcall.HostCallResponseEnvelope {
	found, err := reader.Exists(ctx, key)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	if !found {
		return configValueResponse("", false)
	}
	value, err := reader.Duration(ctx, key, 0)
	if err != nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, err.Error())
	}
	return configValueResponse(value.String(), true)
}

// configValueResponse wraps one config value response in a host-call success envelope.
func configValueResponse(value string, found bool) *bridgehostcall.HostCallResponseEnvelope {
	payload := bridgehostservice.MarshalHostServiceConfigValueResponse(&bridgehostservice.HostServiceConfigValueResponse{
		Value: value,
		Found: found,
	})
	return bridgehostcall.NewHostCallSuccessResponse(payload)
}
