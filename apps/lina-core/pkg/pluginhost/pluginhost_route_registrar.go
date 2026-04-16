// This file defines the public HTTP route registrar contract exposed to source
// plugins and the guarded host implementation that enforces plugin state.

package pluginhost

import (
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
)

// PluginEnabledChecker defines one host callback that reports whether a plugin is currently enabled.
type PluginEnabledChecker func(pluginID string) bool

// RouteMiddleware defines one plugin-usable HTTP middleware.
type RouteMiddleware = ghttp.HandlerFunc

// RouteMiddlewares exposes published host middlewares for plugin route composition.
type RouteMiddlewares interface {
	// NeverDoneCtx returns the host never-done context middleware.
	NeverDoneCtx() RouteMiddleware
	// HandlerResponse returns the host unified response middleware.
	HandlerResponse() RouteMiddleware
	// CORS returns the host CORS middleware.
	CORS() RouteMiddleware
	// Ctx returns the host business-context injection middleware.
	Ctx() RouteMiddleware
	// Auth returns the host JWT authentication middleware.
	Auth() RouteMiddleware
	// OperLog returns the host operation-log middleware.
	OperLog() RouteMiddleware
	// Permission returns the host declarative permission middleware.
	Permission() RouteMiddleware
}

// RouteRegistrar exposes plugin route group registration helpers for one plugin.
type RouteRegistrar interface {
	// Group registers one plugin route group bound to the dedicated plugin router root.
	Group(prefix string, register func(group *ghttp.RouterGroup))
	// Middlewares returns the published host middlewares available to plugins.
	Middlewares() RouteMiddlewares
}

type routeRegistrar struct {
	group          *ghttp.RouterGroup
	pluginID       string
	enabledChecker PluginEnabledChecker
	middlewares    RouteMiddlewares
}

type routeMiddlewares struct {
	neverDoneCtx    RouteMiddleware
	handlerResponse RouteMiddleware
	cors            RouteMiddleware
	ctx             RouteMiddleware
	auth            RouteMiddleware
	operLog         RouteMiddleware
	permission      RouteMiddleware
}

// NewRouteMiddlewares creates and returns a new published host middleware directory for plugins.
func NewRouteMiddlewares(
	neverDoneCtx RouteMiddleware,
	handlerResponse RouteMiddleware,
	cors RouteMiddleware,
	ctx RouteMiddleware,
	auth RouteMiddleware,
	operLog RouteMiddleware,
	permission RouteMiddleware,
) RouteMiddlewares {
	return &routeMiddlewares{
		neverDoneCtx:    neverDoneCtx,
		handlerResponse: handlerResponse,
		cors:            cors,
		ctx:             ctx,
		auth:            auth,
		operLog:         operLog,
		permission:      permission,
	}
}

// NewRouteRegistrar creates and returns a new RouteRegistrar instance.
func NewRouteRegistrar(
	pluginGroup *ghttp.RouterGroup,
	pluginID string,
	enabledChecker PluginEnabledChecker,
	middlewares RouteMiddlewares,
) RouteRegistrar {
	return &routeRegistrar{
		group:          pluginGroup,
		pluginID:       pluginID,
		enabledChecker: enabledChecker,
		middlewares:    middlewares,
	}
}

// Group registers one plugin route group bound to the dedicated plugin router root.
func (r *routeRegistrar) Group(prefix string, register func(group *ghttp.RouterGroup)) {
	if r == nil || r.group == nil || register == nil {
		return
	}

	r.group.Group(normalizeRoutePrefix(prefix), func(group *ghttp.RouterGroup) {
		group.Middleware(func(req *ghttp.Request) {
			if !r.allow(req) {
				return
			}
			req.Middleware.Next()
		})
		register(group)
	})
}

// Middlewares returns the published host middlewares available to plugins.
func (r *routeRegistrar) Middlewares() RouteMiddlewares {
	if r == nil {
		return nil
	}
	return r.middlewares
}

func (r *routeRegistrar) allow(req *ghttp.Request) bool {
	if req == nil {
		return false
	}
	if r.enabledChecker != nil && !r.enabledChecker(r.pluginID) {
		req.Response.WriteStatus(404)
		req.ExitAll()
		return false
	}
	return true
}

func (m *routeMiddlewares) NeverDoneCtx() RouteMiddleware {
	if m == nil {
		return nil
	}
	return m.neverDoneCtx
}

func (m *routeMiddlewares) HandlerResponse() RouteMiddleware {
	if m == nil {
		return nil
	}
	return m.handlerResponse
}

func (m *routeMiddlewares) CORS() RouteMiddleware {
	if m == nil {
		return nil
	}
	return m.cors
}

func (m *routeMiddlewares) Ctx() RouteMiddleware {
	if m == nil {
		return nil
	}
	return m.ctx
}

func (m *routeMiddlewares) Auth() RouteMiddleware {
	if m == nil {
		return nil
	}
	return m.auth
}

func (m *routeMiddlewares) OperLog() RouteMiddleware {
	if m == nil {
		return nil
	}
	return m.operLog
}

func (m *routeMiddlewares) Permission() RouteMiddleware {
	if m == nil {
		return nil
	}
	return m.permission
}

func normalizeRoutePrefix(prefix string) string {
	trimmed := strings.TrimSpace(prefix)
	if trimmed == "" || trimmed == "/" {
		return "/"
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	return strings.TrimRight(trimmed, "/")
}
