// This file defines the published HTTP registrar exposed to source plugins so
// they can register route groups and host-governed global middleware without
// touching the raw GoFrame server instance.

package pluginhost

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/plugin/capability"
)

// MiddlewareScope is one raw GoFrame route pattern used to bind global HTTP middleware.
type MiddlewareScope string

// MiddlewareHandler defines one plugin-owned global HTTP middleware handler.
type MiddlewareHandler = ghttp.HandlerFunc

// GlobalMiddlewareRegistrar exposes host-governed global middleware registration for one plugin.
type GlobalMiddlewareRegistrar interface {
	// Bind registers one guarded global middleware on the supplied GoFrame route pattern.
	Bind(scope MiddlewareScope, handler MiddlewareHandler) error
}

// HTTPRegistrar exposes the complete HTTP registration surface published to one source plugin.
type HTTPRegistrar interface {
	// Routes returns the route-group registrar used to bind plugin-owned HTTP handlers.
	Routes() RouteRegistrar
	// GlobalMiddlewares returns the guarded global middleware registrar for request-level extensions.
	GlobalMiddlewares() GlobalMiddlewareRegistrar
	// Services returns the host-published runtime services for source-plugin construction.
	Services() capability.Services
}
