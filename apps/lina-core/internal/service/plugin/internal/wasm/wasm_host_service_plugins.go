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
	switch method {
	case bridgehostservice.HostServiceMethodPluginsCurrent:
		registry := service.Registry()
		if registry == nil {
			return domainServiceNotScoped("plugins.registry")
		}
		result, err := registry.Current(ctx)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodPluginsBatchGet:
		registry := service.Registry()
		if registry == nil {
			return domainServiceNotScoped("plugins.registry")
		}
		var request idsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := registry.BatchGet(ctx, pluginIDs(request.IDs))
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodPluginsList:
		registry := service.Registry()
		if registry == nil {
			return domainServiceNotScoped("plugins.registry")
		}
		var request plugincap.ListInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := registry.List(ctx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodPluginsListTenant:
		registry := service.Registry()
		if registry == nil {
			return domainServiceNotScoped("plugins.registry")
		}
		var request plugincap.TenantListInput
		if len(payload) > 0 {
			if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
				return invalidCapabilityRequest(err)
			}
		}
		result, err := registry.ListTenantPlugins(ctx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodPluginsStateIsEnabled:
		stateSvc := service.State()
		if stateSvc == nil {
			return domainServiceNotScoped("plugins.state")
		}
		var request pluginIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := stateSvc.IsEnabled(ctx, plugincap.PluginID(request.PluginID))
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodPluginsStateIsProviderEnabled:
		stateSvc := service.State()
		if stateSvc == nil {
			return domainServiceNotScoped("plugins.state")
		}
		var request pluginIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := stateSvc.IsProviderEnabled(ctx, plugincap.PluginID(request.PluginID))
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodPluginsStateIsEnabledAuthoritative:
		stateSvc := service.State()
		if stateSvc == nil {
			return domainServiceNotScoped("plugins.state")
		}
		var request pluginIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := stateSvc.IsEnabledAuthoritative(ctx, plugincap.PluginID(request.PluginID))
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodPluginsLifecycleEnsureTenantPluginDisableAllowed:
		lifecycleSvc := service.Lifecycle()
		if lifecycleSvc == nil {
			return domainServiceNotScoped("plugins.lifecycle")
		}
		var request tenantPluginLifecycleRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := lifecycleSvc.EnsureTenantPluginDisableAllowed(ctx, request.PluginID, request.TenantID)
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodPluginsLifecycleNotifyTenantPluginDisabled:
		lifecycleSvc := service.Lifecycle()
		if lifecycleSvc == nil {
			return domainServiceNotScoped("plugins.lifecycle")
		}
		var request tenantPluginLifecycleRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		lifecycleSvc.NotifyTenantPluginDisabled(ctx, request.PluginID, request.TenantID)
		return capabilityJSONResponse(true)
	case bridgehostservice.HostServiceMethodPluginsLifecycleEnsureTenantDeleteAllowed:
		lifecycleSvc := service.Lifecycle()
		if lifecycleSvc == nil {
			return domainServiceNotScoped("plugins.lifecycle")
		}
		var request tenantIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := lifecycleSvc.EnsureTenantDeleteAllowed(ctx, request.TenantID)
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodPluginsLifecycleNotifyTenantDeleted:
		lifecycleSvc := service.Lifecycle()
		if lifecycleSvc == nil {
			return domainServiceNotScoped("plugins.lifecycle")
		}
		var request tenantIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		lifecycleSvc.NotifyTenantDeleted(ctx, request.TenantID)
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
	// PluginID is the plugin identifier.
	PluginID string `json:"pluginId"`
}

// tenantPluginLifecycleRequest carries one plugin and tenant pair.
type tenantPluginLifecycleRequest struct {
	// PluginID is the plugin identifier.
	PluginID string `json:"pluginId"`
	// TenantID is the tenant identifier.
	TenantID int `json:"tenantId"`
}

// tenantIDRequest carries one tenant identifier.
type tenantIDRequest struct {
	// TenantID is the tenant identifier.
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
