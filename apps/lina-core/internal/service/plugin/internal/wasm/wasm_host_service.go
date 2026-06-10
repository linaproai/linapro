// This file implements the structured host service dispatcher used by the Wasm
// runtime host_call entrypoint. It also owns the shared capability service
// directory and transport helpers used by capability-backed domain dispatchers.

package wasm

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	"lina-core/pkg/plugin/capability/plugincap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// hostServiceRuntime carries the startup-owned services used by dynamic-plugin
// host calls. The pointer stored in hostServiceRuntimeSnapshot is immutable
// after publication so concurrent host calls always read a consistent set.
type hostServiceRuntime struct {
	domainServices      capability.Services
	pluginConfigFactory plugincap.ConfigServiceFactory
	hostConfigService   hostconfigcap.Service
	manifestFactory     manifestcap.ServiceFactory
}

var hostServiceRuntimeSnapshot atomic.Pointer[hostServiceRuntime]

// ConfigureHostServiceRuntime replaces the complete dynamic-plugin host-service
// runtime snapshot with startup-owned shared service instances.
func ConfigureHostServiceRuntime(
	domainServices capability.Services,
	pluginConfigFactory plugincap.ConfigServiceFactory,
	hostConfigService hostconfigcap.Service,
	manifestFactory manifestcap.ServiceFactory,
) error {
	if domainServices == nil {
		return gerror.New("domain host services directory is nil")
	}
	if pluginConfigFactory == nil {
		return gerror.New("wasm plugin config service requires a non-nil config factory")
	}
	if hostConfigService == nil {
		return gerror.New("wasm host config service requires a non-nil adapter")
	}
	if manifestFactory == nil {
		return gerror.New("wasm manifest host service requires a non-nil manifest factory")
	}
	setHostServiceRuntimeSnapshot(&hostServiceRuntime{
		domainServices:      domainServices,
		pluginConfigFactory: pluginConfigFactory,
		hostConfigService:   hostConfigService,
		manifestFactory:     manifestFactory,
	})
	return nil
}

// ConfigureDomainHostServices replaces the shared domain capability directory
// used by dynamic-plugin host calls.
func ConfigureDomainHostServices(services capability.Services) error {
	if services == nil {
		return gerror.New("domain host services directory is nil")
	}
	updateHostServiceRuntimeSnapshot(func(next *hostServiceRuntime) {
		next.domainServices = services
	})
	return nil
}

// capabilityServicesForHostCall returns the plugin-bound shared capability view.
func capabilityServicesForHostCall(hcc *hostCallContext) capability.Services {
	runtime := currentHostServiceRuntime()
	if hcc == nil || runtime == nil || runtime.domainServices == nil {
		return nil
	}
	return capability.ServicesForPlugin(runtime.domainServices, hcc.pluginID)
}

func currentHostServiceRuntime() *hostServiceRuntime {
	return hostServiceRuntimeSnapshot.Load()
}

func setHostServiceRuntimeSnapshot(runtime *hostServiceRuntime) {
	hostServiceRuntimeSnapshot.Store(runtime)
}

func updateHostServiceRuntimeSnapshot(apply func(*hostServiceRuntime)) {
	for {
		current := hostServiceRuntimeSnapshot.Load()
		next := &hostServiceRuntime{}
		if current != nil {
			*next = *current
		}
		apply(next)
		if hostServiceRuntimeSnapshot.CompareAndSwap(current, next) {
			return
		}
	}
}

// decodeCapabilityJSONRequest decodes a generic capability JSON payload into
// one dispatcher-local request structure.
func decodeCapabilityJSONRequest(payload []byte, out any) error {
	if out == nil || len(payload) == 0 {
		return nil
	}
	request, err := bridgehostservice.UnmarshalHostServiceCapabilityJSONRequest(payload)
	if err != nil {
		return err
	}
	if len(request.Value) == 0 {
		return nil
	}
	if err = json.Unmarshal(request.Value, out); err != nil {
		return gerror.Wrap(err, "decode capability JSON request failed")
	}
	return nil
}

// invalidCapabilityRequest returns a transport error for malformed domain
// host-service requests.
func invalidCapabilityRequest(err error) *bridgehostcall.HostCallResponseEnvelope {
	return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
}

// domainCapabilityResult converts a domain capability result into a host-call
// response envelope without leaking capability DTO ownership into pluginbridge.
func domainCapabilityResult(value any, err error) *bridgehostcall.HostCallResponseEnvelope {
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	return capabilityJSONResponse(value)
}

// domainServiceNotScoped returns the common error used when a domain service
// cannot be resolved for the current plugin context.
func domainServiceNotScoped(service string) *bridgehostcall.HostCallResponseEnvelope {
	return bridgehostcall.NewHostCallErrorResponse(
		bridgehostcall.HostCallStatusInternalError,
		service+" host service is not scoped",
	)
}

// domainMethodNotFound returns the common not-found response for unsupported
// domain host-service methods.
func domainMethodNotFound(service string, method string) *bridgehostcall.HostCallResponseEnvelope {
	return bridgehostcall.NewHostCallErrorResponse(
		bridgehostcall.HostCallStatusNotFound,
		service+" host service method not implemented: "+method,
	)
}

// hostCallErrorFromError preserves structured bizerr metadata in host-call
// error payloads and falls back to status-scoped metadata for technical errors.
func hostCallErrorFromError(status uint32, err error) *bridgehostcall.HostCallResponseEnvelope {
	return bridgehostcall.NewHostCallErrorResponseFromError(status, err)
}

// idsRequest carries a generic string identifier list.
type idsRequest struct {
	IDs []string `json:"ids"`
}

// handleHostServiceInvoke validates capability and authorization state before
// dispatching one structured host service invocation.
func handleHostServiceInvoke(
	ctx context.Context,
	hcc *hostCallContext,
	reqBytes []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	ctx = contextWithHostCallBizContext(ctx, hcc)

	request, err := bridgehostservice.UnmarshalHostServiceRequestEnvelope(reqBytes)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}

	requiredCapability := bridgehostservice.RequiredCapabilityForHostServiceMethod(request.Service, request.Method)
	if requiredCapability == "" {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			fmt.Sprintf("unsupported host service method: %s.%s", request.Service, request.Method),
		)
	}
	if !hcc.hasCapability(requiredCapability) {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusCapabilityDenied,
			fmt.Sprintf("plugin %s lacks capability %s", hcc.pluginID, requiredCapability),
		)
	}
	if !hcc.hasHostServiceAccess(request.Service, request.Method, request.ResourceRef, request.Table) {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusCapabilityDenied,
			fmt.Sprintf(
				"plugin %s is not authorized for host service %s.%s resource=%s table=%s",
				hcc.pluginID,
				request.Service,
				request.Method,
				request.ResourceRef,
				request.Table,
			),
		)
	}

	switch request.Service {
	case bridgehostservice.HostServiceRuntime:
		return dispatchRuntimeHostService(ctx, hcc, request.Method, request.Payload)
	case bridgehostservice.HostServiceStorage:
		return dispatchStorageHostService(ctx, hcc, request.ResourceRef, request.Method, request.Payload)
	case bridgehostservice.HostServiceNetwork:
		return dispatchNetworkHostService(ctx, hcc, request.ResourceRef, request.Method, request.Payload)
	case bridgehostservice.HostServiceData:
		return dispatchDataHostService(ctx, hcc, request.Table, request.Method, request.Payload)
	case bridgehostservice.HostServiceCache:
		return dispatchCacheHostService(ctx, hcc, request.ResourceRef, request.Method, request.Payload)
	case bridgehostservice.HostServiceLock:
		return dispatchLockHostService(ctx, hcc, request.ResourceRef, request.Method, request.Payload)
	case bridgehostservice.HostServiceHostConfig:
		return dispatchHostConfigService(ctx, hcc, request.Method, request.Payload)
	case bridgehostservice.HostServiceManifest:
		return dispatchManifestHostService(ctx, hcc, request.Method, request.Payload)
	case bridgehostservice.HostServiceAPIDoc:
		return dispatchAPIDocHostService(ctx, hcc, request.Method, request.Payload)
	case bridgehostservice.HostServiceAuth:
		return dispatchAuthHostService(ctx, hcc, request.Method, request.Payload)
	case bridgehostservice.HostServiceAuthz:
		return dispatchAuthzHostService(ctx, hcc, request.Method, request.Payload)
	case bridgehostservice.HostServiceAI:
		return dispatchAIHostService(ctx, hcc, request.ResourceRef, request.Method, request.Payload)
	case bridgehostservice.HostServiceUsers:
		return dispatchUsersHostService(ctx, hcc, request.Method, request.Payload)
	case bridgehostservice.HostServiceBizCtx:
		return dispatchBizCtxHostService(ctx, hcc, request.Method, request.Payload)
	case bridgehostservice.HostServiceDict:
		return dispatchDictHostService(ctx, hcc, request.Method, request.Payload)
	case bridgehostservice.HostServiceFiles:
		return dispatchFilesHostService(ctx, hcc, request.Method, request.Payload)
	case bridgehostservice.HostServiceI18n:
		return dispatchI18nHostService(ctx, hcc, request.Method, request.Payload)
	case bridgehostservice.HostServiceInfra:
		return dispatchInfraHostService(ctx, hcc, request.Method, request.Payload)
	case bridgehostservice.HostServiceJobs:
		return dispatchJobsHostService(ctx, hcc, request.Method, request.Payload)
	case bridgehostservice.HostServiceNotifications:
		return dispatchNotificationsHostService(ctx, hcc, request.ResourceRef, request.Method, request.Payload)
	case bridgehostservice.HostServicePlugins:
		return dispatchPluginsHostService(ctx, hcc, request.Method, request.Payload)
	case bridgehostservice.HostServiceRoute:
		return dispatchRouteHostService(ctx, hcc, request.Method, request.Payload)
	case bridgehostservice.HostServiceSessions:
		return dispatchSessionsHostService(ctx, hcc, request.Method, request.Payload)
	case bridgehostservice.HostServiceOrg:
		return dispatchOrgHostService(ctx, hcc, request.Method, request.Payload)
	case bridgehostservice.HostServiceTenant:
		return dispatchTenantHostService(ctx, hcc, request.Method, request.Payload)
	default:
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusNotFound,
			fmt.Sprintf("host service not implemented yet: %s", request.Service),
		)
	}
}

// contextWithHostCallBizContext exposes the dynamic-plugin identity snapshot
// through the same bizctxcap context path used by source-plugin capabilities.
func contextWithHostCallBizContext(ctx context.Context, hcc *hostCallContext) context.Context {
	if hcc == nil || hcc.identity == nil {
		return ctx
	}
	identity := hcc.identity
	return bizctxcap.WithCurrentContext(ctx, bizctxcap.CurrentContext{
		UserID:          int(identity.UserID),
		Username:        identity.Username,
		TenantID:        int(identity.TenantId),
		ActingUserID:    int(identity.ActingUserId),
		ActingAsTenant:  identity.ActingAsTenant,
		IsImpersonation: identity.IsImpersonation,
	})
}
