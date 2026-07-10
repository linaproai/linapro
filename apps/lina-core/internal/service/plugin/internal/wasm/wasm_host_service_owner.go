// This file implements the plugin-owned host-service dispatcher. It validates
// owner descriptor registration and the active authorization snapshot before
// forwarding a request to the owner-provided generic invoker.

package wasm

import (
	"context"
	"strings"

	"lina-core/internal/service/plugin/internal/capabilityowner"
	"lina-core/pkg/plugin/capability/capregistry"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

func dispatchOwnerHostService(
	ctx context.Context,
	hcc *hostCallContext,
	request *protocol.HostServiceRequestEnvelope,
) *protocol.HostCallResponseEnvelope {
	if request == nil {
		return protocol.NewHostCallErrorResponse(
			protocol.HostCallStatusInvalidRequest,
			"owner host service request is nil",
		)
	}

	var (
		owner       = strings.ToLower(strings.TrimSpace(request.Owner))
		service     = strings.ToLower(strings.TrimSpace(request.Service))
		version     = strings.ToLower(strings.TrimSpace(request.Version))
		method      = strings.ToLower(strings.TrimSpace(request.Method))
		resourceRef = strings.TrimSpace(request.ResourceRef)
		table       = strings.TrimSpace(request.Table)
	)
	if owner == "" || service == "" || version == "" || method == "" {
		return protocol.NewHostCallErrorResponse(
			protocol.HostCallStatusInvalidRequest,
			"owner host service request requires owner, service, version, and method",
		)
	}
	if hcc == nil || hcc.runtime == nil || hcc.runtime.ownerCapabilities == nil {
		return protocol.NewHostCallErrorResponse(
			protocol.HostCallStatusInternalError,
			"owner capability registry is not configured",
		)
	}

	registered, ok := hcc.runtime.ownerCapabilities.LookupMethod(owner, service, version, method)
	if !ok {
		return protocol.NewHostCallErrorResponse(
			protocol.HostCallStatusNotFound,
			"owner host service method not registered: "+protocol.HostServiceIdentityLabel(owner, service, version)+"."+method,
		)
	}
	if !hcc.hasOwnerHostServiceAccess(owner, service, version, method, resourceRef, table) {
		return protocol.NewHostCallErrorResponse(
			protocol.HostCallStatusCapabilityDenied,
			"plugin "+hcc.pluginID+" is not authorized for owner host service "+
				protocol.HostServiceIdentityLabel(owner, service, version)+"."+method,
		)
	}
	if registered.Invoker == nil {
		return protocol.NewHostCallErrorResponse(
			protocol.HostCallStatusNotFound,
			"owner host service handler not registered: "+protocol.HostServiceIdentityLabel(owner, service, version)+"."+method,
		)
	}

	result, err := registered.Invoker.Invoke(ctx, capregistry.Invocation{
		CallerPluginID: hcc.pluginID,
		OwnerPluginID:  owner,
		Services:       capabilityowner.ServicesForPlugin(hcc.runtime.domainServices, owner),
		Service:        service,
		Version:        version,
		Method:         method,
		ResourceRef:    resourceRef,
		Table:          table,
		Payload:        append([]byte(nil), request.Payload...),
	})
	if err != nil {
		return hostCallErrorFromError(protocol.HostCallStatusInvalidRequest, err)
	}
	if result == nil {
		return protocol.NewHostCallErrorResponse(
			protocol.HostCallStatusInternalError,
			"owner host service handler returned nil response",
		)
	}
	return &protocol.HostCallResponseEnvelope{
		Status:  result.Status,
		Payload: append([]byte(nil), result.Payload...),
	}
}
