// Package authcap defines the authentication and authorization capability
// namespace exposed through the plugin capability directory. The package only
// aggregates token, authorization, and external-login sub capabilities; each
// sub capability keeps its own DTOs, data boundary, and error semantics.
package authcap

import (
	"lina-core/pkg/plugin/capability/authcap/authz"
	"lina-core/pkg/plugin/capability/authcap/externallogin"
	"lina-core/pkg/plugin/capability/authcap/token"
)

// Service aggregates authentication token, authorization, and external-login
// sub capabilities.
type Service interface {
	// Token returns tenant token handoff and impersonation token operations.
	Token() token.Service
	// Authz returns authorization-domain governed capability operations.
	Authz() authz.Service
	// ExternalLogin returns external-identity login operations. The returned
	// service is plugin-scoped by the host so provider ownership and the calling
	// plugin identity are enforced by the host, not by plugin input.
	ExternalLogin() externallogin.Service
}

// serviceImpl stores authentication, authorization, and external-login sub
// capability services.
type serviceImpl struct {
	token         token.Service
	authz         authz.Service
	externalLogin externallogin.Service
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// New creates an authentication capability namespace from explicit sub services.
func New(tokenSvc token.Service, authzSvc authz.Service, externalLoginSvc externallogin.Service) Service {
	return &serviceImpl{token: tokenSvc, authz: authzSvc, externalLogin: externalLoginSvc}
}

// ForPlugin returns a plugin-scoped authentication namespace. Token and Authz
// sub capabilities are shared and returned unchanged; ExternalLogin is replaced
// with the supplied plugin-bound service so external-identity login enforces the
// calling plugin's provider ownership and enablement.
func ForPlugin(base Service, externalLoginSvc externallogin.Service) Service {
	if base == nil {
		return nil
	}
	return &serviceImpl{
		token:         base.Token(),
		authz:         base.Authz(),
		externalLogin: externalLoginSvc,
	}
}

// Token returns tenant token handoff and impersonation token operations.
func (s *serviceImpl) Token() token.Service {
	if s == nil {
		return nil
	}
	return s.token
}

// Authz returns authorization-domain ordinary capability operations.
func (s *serviceImpl) Authz() authz.Service {
	if s == nil {
		return nil
	}
	return s.authz
}

// ExternalLogin returns external-identity login operations.
func (s *serviceImpl) ExternalLogin() externallogin.Service {
	if s == nil {
		return nil
	}
	return s.externalLogin
}
