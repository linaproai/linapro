// This file implements the read-only config host service for dynamic plugins.

package wasm

import (
	"context"
	"strconv"

	"github.com/gogf/gf/v2/encoding/gjson"

	"lina-core/pkg/pluginbridge"
	configsvc "lina-core/pkg/pluginservice/config"
)

// dispatchConfigHostService routes config host service methods to the generic
// read-only plugin configuration service.
func dispatchConfigHostService(
	ctx context.Context,
	_ *hostCallContext,
	method string,
	payload []byte,
) *pluginbridge.HostCallResponseEnvelope {
	request, err := pluginbridge.UnmarshalHostServiceConfigKeyRequest(payload)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInvalidRequest, err.Error())
	}

	reader := configsvc.New()
	switch method {
	case pluginbridge.HostServiceMethodConfigGet:
		return handleConfigGet(ctx, reader, request.Key)
	case pluginbridge.HostServiceMethodConfigExists:
		return handleConfigExists(ctx, reader, request.Key)
	case pluginbridge.HostServiceMethodConfigString:
		return handleConfigString(ctx, reader, request.Key)
	case pluginbridge.HostServiceMethodConfigBool:
		return handleConfigBool(ctx, reader, request.Key)
	case pluginbridge.HostServiceMethodConfigInt:
		return handleConfigInt(ctx, reader, request.Key)
	case pluginbridge.HostServiceMethodConfigDuration:
		return handleConfigDuration(ctx, reader, request.Key)
	default:
		return pluginbridge.NewHostCallErrorResponse(
			pluginbridge.HostCallStatusNotFound,
			"unsupported config host service method: "+method,
		)
	}
}

// handleConfigGet reads one raw configuration value and returns its JSON representation.
func handleConfigGet(ctx context.Context, reader configsvc.Service, key string) *pluginbridge.HostCallResponseEnvelope {
	found, err := reader.Exists(ctx, key)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}
	if !found {
		return configValueResponse("", false)
	}

	value, err := reader.Get(ctx, key)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}
	encoded, err := gjson.Encode(value.Val())
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}
	return configValueResponse(string(encoded), true)
}

// handleConfigExists reports whether one configuration key exists.
func handleConfigExists(ctx context.Context, reader configsvc.Service, key string) *pluginbridge.HostCallResponseEnvelope {
	found, err := reader.Exists(ctx, key)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}
	return configValueResponse("", found)
}

// handleConfigString reads one configuration value as a string.
func handleConfigString(ctx context.Context, reader configsvc.Service, key string) *pluginbridge.HostCallResponseEnvelope {
	found, err := reader.Exists(ctx, key)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}
	if !found {
		return configValueResponse("", false)
	}
	value, err := reader.String(ctx, key, "")
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}
	return configValueResponse(value, true)
}

// handleConfigBool reads one configuration value as a bool string.
func handleConfigBool(ctx context.Context, reader configsvc.Service, key string) *pluginbridge.HostCallResponseEnvelope {
	found, err := reader.Exists(ctx, key)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}
	if !found {
		return configValueResponse("", false)
	}
	value, err := reader.Bool(ctx, key, false)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}
	return configValueResponse(strconv.FormatBool(value), true)
}

// handleConfigInt reads one configuration value as an int string.
func handleConfigInt(ctx context.Context, reader configsvc.Service, key string) *pluginbridge.HostCallResponseEnvelope {
	found, err := reader.Exists(ctx, key)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}
	if !found {
		return configValueResponse("", false)
	}
	value, err := reader.Int(ctx, key, 0)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}
	return configValueResponse(strconv.Itoa(value), true)
}

// handleConfigDuration reads one configuration value as a duration string.
func handleConfigDuration(ctx context.Context, reader configsvc.Service, key string) *pluginbridge.HostCallResponseEnvelope {
	found, err := reader.Exists(ctx, key)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}
	if !found {
		return configValueResponse("", false)
	}
	value, err := reader.Duration(ctx, key, 0)
	if err != nil {
		return pluginbridge.NewHostCallErrorResponse(pluginbridge.HostCallStatusInternalError, err.Error())
	}
	return configValueResponse(value.String(), true)
}

// configValueResponse wraps one config value response in a host-call success envelope.
func configValueResponse(value string, found bool) *pluginbridge.HostCallResponseEnvelope {
	payload := pluginbridge.MarshalHostServiceConfigValueResponse(&pluginbridge.HostServiceConfigValueResponse{
		Value: value,
		Found: found,
	})
	return pluginbridge.NewHostCallSuccessResponse(payload)
}
