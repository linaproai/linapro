// This file adapts authorization host-service calls to the shared authz
// capability service.

package wasm

import (
	"context"
	"strings"

	"lina-core/pkg/plugin/capability/authcap/authz"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchAuthzHostService routes authorization-domain host-service calls.
func dispatchAuthzHostService(
	ctx context.Context,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	service := authzServiceForHostCall(hcc)
	if service == nil {
		return domainServiceNotScoped("authz")
	}
	capCtx := capabilityContextForHostCall(hcc, bridgehostservice.HostServiceAuthz, method)
	switch method {
	case bridgehostservice.HostServiceMethodAuthzBatchGetPermissions:
		var request idsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.BatchGetPermissions(ctx, capCtx, permissionKeys(request.IDs))
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodAuthzBatchHasPermissions:
		var request idsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.BatchHasPermissions(ctx, capCtx, permissionKeys(request.IDs))
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodAuthzHasPermission:
		var request keyRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.HasPermission(ctx, capCtx, authz.PermissionKey(request.Key))
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodAuthzIsPlatformAdmin:
		var request userIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.IsPlatformAdmin(ctx, capCtx, authz.UserID(request.UserID))
		return domainCapabilityResult(result, err)
	default:
		return domainMethodNotFound("authz", method)
	}
}

// authzServiceForHostCall resolves the authorization service for one host call.
func authzServiceForHostCall(hcc *hostCallContext) authz.Service {
	services := capabilityServicesForHostCall(hcc)
	if services == nil || services.Auth() == nil {
		return nil
	}
	return services.Auth().Authz()
}

// keyRequest carries one string key.
type keyRequest struct {
	Key string `json:"key"`
}

// userIDRequest carries one user identifier.
type userIDRequest struct {
	UserID string `json:"userId"`
}

// permissionKeys converts transport string identifiers into typed permission keys.
func permissionKeys(ids []string) []authz.PermissionKey {
	out := make([]authz.PermissionKey, 0, len(ids))
	for _, id := range ids {
		if value := strings.TrimSpace(id); value != "" {
			out = append(out, authz.PermissionKey(value))
		}
	}
	return out
}
