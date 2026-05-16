// middleware_impl.go implements request middleware for sessions, CORS,
// localization, and plugin route filtering. It relies on the injected auth,
// tenant, i18n, and plugin services so request paths share runtime state and
// do not create independent service graphs while handling HTTP traffic.

package middleware

import (
	"lina-core/internal/model"
	"lina-core/internal/service/session"
	"lina-core/pkg/pluginhost"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/i18n/gi18n"
	"github.com/gogf/gf/v2/net/ghttp"
)

// SessionStore returns the session store for external use (e.g., cleanup tasks).
func (s *serviceImpl) SessionStore() session.Store {
	return s.authSvc.SessionStore()
}

// PublishedRouteMiddlewares returns the published host middleware directory for plugin route composition.
func (s *serviceImpl) PublishedRouteMiddlewares() pluginhost.RouteMiddlewares {
	if s == nil {
		return nil
	}

	return pluginhost.NewRouteMiddlewares(
		ghttp.MiddlewareNeverDoneCtx,
		s.Response,
		s.CORS,
		s.RequestBodyLimit,
		s.Ctx,
		s.Auth,
		s.Tenancy,
		s.Permission,
	)
}

// Ctx injects business context into request.
func (s *serviceImpl) Ctx(r *ghttp.Request) {
	customCtx := &model.Context{}
	s.bizCtxSvc.Init(r, customCtx)
	locale := s.i18nSvc.ResolveRequestLocale(r)
	r.SetCtx(gi18n.WithLanguage(r.Context(), locale))
	s.bizCtxSvc.SetLocale(r.Context(), locale)
	r.Response.Header().Set("Content-Language", locale)
	r.Middleware.Next()
}

// CORS handles cross-origin requests.
func (s *serviceImpl) CORS(r *ghttp.Request) {
	r.Response.CORSDefault()
	r.Middleware.Next()
}

// Auth validates JWT token and injects user info into context.
func (s *serviceImpl) Auth(r *ghttp.Request) {
	tokenHeader := r.GetHeader("Authorization")
	if tokenHeader == "" {
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(tokenHeader, "Bearer ")
	if tokenString == tokenHeader {
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}

	claims, err := s.authSvc.ParseToken(r.Context(), tokenString)
	if err != nil {
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}

	sessionTimeout, err := s.configSvc.GetSessionTimeout(r.Context())
	if err != nil {
		r.SetError(err)
		return
	}

	// Update last active time and validate session exists (supports forced logout and timeout cleanup)
	exists, err := s.authSvc.SessionStore().TouchOrValidate(
		r.Context(),
		claims.TenantId,
		claims.TokenId,
		sessionTimeout,
	)
	if err != nil || !exists {
		s.roleSvc.InvalidateTokenAccessContext(r.Context(), claims.TokenId)
		if err != nil {
			r.SetError(err)
		}
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}

	// Inject user info into business context.
	s.bizCtxSvc.SetUser(r.Context(), claims.TokenId, claims.UserId, claims.Username, claims.Status)
	s.bizCtxSvc.SetTenant(r.Context(), claims.TenantId)
	s.bizCtxSvc.SetImpersonation(
		r.Context(),
		claims.ActingUserId,
		claims.TenantId,
		claims.IsImpersonation,
		claims.IsImpersonation,
	)
	r.Middleware.Next()
}
