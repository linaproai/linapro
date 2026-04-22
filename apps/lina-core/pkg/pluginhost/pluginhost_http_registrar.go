// This file defines the published HTTP registrar exposed to source plugins so
// they can register route groups and host-governed global middleware without
// touching the raw GoFrame server instance.

package pluginhost

import (
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
)

// MiddlewareScope is one raw GoFrame route pattern used to bind global HTTP middleware.
type MiddlewareScope string

// MiddlewareHandler defines one plugin-owned global HTTP middleware handler.
type MiddlewareHandler = ghttp.HandlerFunc

// GlobalMiddlewareRegistrar exposes host-governed global middleware registration for one plugin.
type GlobalMiddlewareRegistrar interface {
	// Bind registers one guarded global middleware on the supplied GoFrame route pattern.
	Bind(scope MiddlewareScope, handler MiddlewareHandler)
}

// HTTPRegistrar exposes the complete HTTP registration surface published to one source plugin.
type HTTPRegistrar interface {
	// Routes returns the route-group registrar used to bind plugin-owned HTTP handlers.
	Routes() RouteRegistrar
	// GlobalMiddlewares returns the guarded global middleware registrar for request-level extensions.
	GlobalMiddlewares() GlobalMiddlewareRegistrar
}

// httpRegistrar is the host-owned HTTPRegistrar implementation for one source-plugin registration session.
type httpRegistrar struct {
	routes            RouteRegistrar
	globalMiddlewares GlobalMiddlewareRegistrar
}

// globalMiddlewareRegistrar is the host-owned implementation of the published
// global middleware registrar.
type globalMiddlewareRegistrar struct {
	server         *ghttp.Server
	pluginID       string
	enabledChecker PluginEnabledChecker
}

// NewHTTPRegistrar creates and returns a new HTTPRegistrar instance.
func NewHTTPRegistrar(
	server *ghttp.Server,
	pluginGroup *ghttp.RouterGroup,
	pluginID string,
	enabledChecker PluginEnabledChecker,
	middlewares RouteMiddlewares,
) HTTPRegistrar {
	return &httpRegistrar{
		routes: NewRouteRegistrar(
			pluginGroup,
			pluginID,
			enabledChecker,
			middlewares,
		),
		globalMiddlewares: NewGlobalMiddlewareRegistrar(
			server,
			pluginID,
			enabledChecker,
		),
	}
}

// NewGlobalMiddlewareRegistrar creates and returns a new guarded global middleware registrar.
func NewGlobalMiddlewareRegistrar(
	server *ghttp.Server,
	pluginID string,
	enabledChecker PluginEnabledChecker,
) GlobalMiddlewareRegistrar {
	return &globalMiddlewareRegistrar{
		server:         server,
		pluginID:       pluginID,
		enabledChecker: enabledChecker,
	}
}

// Routes returns the route-group registrar used to bind plugin-owned HTTP handlers.
func (r *httpRegistrar) Routes() RouteRegistrar {
	if r == nil {
		return nil
	}
	return r.routes
}

// GlobalMiddlewares returns the guarded global middleware registrar for request-level extensions.
func (r *httpRegistrar) GlobalMiddlewares() GlobalMiddlewareRegistrar {
	if r == nil {
		return nil
	}
	return r.globalMiddlewares
}

// Bind registers one guarded global middleware on the supplied GoFrame route pattern.
func (r *globalMiddlewareRegistrar) Bind(scope MiddlewareScope, handler MiddlewareHandler) {
	if r == nil || r.server == nil {
		return
	}
	if handler == nil {
		panic("pluginhost: global middleware handler is nil")
	}

	normalizedScope := normalizeMiddlewareScope(scope)
	r.server.BindMiddleware(normalizedScope, func(req *ghttp.Request) {
		if r.enabledChecker != nil && !r.enabledChecker(r.pluginID) {
			req.Middleware.Next()
			return
		}
		handler(req)
	})
}

// normalizeMiddlewareScope canonicalizes one raw GoFrame middleware pattern while
// keeping plugin-declared wildcard semantics intact.
func normalizeMiddlewareScope(scope MiddlewareScope) string {
	trimmed := strings.TrimSpace(string(scope))
	if trimmed == "" {
		return "/*"
	}
	if strings.Contains(trimmed, ":/") {
		return trimmed
	}
	if !strings.HasPrefix(trimmed, "/") {
		return "/" + trimmed
	}
	return trimmed
}
