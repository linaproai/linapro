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
	// Authz returns authorization-domain ordinary capability operations.
	Authz() authz.Service
}

// AdminService aggregates authentication and authorization management commands.
type AdminService interface {
	// Authz returns authorization-domain management commands.
	Authz() authz.AdminService
}

// serviceImpl stores authentication and authorization sub capability services.
type serviceImpl struct {
	token token.Service
	authz authz.Service
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// adminServiceImpl stores authentication and authorization management services.
type adminServiceImpl struct {
	authz authz.AdminService
}

// Ensure adminServiceImpl implements AdminService.
var _ AdminService = (*adminServiceImpl)(nil)

// New creates an authentication capability namespace from explicit sub services.
func New(tokenSvc token.Service, authzSvc authz.Service) Service {
	return &serviceImpl{token: tokenSvc, authz: authzSvc}
}

// NewAdmin creates an authentication management namespace from explicit sub services.
func NewAdmin(authzSvc authz.AdminService) AdminService {
	return &adminServiceImpl{authz: authzSvc}
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

// Authz returns authorization-domain management commands.
func (s *adminServiceImpl) Authz() authz.AdminService {
	if s == nil {
		return nil
	}
	return s.authz
}
