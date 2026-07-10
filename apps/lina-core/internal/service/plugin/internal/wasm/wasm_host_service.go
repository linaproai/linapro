// This file implements the structured host service dispatcher used by the Wasm
// runtime host_call entrypoint. It also owns the shared capability service
// directory and transport helpers used by capability-backed domain dispatchers.

package wasm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/plugin/internal/capabilityowner"
	"lina-core/internal/service/plugin/internal/manifestresource"
	"lina-core/internal/service/plugin/internal/pluginconfig"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capregistry"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// hostServiceRuntime carries the startup-owned services used by dynamic-plugin
// host calls for one WASM runtime instance.
type hostServiceRuntime struct {
	domainServices      capability.Services
	ownerCapabilities   *capregistry.Registry
	pluginConfigFactory pluginconfig.Factory
	hostConfigService   hostconfigcap.Service
	manifestFactory     manifestresource.Factory
	storageUploads      *storageUploadSessions
}

// NewRuntime creates a dynamic-plugin WASM runtime from startup-owned shared
// service instances.
func NewRuntime(
	domainServices capability.Services,
	ownerCapabilities *capregistry.Registry,
	pluginConfigFactory pluginconfig.Factory,
	hostConfigService hostconfigcap.Service,
	manifestFactory manifestresource.Factory,
) (Runtime, error) {
	if domainServices == nil {
		return nil, gerror.New("domain host services directory is nil")
	}
	if ownerCapabilities == nil {
		return nil, gerror.New("owner capability registry is nil")
	}
	if pluginConfigFactory == nil {
		return nil, gerror.New("wasm plugin config service requires a non-nil config factory")
	}
	if hostConfigService == nil {
		return nil, gerror.New("wasm host config service requires a non-nil adapter")
	}
	if manifestFactory == nil {
		return nil, gerror.New("wasm manifest host service requires a non-nil manifest factory")
	}
	return &runtimeImpl{
		hostServices: &hostServiceRuntime{
			domainServices:      domainServices,
			ownerCapabilities:   ownerCapabilities,
			pluginConfigFactory: pluginConfigFactory,
			hostConfigService:   hostConfigService,
			manifestFactory:     manifestFactory,
			storageUploads:      newStorageUploadSessions(),
		},
		cache:    make(map[string]*wasmCacheEntry),
		inflight: make(map[string]*wasmCompileInflight),
	}, nil
}

// capabilityServicesForHostCall returns the plugin-bound shared capability view.
func capabilityServicesForHostCall(hcc *hostCallContext) capability.Services {
	if hcc == nil || hcc.runtime == nil || hcc.runtime.domainServices == nil {
		return nil
	}
	return capabilityowner.ServicesForPlugin(hcc.runtime.domainServices, hcc.pluginID)
}

// decodeCapabilityJSONRequest decodes a generic capability JSON payload into
// one dispatcher-local request structure.
func decodeCapabilityJSONRequest(payload []byte, out any) error {
	if out == nil || len(payload) == 0 {
		return nil
	}
	request, err := bridgehostservice.UnmarshalHostServiceJSONRequest(payload)
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

// capabilityJSONResponse encodes one ordinary domain value as a JSON host-service
// success envelope. New JSON host-service methods should use this helper with
// decodeCapabilityJSONRequest instead of dedicated binary codecs.
func capabilityJSONResponse(value any) *bridgehostcall.HostCallResponseEnvelope {
	content, err := json.Marshal(value)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	payload := bridgehostservice.MarshalHostServiceJSONResponse(
		&bridgehostservice.HostServiceJSONResponse{Value: content},
	)
	return bridgehostcall.NewHostCallSuccessResponse(payload)
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
	if hcc == nil || hcc.runtime == nil {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusInternalError,
			"wasm host service runtime is not configured",
		)
	}
	ctx = contextWithHostCallBizContext(ctx, hcc)

	request, err := bridgehostservice.UnmarshalHostServiceRequestEnvelope(reqBytes)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	if isOwnerHostServiceRequest(request) {
		return dispatchOwnerHostService(ctx, hcc, request)
	}
	if strings.TrimSpace(request.Version) != "" {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusInvalidRequest,
			"host service version requires owner",
		)
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

	return dispatchRegisteredHostService(ctx, hcc, request)
}

func isOwnerHostServiceRequest(request *bridgehostservice.HostServiceRequestEnvelope) bool {
	if request == nil {
		return false
	}
	return strings.TrimSpace(request.Owner) != ""
}

// contextWithHostCallBizContext exposes the dynamic-plugin identity snapshot
// through the same bizctxcap context path used by source-plugin capabilities.
func contextWithHostCallBizContext(ctx context.Context, hcc *hostCallContext) context.Context {
	if hcc == nil || hcc.identity == nil {
		return ctx
	}
	identity := hcc.identity
	return bizctxcap.WithCurrentContext(ctx, bizctxcap.CurrentContext{
		TokenID:              identity.TokenID,
		UserID:               int(identity.UserID),
		Username:             identity.Username,
		TenantID:             int(identity.TenantId),
		ActingUserID:         int(identity.ActingUserId),
		ActingAsTenant:       identity.ActingAsTenant,
		IsImpersonation:      identity.IsImpersonation,
		Permissions:          identity.Permissions,
		DataScope:            int(identity.DataScope),
		DataScopeUnsupported: identity.DataScopeUnsupported,
		UnsupportedDataScope: int(identity.UnsupportedDataScope),
		IsSuperAdmin:         identity.IsSuperAdmin,
	})
}
