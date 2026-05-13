// Package resolver implements the plugin-owned tenant resolution chain.
package resolver

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"
	plugincontract "lina-core/pkg/pluginservice/contract"
	"lina-plugin-multi-tenant/backend/internal/service/membership"
	"lina-plugin-multi-tenant/backend/internal/service/shared"
)

// Result is the output of one resolver.
type Result struct {
	TenantID       int64
	Source         string
	ActingAsTenant bool
}

// Resolver resolves a tenant from a request.
type Resolver interface {
	// Name returns the configured resolver name.
	Name() string
	// Resolve returns a tenant result and whether this resolver matched.
	Resolve(ctx context.Context, r *ghttp.Request, identity Identity, config Config) (*Result, bool, error)
}

// Identity describes the authenticated user known to the resolver chain.
type Identity struct {
	UserID          int64 // UserID is the authenticated host user identifier.
	TenantID        int64 // TenantID is the tenant already attached by host JWT auth.
	ActingUserID    int64 // ActingUserID is the real platform user during impersonation.
	ActingAsTenant  bool  // ActingAsTenant reports whether the request is operating through a tenant view.
	IsImpersonation bool  // IsImpersonation reports whether the token is an impersonation token.
	IsPlatform      bool  // IsPlatform reports whether the caller is in platform context.
}

// Config defines resolver chain behavior.
type Config struct {
	Chain              []string
	ReservedSubdomains []string
	RootDomain         string
	OnAmbiguous        string
}

// Service defines tenant resolution operations.
type Service interface {
	// Resolve runs the configured resolver chain.
	Resolve(ctx context.Context, r *ghttp.Request, config Config) (*Result, error)
	// Register registers or replaces one resolver implementation.
	Register(resolver Resolver)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc     plugincontract.BizCtxService
	membershipSvc membership.Service
	resolvers     map[string]Resolver
}

// New creates and returns a resolver service with the built-in resolver set.
func New(bizCtxSvc plugincontract.BizCtxService, membershipSvc membership.Service) Service {
	s := &serviceImpl{
		bizCtxSvc:     bizCtxSvc,
		membershipSvc: membershipSvc,
		resolvers:     make(map[string]Resolver),
	}
	s.Register(overrideResolver{})
	s.Register(jwtResolver{})
	s.Register(sessionResolver{})
	s.Register(headerResolver{})
	s.Register(subdomainResolver{})
	s.Register(defaultResolver{membershipSvc: s.membershipSvc})
	return s
}

// Register registers or replaces one resolver implementation.
func (s *serviceImpl) Register(resolver Resolver) {
	if resolver == nil {
		return
	}
	s.resolvers[resolver.Name()] = resolver
}

// Resolve runs the configured resolver chain.
func (s *serviceImpl) Resolve(ctx context.Context, r *ghttp.Request, config Config) (*Result, error) {
	bizCtx := s.bizCtxSvc.Current(ctx)
	identity := Identity{
		UserID:          int64(bizCtx.UserID),
		TenantID:        int64(bizCtx.TenantID),
		ActingUserID:    int64(bizCtx.ActingUserID),
		ActingAsTenant:  bizCtx.ActingAsTenant,
		IsImpersonation: bizCtx.IsImpersonation,
		IsPlatform:      int64(bizCtx.TenantID) == shared.PlatformTenantID,
	}
	chain := config.Chain
	if len(chain) == 0 {
		chain = shared.DefaultResolverChain()
	}
	for _, name := range chain {
		resolverImpl := s.resolvers[strings.TrimSpace(name)]
		if resolverImpl == nil {
			continue
		}
		result, ok, err := resolverImpl.Resolve(ctx, r, identity, config)
		if err != nil || !ok {
			if err != nil {
				return nil, err
			}
			continue
		}
		if err = s.validateMembership(ctx, identity, result); err != nil {
			return nil, err
		}
		return result, nil
	}
	switch config.OnAmbiguous {
	case shared.OnAmbiguousReject:
		return nil, bizerr.NewCode(CodeTenantForbidden)
	default:
		return nil, bizerr.NewCode(CodeTenantRequired)
	}
}

// validateMembership verifies non-platform users belong to the resolved tenant.
func (s *serviceImpl) validateMembership(ctx context.Context, identity Identity, result *Result) error {
	if result == nil || result.TenantID == shared.PlatformTenantID {
		return nil
	}
	if result.ActingAsTenant || identity.ActingAsTenant || identity.IsImpersonation {
		return nil
	}
	if identity.UserID == 0 {
		return nil
	}
	_, err := s.membershipSvc.GetByUserAndTenant(ctx, identity.UserID, result.TenantID)
	return err
}
