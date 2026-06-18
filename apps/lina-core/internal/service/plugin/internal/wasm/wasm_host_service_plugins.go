// This file adapts plugin-governance host-service calls to the shared plugin
// capability service, including plugin-scoped config reads exposed as
// plugins.config.get.

package wasm

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"

	"lina-core/pkg/plugin/capability/plugincap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchPluginsHostService routes plugin-governance domain host-service calls.
func dispatchPluginsHostService(
	ctx context.Context,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	if method == bridgehostservice.HostServiceMethodPluginsConfigGet {
		return dispatchPluginsConfigGet(ctx, hcc, payload)
	}
	service := pluginsServiceForHostCall(hcc)
	if service == nil {
		return domainServiceNotScoped("plugins")
	}
	capCtx := capabilityContextForHostCall(hcc, bridgehostservice.HostServicePlugins, method)
	switch method {
	case bridgehostservice.HostServiceMethodPluginsCurrent:
		result, err := service.Current(ctx, capCtx)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodPluginsBatchGet:
		var request idsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.BatchGet(ctx, capCtx, pluginIDs(request.IDs))
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodPluginsSearch:
		var request plugincap.SearchInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.Search(ctx, capCtx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodPluginsListTenant:
		var request plugincap.TenantListInput
		if len(payload) > 0 {
			if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
				return invalidCapabilityRequest(err)
			}
		}
		result, err := service.ListTenantPlugins(ctx, capCtx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodPluginsBatchGetCapabilityStatus:
		var request capabilityKeysRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.BatchGetCapabilityStatus(ctx, capCtx, capabilityKeys(request.Keys))
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodPluginsIsEnabled:
		state := service.State()
		if state == nil {
			return domainServiceNotScoped("plugins.state")
		}
		var request pluginIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		return capabilityJSONResponse(state.IsEnabled(ctx, request.PluginID))
	case bridgehostservice.HostServiceMethodPluginsIsProviderEnabled:
		state := service.State()
		if state == nil {
			return domainServiceNotScoped("plugins.state")
		}
		var request pluginIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		return capabilityJSONResponse(state.IsProviderEnabled(ctx, request.PluginID))
	case bridgehostservice.HostServiceMethodPluginsIsEnabledAuthoritative:
		state := service.State()
		if state == nil {
			return domainServiceNotScoped("plugins.state")
		}
		var request pluginIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		return capabilityJSONResponse(state.IsEnabledAuthoritative(ctx, request.PluginID))
	case bridgehostservice.HostServiceMethodPluginsLifecycleEnsureTenantPluginDisable:
		lifecycle := service.Lifecycle()
		if lifecycle == nil {
			return domainServiceNotScoped("plugins.lifecycle")
		}
		var request pluginTenantLifecycleRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := lifecycle.EnsureTenantPluginDisableAllowed(ctx, request.PluginID, request.TenantID)
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodPluginsLifecycleNotifyTenantPluginDisabled:
		lifecycle := service.Lifecycle()
		if lifecycle == nil {
			return domainServiceNotScoped("plugins.lifecycle")
		}
		var request pluginTenantLifecycleRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		lifecycle.NotifyTenantPluginDisabled(ctx, request.PluginID, request.TenantID)
		return capabilityJSONResponse(true)
	case bridgehostservice.HostServiceMethodPluginsLifecycleEnsureTenantDelete:
		lifecycle := service.Lifecycle()
		if lifecycle == nil {
			return domainServiceNotScoped("plugins.lifecycle")
		}
		var request tenantLifecycleRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := lifecycle.EnsureTenantDeleteAllowed(ctx, request.TenantID)
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodPluginsLifecycleNotifyTenantDeleted:
		lifecycle := service.Lifecycle()
		if lifecycle == nil {
			return domainServiceNotScoped("plugins.lifecycle")
		}
		var request tenantLifecycleRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		lifecycle.NotifyTenantDeleted(ctx, request.TenantID)
		return capabilityJSONResponse(true)
	default:
		return domainMethodNotFound("plugins", method)
	}
}

// pluginsServiceForHostCall resolves the plugin governance service for one host call.
func pluginsServiceForHostCall(hcc *hostCallContext) plugincap.Service {
	services := capabilityServicesForHostCall(hcc)
	if services == nil {
		return nil
	}
	return services.Plugins()
}

// capabilityKeysRequest carries framework capability status keys.
type capabilityKeysRequest struct {
	// Keys are stable framework capability identifiers.
	Keys []string `json:"keys"`
}

// capabilityKeys converts transport keys to plugin capability keys.
func capabilityKeys(keys []string) []plugincap.CapabilityKey {
	out := make([]plugincap.CapabilityKey, 0, len(keys))
	for _, key := range keys {
		out = append(out, plugincap.CapabilityKey(key))
	}
	return out
}

// dispatchPluginsConfigGet routes plugins.config.get calls to the generic
// read-only plugin configuration service.
func dispatchPluginsConfigGet(
	ctx context.Context,
	hcc *hostCallContext,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceConfigKeyRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	if hcc == nil || hcc.runtime == nil || hcc.runtime.pluginConfigFactory == nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, "plugin config service is not configured")
	}
	factory := hcc.runtime.pluginConfigFactory
	if len(hcc.artifactDefaultConfig) > 0 {
		factory = factory.WithArtifactConfig(hcc.pluginID, hcc.artifactDefaultConfig)
	}
	return handlePluginConfigGet(ctx, factory.ForPlugin(hcc.pluginID), request.Key)
}

// handlePluginConfigGet reads one raw configuration value and returns its JSON
// representation.
func handlePluginConfigGet(
	ctx context.Context,
	reader plugincap.ConfigService,
	key string,
) *bridgehostcall.HostCallResponseEnvelope {
	if reader == nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, "plugin config service is not scoped")
	}
	found, err := reader.Exists(ctx, key)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	if !found {
		return configValueResponse("", false)
	}

	value, err := reader.Get(ctx, key)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	encoded, err := gjson.Encode(value.Val())
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	return configValueResponse(string(encoded), true)
}

// configValueResponse wraps one config value response in a host-call success
// envelope.
func configValueResponse(value string, found bool) *bridgehostcall.HostCallResponseEnvelope {
	payload := bridgehostservice.MarshalHostServiceConfigValueResponse(&bridgehostservice.HostServiceConfigValueResponse{
		Value: value,
		Found: found,
	})
	return bridgehostcall.NewHostCallSuccessResponse(payload)
}

// pluginIDRequest carries one plugin identifier.
type pluginIDRequest struct {
	PluginID string `json:"pluginId"`
}

// pluginTenantLifecycleRequest carries one tenant-scoped plugin lifecycle target.
type pluginTenantLifecycleRequest struct {
	PluginID string `json:"pluginId"`
	TenantID int    `json:"tenantId"`
}

// tenantLifecycleRequest carries one tenant lifecycle target.
type tenantLifecycleRequest struct {
	TenantID int `json:"tenantId"`
}

// pluginIDs converts transport string identifiers into typed plugin IDs.
func pluginIDs(ids []string) []plugincap.PluginID {
	out := make([]plugincap.PluginID, 0, len(ids))
	for _, id := range ids {
		if value := strings.TrimSpace(id); value != "" {
			out = append(out, plugincap.PluginID(value))
		}
	}
	return out
}
