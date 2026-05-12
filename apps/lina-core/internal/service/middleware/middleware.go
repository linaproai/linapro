// Package middleware implements HTTP authentication, authorization, and related
// request middleware for the Lina core host service.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/i18n/gi18n"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/model"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/config"
	i18nsvc "lina-core/internal/service/i18n"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	tenantcapsvc "lina-core/internal/service/tenantcap"
	"lina-core/pkg/pluginhost"
)

// Service defines the complete middleware service contract by composing
// request middleware and non-middleware support capabilities.
type Service interface {
	HTTPMiddleware
	RuntimeSupport
}

// HTTPMiddleware defines the request handlers that can be installed directly
// into GoFrame HTTP route groups.
type HTTPMiddleware interface {
	// Response serializes the unified JSON response payload.
	Response(r *ghttp.Request)
	// Ctx injects business context into request.
	Ctx(r *ghttp.Request)
	// CORS handles cross-origin requests.
	CORS(r *ghttp.Request)
	// RequestBodyLimit applies host request-body size limits before handlers parse form data.
	RequestBodyLimit(r *ghttp.Request)
	// Auth validates JWT token and injects user info into context.
	Auth(r *ghttp.Request)
	// Tenancy resolves tenant identity and injects it into context.
	Tenancy(r *ghttp.Request)
	// RequirePermission declares static permission requirements for manually registered routes.
	RequirePermission(permissions ...string) ghttp.HandlerFunc
	// Permission enforces declarative permission requirements declared on static host API handlers.
	Permission(r *ghttp.Request)
}

// RuntimeSupport defines non-middleware helpers shared with host runtime
// services and source-plugin route publication.
type RuntimeSupport interface {
	// SessionStore returns the session store for external use, such as cleanup tasks.
	SessionStore() session.Store
	// PublishedRouteMiddlewares returns the published host middleware directory for plugin route composition.
	PublishedRouteMiddlewares() pluginhost.RouteMiddlewares
}

// Interface compliance assertion for the default middleware service
// implementation.
var (
	_ Service        = (*serviceImpl)(nil)
	_ HTTPMiddleware = (*serviceImpl)(nil)
	_ RuntimeSupport = (*serviceImpl)(nil)
)

// serviceImpl implements Service.
type serviceImpl struct {
	authSvc   auth.Service          // Authentication service
	bizCtxSvc bizctx.Service        // Business context service
	configSvc config.Service        // Runtime configuration service
	i18nSvc   middlewareI18nService // i18nSvc resolves request locale and translation context.
	pluginSvc pluginsvc.Service     // Plugin service
	roleSvc   role.Service          // Role and permission service
	tenantSvc tenantcapsvc.Service  // Tenant capability service
}

// middlewareI18nService defines the locale and error localization capabilities middleware needs.
type middlewareI18nService interface {
	i18nsvc.LocaleResolver
	i18nsvc.Translator
}

// New creates and returns a new Service instance.
func New() Service {
	pluginSvc := pluginsvc.Instance()
	return &serviceImpl{
		authSvc:   auth.New(nil),
		bizCtxSvc: bizctx.New(),
		configSvc: config.New(),
		i18nSvc:   i18nsvc.New(),
		pluginSvc: pluginSvc,
		roleSvc:   role.New(pluginSvc),
		tenantSvc: tenantcapsvc.New(pluginSvc),
	}
}

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
