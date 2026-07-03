// Package authcap defines the authentication and authorization capability
// namespace exposed through the plugin capability directory. The package only
// aggregates token and authorization sub capabilities; each sub capability keeps
// its own DTOs, data boundary, and error semantics.
package authcap

import (
	"lina-core/pkg/plugin/capability/authcap/authz"
	"lina-core/pkg/plugin/capability/authcap/token"
)

// Service aggregates authentication token and authorization sub capabilities.
type Service interface {
	// Token returns tenant token handoff and impersonation token operations.
	Token() token.Service
	// Authz returns authorization-domain governed capability operations.
	Authz() authz.Service
}

// serviceImpl stores authentication and authorization sub capability services.
type serviceImpl struct {
	token token.Service
	authz authz.Service
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// New creates an authentication capability namespace from explicit sub services.
func New(tokenSvc token.Service, authzSvc authz.Service) Service {
	return &serviceImpl{token: tokenSvc, authz: authzSvc}
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
