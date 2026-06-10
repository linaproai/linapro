// This file implements guest-side authentication token hostcall clients and
// the authentication namespace that links token and authz sub-capabilities.

package domainhostcall

import (
	"context"

	"lina-core/pkg/plugin/capability/authcap"
	"lina-core/pkg/plugin/capability/authcap/authz"
	"lina-core/pkg/plugin/capability/authcap/token"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// authService exposes authentication sub-capabilities through host services.
type authService struct{ baseService }

// authTokenService adapts token handoff calls to host services.
type authTokenService struct{ baseService }

// Auth creates the authentication and authorization guest namespace.
func Auth(invoker Invoker) authcap.Service {
	return authService{baseService: newBaseService(invoker)}
}

// Token returns tenant-token handoff operations.
func (s authService) Token() token.Service {
	return authTokenService{baseService: s.baseService}
}

// Authz returns authorization-domain ordinary operations.
func (s authService) Authz() authz.Service {
	return authzService{baseService: s.baseService}
}

// SelectTenant consumes a pre-login token and issues a tenant-bound token.
func (s authTokenService) SelectTenant(_ context.Context, in token.SelectTenantInput) (*token.TenantTokenOutput, error) {
	out := &token.TenantTokenOutput{}
	err := s.callJSONRequest(protocol.HostServiceAuth, protocol.HostServiceMethodAuthSelectTenant, in, out)
	return out, err
}

// SwitchTenant validates membership, revokes the current token, and issues a new token.
func (s authTokenService) SwitchTenant(_ context.Context, in token.SwitchTenantInput) (*token.TenantTokenOutput, error) {
	out := &token.TenantTokenOutput{}
	err := s.callJSONRequest(protocol.HostServiceAuth, protocol.HostServiceMethodAuthSwitchTenant, in, out)
	return out, err
}

// IssueImpersonationToken asks the host to issue one impersonation token.
func (s authTokenService) IssueImpersonationToken(_ context.Context, in token.ImpersonationTokenIssueInput) (*token.ImpersonationTokenOutput, error) {
	out := &token.ImpersonationTokenOutput{}
	err := s.callJSONRequest(protocol.HostServiceAuth, protocol.HostServiceMethodAuthIssueImpersonationToken, in, out)
	return out, err
}

// RevokeImpersonationToken asks the host to revoke one impersonation token.
func (s authTokenService) RevokeImpersonationToken(_ context.Context, in token.ImpersonationTokenRevokeInput) error {
	return s.callJSONRequest(protocol.HostServiceAuth, protocol.HostServiceMethodAuthRevokeImpersonationToken, in, nil)
}

var (
	_ authcap.Service = (*authService)(nil)
	_ token.Service   = (*authTokenService)(nil)
)
