// This file defines the public HTTP route registrar contract exposed to source
// plugins and the lightweight request metadata helpers used by host adapters.

package pluginhost

import (
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
)

// sourceRouteCtxVarPluginID stores the source-plugin id that owns the matched route.
const sourceRouteCtxVarPluginID gctx.StrKey = "plugin_source_route_plugin_id"

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
	// RequestBodyLimit returns the host request-body limit middleware.
	RequestBodyLimit() RouteMiddleware
	// Ctx returns the host business-context injection middleware.
	Ctx() RouteMiddleware
	// Auth returns the host JWT authentication middleware.
	Auth() RouteMiddleware
	// Tenancy returns the host tenant-resolution middleware.
	Tenancy() RouteMiddleware
	// Permission returns the host declarative permission middleware.
	Permission() RouteMiddleware
}

// RouteRegistrar exposes plugin route group registration helpers for one plugin.
type RouteRegistrar interface {
	// APIPrefix returns the mandatory public API prefix for this plugin.
	APIPrefix() string
	// Err returns the first route registration error captured by the registrar.
	Err() error
	// Group registers one plugin route group bound to the dedicated plugin router root.
	Group(prefix string, register func(group RouteGroup))
	// Middlewares returns the published host middlewares available to plugins.
	Middlewares() RouteMiddlewares
	// RouteBindings returns the source-plugin route bindings captured by the host.
	RouteBindings() []SourceRouteBinding
}

// RouteGroup exposes one host-observable subset of GoFrame router-group
// registration methods for source plugins.
type RouteGroup interface {
	// Err returns the first route registration error captured by this group.
	Err() error
	// Group registers one nested route group.
	Group(prefix string, register func(group RouteGroup))
	// Middleware appends one or more middleware handlers to the current group.
	Middleware(handlers ...RouteMiddleware)
	// Bind registers one or more handlers or controller objects.
	Bind(handlerOrObject ...interface{})
	// ALL registers one handler for all HTTP methods.
	ALL(pattern string, object interface{}, params ...interface{})
	// GET registers one GET handler.
	GET(pattern string, object interface{}, params ...interface{})
	// PUT registers one PUT handler.
	PUT(pattern string, object interface{}, params ...interface{})
	// POST registers one POST handler.
	POST(pattern string, object interface{}, params ...interface{})
	// DELETE registers one DELETE handler.
	DELETE(pattern string, object interface{}, params ...interface{})
	// PATCH registers one PATCH handler.
	PATCH(pattern string, object interface{}, params ...interface{})
	// HEAD registers one HEAD handler.
	HEAD(pattern string, object interface{}, params ...interface{})
	// CONNECT registers one CONNECT handler.
	CONNECT(pattern string, object interface{}, params ...interface{})
	// OPTIONS registers one OPTIONS handler.
	OPTIONS(pattern string, object interface{}, params ...interface{})
	// TRACE registers one TRACE handler.
	TRACE(pattern string, object interface{}, params ...interface{})
}

// routeMiddlewares stores the published host middleware directory that source
// plugins are allowed to reuse.
type routeMiddlewares struct {
	neverDoneCtx    RouteMiddleware
	handlerResponse RouteMiddleware
	cors            RouteMiddleware
	requestBody     RouteMiddleware
	ctx             RouteMiddleware
	auth            RouteMiddleware
	tenancy         RouteMiddleware
	permission      RouteMiddleware
}

// NewRouteMiddlewares creates and returns a new published host middleware directory for plugins.
func NewRouteMiddlewares(
	neverDoneCtx RouteMiddleware,
	handlerResponse RouteMiddleware,
	cors RouteMiddleware,
	requestBody RouteMiddleware,
	ctx RouteMiddleware,
	auth RouteMiddleware,
	tenancy RouteMiddleware,
	permission RouteMiddleware,
) RouteMiddlewares {
	return &routeMiddlewares{
		neverDoneCtx:    neverDoneCtx,
		handlerResponse: handlerResponse,
		cors:            cors,
		requestBody:     requestBody,
		ctx:             ctx,
		auth:            auth,
		tenancy:         tenancy,
		permission:      permission,
	}
}

// SetSourcePluginIDForRequest attaches the source-plugin id that owns the
// matched route to the current request for downstream host middleware and
// capability code.
func SetSourcePluginIDForRequest(request *ghttp.Request, pluginID string) {
	if request == nil {
		return
	}
	request.SetCtxVar(sourceRouteCtxVarPluginID, strings.TrimSpace(pluginID))
}

// SourcePluginIDFromRequest returns the source-plugin id attached to the current
// request, or an empty string when the request is not handled by a source plugin.
func SourcePluginIDFromRequest(request *ghttp.Request) string {
	if request == nil {
		return ""
	}
	return strings.TrimSpace(request.GetCtxVar(sourceRouteCtxVarPluginID).String())
}

// NeverDoneCtx returns the published never-done context middleware.
func (m *routeMiddlewares) NeverDoneCtx() RouteMiddleware {
	if m == nil {
		return nil
	}
	return m.neverDoneCtx
}

// HandlerResponse returns the published unified response middleware.
func (m *routeMiddlewares) HandlerResponse() RouteMiddleware {
	if m == nil {
		return nil
	}
	return m.handlerResponse
}

// CORS returns the published CORS middleware.
func (m *routeMiddlewares) CORS() RouteMiddleware {
	if m == nil {
		return nil
	}
	return m.cors
}

// RequestBodyLimit returns the published request-body limit middleware.
func (m *routeMiddlewares) RequestBodyLimit() RouteMiddleware {
	if m == nil {
		return nil
	}
	return m.requestBody
}

// Ctx returns the published business-context injection middleware.
func (m *routeMiddlewares) Ctx() RouteMiddleware {
	if m == nil {
		return nil
	}
	return m.ctx
}

// Auth returns the published authentication middleware.
func (m *routeMiddlewares) Auth() RouteMiddleware {
	if m == nil {
		return nil
	}
	return m.auth
}

// Tenancy returns the published tenant-resolution middleware.
func (m *routeMiddlewares) Tenancy() RouteMiddleware {
	if m == nil {
		return nil
	}
	return m.tenancy
}

// Permission returns the published declarative permission middleware.
func (m *routeMiddlewares) Permission() RouteMiddleware {
	if m == nil {
		return nil
	}
	return m.permission
}
