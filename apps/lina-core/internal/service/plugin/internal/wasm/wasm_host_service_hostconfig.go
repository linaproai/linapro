// This file implements the hostConfig host service for dynamic plugins.

package wasm

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"

	"lina-core/pkg/plugin/capability/hostconfigcap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchHostConfigService routes authorized hostConfig calls to the host
// config reader and sys_config owner service.
func dispatchHostConfigService(
	ctx context.Context,
	hcc *hostCallContext,
	resourceRef string,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	switch method {
	case bridgehostservice.HostServiceMethodHostConfigGet:
		request, err := bridgehostservice.UnmarshalHostServiceConfigKeyRequest(payload)
		if err != nil {
			return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
		}
		if hcc == nil || hcc.runtime == nil || hcc.runtime.hostConfigService == nil {
			return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, "host config service is not configured")
		}
		key := strings.TrimSpace(request.Key)
		if response := validateHostConfigRequestKey(resourceRef, key); response != nil {
			return response
		}
		return handleHostConfigGet(ctx, hcc.runtime.hostConfigService, key)
	case bridgehostservice.HostServiceMethodHostConfigSysConfigGet:
		service := sysConfigServiceForHostCall(hcc)
		if service == nil {
			return domainServiceNotScoped("hostconfig.sys_config")
		}
		var request hostConfigSysConfigKeyRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		key := strings.TrimSpace(request.Key)
		if response := validateHostConfigRequestKey(resourceRef, key); response != nil {
			return response
		}
		result, err := service.Get(ctx, hostconfigcap.SysConfigKey(key))
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodHostConfigSysConfigSetValue:
		service := sysConfigServiceForHostCall(hcc)
		if service == nil {
			return domainServiceNotScoped("hostconfig.sys_config")
		}
		var request hostConfigSysConfigSetValueRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		key := strings.TrimSpace(request.Key)
		if response := validateHostConfigRequestKey(resourceRef, key); response != nil {
			return response
		}
		var options *hostconfigcap.SetSysConfigValueOptions
		if request.SystemManageable != nil {
			options = &hostconfigcap.SetSysConfigValueOptions{
				SystemManageable: request.SystemManageable,
			}
		}
		err := service.SetValue(ctx, hostconfigcap.SysConfigKey(key), request.Value, options)
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodHostConfigSysConfigReset:
		service := sysConfigServiceForHostCall(hcc)
		if service == nil {
			return domainServiceNotScoped("hostconfig.sys_config")
		}
		var request hostConfigSysConfigKeyRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		key := strings.TrimSpace(request.Key)
		if response := validateHostConfigRequestKey(resourceRef, key); response != nil {
			return response
		}
		err := service.Reset(ctx, hostconfigcap.SysConfigKey(key))
		return domainCapabilityResult(true, err)
	default:
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			"unsupported host config service method: "+method,
		)
	}
}

// validateHostConfigRequestKey ensures the authorized envelope key and payload
// key describe the same governed host configuration resource.
func validateHostConfigRequestKey(resourceRef string, key string) *bridgehostcall.HostCallResponseEnvelope {
	authorizedKey := strings.TrimSpace(resourceRef)
	if authorizedKey == "" || strings.TrimSpace(key) != authorizedKey {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusCapabilityDenied,
			"hostconfig request key must match the authorized resourceRef",
		)
	}
	return nil
}

// handleHostConfigGet reads one authorized host config value and returns JSON.
func handleHostConfigGet(ctx context.Context, reader hostconfigcap.Service, key string) *bridgehostcall.HostCallResponseEnvelope {
	found, err := reader.Exists(ctx, key)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	if !found {
		return configValueResponse("", false)
	}

	value, err := reader.Get(ctx, key, nil)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	encoded, err := gjson.Encode(value.Val())
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	return configValueResponse(string(encoded), true)
}

func sysConfigServiceForHostCall(hcc *hostCallContext) hostconfigcap.SysConfigService {
	services := capabilityServicesForHostCall(hcc)
	if services == nil || services.HostConfig() == nil {
		return nil
	}
	return services.HostConfig().SysConfig()
}

// hostConfigSysConfigKeyRequest carries one sys_config key in JSON host calls.
type hostConfigSysConfigKeyRequest struct {
	Key string `json:"key"`
}

// hostConfigSysConfigSetValueRequest carries one sys_config value mutation.
type hostConfigSysConfigSetValueRequest struct {
	Key              string `json:"key"`
	Value            string `json:"value"`
	SystemManageable *bool  `json:"systemManageable,omitempty"`
}
