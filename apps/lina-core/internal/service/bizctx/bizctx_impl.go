// bizctx_impl.go implements request business-context storage and retrieval.
// It keeps the context key handling centralized so middleware, controllers,
// and services share one request-scoped identity, tenant, role, and permission
// snapshot without rebuilding context state downstream.

package bizctx

import (
	"context"
	"lina-core/internal/model"

	"github.com/gogf/gf/v2/net/ghttp"
)

// Init initializes and injects business context into request.
func (s *serviceImpl) Init(r *ghttp.Request, ctx *model.Context) {
	r.SetCtxVar(ContextKey, ctx)
}

// Get retrieves business context from context.
func (s *serviceImpl) Get(ctx context.Context) *model.Context {
	value := ctx.Value(ContextKey)
	if value == nil {
		return nil
	}
	if localCtx, ok := value.(*model.Context); ok {
		return localCtx
	}
	return nil
}

// SetLocale sets locale info into business context.
func (s *serviceImpl) SetLocale(ctx context.Context, locale string) {
	if c := s.Get(ctx); c != nil {
		c.Locale = locale
	}
}

// SetUser sets user info into business context.
func (s *serviceImpl) SetUser(ctx context.Context, tokenId string, userId int, username string, status int) {
	if c := s.Get(ctx); c != nil {
		c.TokenId = tokenId
		c.UserId = userId
		c.Username = username
		c.Status = status
	}
}

// SetTenant sets tenant info into business context.
func (s *serviceImpl) SetTenant(ctx context.Context, tenantId int) {
	if c := s.Get(ctx); c != nil {
		c.TenantId = tenantId
	}
}

// SetImpersonation sets platform impersonation info into business context.
func (s *serviceImpl) SetImpersonation(ctx context.Context, actingUserId int, tenantId int, actingAsTenant bool, isImpersonation bool) {
	if c := s.Get(ctx); c != nil {
		c.ActingUserId = actingUserId
		c.TenantId = tenantId
		c.ActingAsTenant = actingAsTenant
		c.IsImpersonation = isImpersonation
	}
}

// SetUserAccess sets cached access-snapshot fields into business context.
func (s *serviceImpl) SetUserAccess(ctx context.Context, dataScope int, dataScopeUnsupported bool, unsupportedDataScope int) {
	if c := s.Get(ctx); c != nil {
		c.DataScope = dataScope
		c.DataScopeUnsupported = dataScopeUnsupported
		c.UnsupportedDataScope = unsupportedDataScope
	}
}
