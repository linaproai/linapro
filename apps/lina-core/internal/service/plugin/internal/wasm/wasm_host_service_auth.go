// This file adapts dynamic-plugin token-authentication host-service calls to
// the shared capability.Services directory.

package wasm

import (
	"context"

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
