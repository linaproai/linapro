// This file adapts dynamic-plugin auth-domain host-service calls, including
// token handoff and authorization method families, to capability.Services.

package wasm

import (
	"context"
	"strings"

	"lina-core/pkg/plugin/capability/authcap/authz"
	"lina-core/pkg/plugin/capability/authcap/token"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchAuthHostService routes auth token host-service calls.
func dispatchAuthHostService(
	ctx context.Context,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	switch method {
	case bridgehostservice.HostServiceMethodAuthzBatchGetPermissions,
		bridgehostservice.HostServiceMethodAuthzBatchHasPermissions,
		bridgehostservice.HostServiceMethodAuthzHasPermission,
		bridgehostservice.HostServiceMethodAuthzIsPlatformAdmin:
		return dispatchAuthAuthorizationMethods(ctx, hcc, method, payload)
	}

	services := capabilityServicesForHostCall(hcc)
	if services == nil || services.Auth() == nil || services.Auth().Token() == nil {
		return domainServiceNotScoped("auth")
	}
	service := services.Auth().Token()
	switch method {
	case bridgehostservice.HostServiceMethodAuthSelectTenant:
		var request token.SelectTenantInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.SelectTenant(ctx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodAuthSwitchTenant:
		var request token.SwitchTenantInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.SwitchTenant(ctx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodAuthIssueImpersonationToken:
		var request token.ImpersonationTokenIssueInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.IssueImpersonationToken(ctx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodAuthRevokeImpersonationToken:
		var request token.ImpersonationTokenRevokeInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := service.RevokeImpersonationToken(ctx, request)
		return domainCapabilityResult(true, err)
	default:
		return domainMethodNotFound("auth", method)
	}
}

// dispatchAuthAuthorizationMethods routes auth-domain authorization method calls.
func dispatchAuthAuthorizationMethods(
	ctx context.Context,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	service := authzServiceForHostCall(hcc)
	if service == nil {
		return domainServiceNotScoped("auth.authz")
	}
	switch method {
	case bridgehostservice.HostServiceMethodAuthzBatchGetPermissions:
		var request idsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.BatchGetPermissions(ctx, permissionKeys(request.IDs))
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodAuthzBatchHasPermissions:
		var request idsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.BatchHasPermissions(ctx, permissionKeys(request.IDs))
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodAuthzHasPermission:
		var request keyRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.HasPermission(ctx, authz.PermissionKey(request.Key))
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodAuthzIsPlatformAdmin:
		var request userIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.IsPlatformAdmin(ctx, authz.UserID(request.UserID))
		return domainCapabilityResult(result, err)
	default:
		return domainMethodNotFound("auth", method)
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
