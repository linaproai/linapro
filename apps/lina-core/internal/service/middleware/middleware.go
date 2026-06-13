// Package middleware implements HTTP authentication, authorization, and related
// request middleware for the Lina core host service.
package middleware

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/service/auth"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/config"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	tenantcapsvc "lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/pluginhost"
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
	roleSvc   role.Service          // Role and permission service
	tenantSvc tenantMiddlewareService
}

// middlewareI18nService defines the locale and error localization capabilities middleware needs.
type middlewareI18nService interface {
	i18nsvc.LocaleResolver
	i18nsvc.Translator
}

// tenantMiddlewareService is the host-internal tenant slice needed by request
// middleware. It deliberately avoids tenant query-scope and membership write
// methods because middleware only resolves request context and platform bypass.
type tenantMiddlewareService interface {
	// Available reports whether tenant resolution is active.
	Available(ctx context.Context) bool
	// PlatformBypass reports whether the request may operate in platform scope.
	PlatformBypass(ctx context.Context) bool
	// ResolveTenant resolves tenant identity from one HTTP request.
	ResolveTenant(ctx context.Context, r *ghttp.Request) (*tenantcapsvc.ResolverResult, error)
}

// New creates a middleware service from explicit runtime-owned dependencies.
func New(authSvc auth.Service, bizCtxSvc bizctx.Service, configSvc config.Service, i18nSvc middlewareI18nService, roleSvc role.Service, tenantSvc tenantMiddlewareService) Service {
	return &serviceImpl{
		authSvc:   authSvc,
		bizCtxSvc: bizCtxSvc,
		configSvc: configSvc,
		i18nSvc:   i18nSvc,
		roleSvc:   roleSvc,
		tenantSvc: tenantSvc,
	}
}
