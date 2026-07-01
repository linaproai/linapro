// This file owns the source-plugin HTTP registrar implementation used by
// integration startup. Keeping it in integration lets route and middleware
// guards reuse plugin enablement services directly.

package integration

import (
	"context"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/pluginhost"
)

// sourceHTTPRegistrar is the host-owned HTTP registrar for one source-plugin
// registration session.
type sourceHTTPRegistrar struct {
	routes            *sourceRouteRegistrar
	globalMiddlewares *sourceGlobalMiddlewareRegistrar
	services          capability.Services
}

// sourceGlobalMiddlewareRegistrar registers guarded global HTTP middleware.
type sourceGlobalMiddlewareRegistrar struct {
	server   *ghttp.Server
	pluginID string
	service  *serviceImpl
}

// sourceRouteRegistrar registers source-plugin HTTP routes under host guards.
type sourceRouteRegistrar struct {
	group       *ghttp.RouterGroup
	pluginID    string
	service     *serviceImpl
	middlewares pluginhost.RouteMiddlewares
	bindingsMu  sync.RWMutex
	bindings    []pluginhost.SourceRouteBinding
	errMu       sync.RWMutex
	err         error
}

// sourceRouteGroup adapts one GoFrame router group to the reduced plugin
// RouteGroup contract while preserving host-side route capture.
type sourceRouteGroup struct {
	group     *ghttp.RouterGroup
	pluginID  string
	prefix    string
	bindRoute func(bindings ...pluginhost.SourceRouteBinding)
	setError  func(error)
	getError  func() error
}

var (
	_ pluginhost.HTTPRegistrar             = (*sourceHTTPRegistrar)(nil)
	_ pluginhost.GlobalMiddlewareRegistrar = (*sourceGlobalMiddlewareRegistrar)(nil)
	_ pluginhost.RouteRegistrar            = (*sourceRouteRegistrar)(nil)
	_ pluginhost.RouteGroup                = (*sourceRouteGroup)(nil)
)

// newSourceHTTPRegistrar creates one host-owned HTTP registrar for a source plugin.
func newSourceHTTPRegistrar(
	server *ghttp.Server,
	pluginGroup *ghttp.RouterGroup,
	pluginID string,
	middlewares pluginhost.RouteMiddlewares,
	service *serviceImpl,
) pluginhost.HTTPRegistrar {
	normalizedPluginID := strings.TrimSpace(pluginID)
	var services capability.Services
	if service != nil {
		services = service.sourceServicesForPlugin(normalizedPluginID)
	}
	return &sourceHTTPRegistrar{
		routes: newSourceRouteRegistrar(
			pluginGroup,
			normalizedPluginID,
			middlewares,
			service,
		),
		globalMiddlewares: &sourceGlobalMiddlewareRegistrar{
			server:   server,
			pluginID: normalizedPluginID,
			service:  service,
		},
		services: services,
	}
}

// newSourceRouteRegistrar creates one guarded route registrar for a source plugin.
func newSourceRouteRegistrar(
	pluginGroup *ghttp.RouterGroup,
	pluginID string,
	middlewares pluginhost.RouteMiddlewares,
	service *serviceImpl,
) *sourceRouteRegistrar {
	return &sourceRouteRegistrar{
		group:       pluginGroup,
		pluginID:    strings.TrimSpace(pluginID),
		service:     service,
		middlewares: middlewares,
	}
}

// Routes returns the route-group registrar used to bind plugin-owned HTTP handlers.
func (r *sourceHTTPRegistrar) Routes() pluginhost.RouteRegistrar {
	if r == nil {
		return nil
	}
	return r.routes
}

// GlobalMiddlewares returns the guarded global middleware registrar for request-level extensions.
func (r *sourceHTTPRegistrar) GlobalMiddlewares() pluginhost.GlobalMiddlewareRegistrar {
	if r == nil {
		return nil
	}
	return r.globalMiddlewares
}

// Services returns the host-published runtime services for source-plugin construction.
func (r *sourceHTTPRegistrar) Services() capability.Services {
	if r == nil {
		return nil
	}
	return r.services
}

// Bind registers one guarded global middleware on the supplied GoFrame route pattern.
func (r *sourceGlobalMiddlewareRegistrar) Bind(
	scope pluginhost.MiddlewareScope,
	handler pluginhost.MiddlewareHandler,
) error {
	if r == nil || r.server == nil {
		return nil
	}
	if handler == nil {
		return gerror.New("pluginhost: global middleware handler is nil")
	}

	normalizedScope := normalizeMiddlewareScope(scope)
	r.server.BindMiddleware(normalizedScope, func(req *ghttp.Request) {
		if !r.canRun(req.Context()) {
			req.Middleware.Next()
			return
		}
		handler(req)
	})
	return nil
}

// canRun reports whether the owning plugin may execute request middleware.
func (r *sourceGlobalMiddlewareRegistrar) canRun(ctx context.Context) bool {
	if r == nil || r.service == nil {
		return false
	}
	return r.service.canExposePluginBusinessEntries(ctx, r.pluginID)
}

// APIPrefix returns the mandatory public API prefix for this plugin.
func (r *sourceRouteRegistrar) APIPrefix() string {
	if r == nil {
		return ""
	}
	return pluginAPIPrefix(r.pluginID)
}

// Err returns the first route registration error captured by this registrar.
func (r *sourceRouteRegistrar) Err() error {
	if r == nil {
		return nil
	}
	r.errMu.RLock()
	defer r.errMu.RUnlock()
	return r.err
}

// Group registers one plugin route group bound to the dedicated plugin router root.
func (r *sourceRouteRegistrar) Group(prefix string, register func(group pluginhost.RouteGroup)) {
	if r == nil || r.group == nil || register == nil {
		return
	}

	normalizedPrefix := normalizeRoutePrefix(prefix)
	if err := validateSourceRoutePrefix(r.pluginID, normalizedPrefix); err != nil {
		r.setError(err)
		return
	}
	r.group.Group(normalizedPrefix, func(group *ghttp.RouterGroup) {
		group.Middleware(func(req *ghttp.Request) {
			if !r.allow(req) {
				return
			}
			req.Middleware.Next()
		})
		register(&sourceRouteGroup{
			group:     group,
			pluginID:  r.pluginID,
			prefix:    normalizedPrefix,
			bindRoute: r.appendBindings,
			setError:  r.setError,
			getError:  r.Err,
		})
	})
}

// Middlewares returns the published host middlewares available to plugins.
func (r *sourceRouteRegistrar) Middlewares() pluginhost.RouteMiddlewares {
	if r == nil {
		return nil
	}
	return r.middlewares
}

// RouteBindings returns the source-plugin route bindings captured by the host.
func (r *sourceRouteRegistrar) RouteBindings() []pluginhost.SourceRouteBinding {
	if r == nil {
		return nil
	}
	r.bindingsMu.RLock()
	defer r.bindingsMu.RUnlock()
	return pluginhost.CloneSourceRouteBindings(r.bindings)
}

// allow rejects plugin route traffic when the owning plugin is not currently
// enabled, matching the host-side plugin state governance used elsewhere.
func (r *sourceRouteRegistrar) allow(req *ghttp.Request) bool {
	if req == nil {
		return false
	}
	if r == nil || r.service == nil || !r.service.canExposePluginBusinessEntries(req.Context(), r.pluginID) {
		req.Response.WriteStatus(404)
		req.ExitAll()
		return false
	}
	pluginhost.SetSourcePluginIDForRequest(req, r.pluginID)
	return true
}

// appendBindings appends captured plugin route bindings to the registrar-local
// snapshot in a thread-safe way.
func (r *sourceRouteRegistrar) appendBindings(bindings ...pluginhost.SourceRouteBinding) {
	if r == nil || len(bindings) == 0 {
		return
	}
	r.bindingsMu.Lock()
	defer r.bindingsMu.Unlock()
	r.bindings = append(r.bindings, bindings...)
}

// setError records the first route registration error for the registrar.
func (r *sourceRouteRegistrar) setError(err error) {
	if r == nil || err == nil {
		return
	}
	r.errMu.Lock()
	defer r.errMu.Unlock()
	if r.err == nil {
		r.err = err
	}
}

// Err returns the first route registration error captured by this group.
func (g *sourceRouteGroup) Err() error {
	if g == nil || g.getError == nil {
		return nil
	}
	return g.getError()
}

// Group registers one nested route group.
func (g *sourceRouteGroup) Group(prefix string, register func(group pluginhost.RouteGroup)) {
	if g == nil || g.group == nil || register == nil {
		return
	}
	normalizedPrefix := normalizeRoutePrefix(prefix)
	joinedPrefix := joinRoutePatterns(g.prefix, normalizedPrefix)
	if err := validateSourceRoutePath(g.pluginID, joinedPrefix); err != nil {
		g.setRegistrationError(err)
		return
	}
	g.group.Group(normalizedPrefix, func(group *ghttp.RouterGroup) {
		register(&sourceRouteGroup{
			group:     group,
			pluginID:  g.pluginID,
			prefix:    joinedPrefix,
			bindRoute: g.bindRoute,
			setError:  g.setError,
			getError:  g.getError,
		})
	})
}

// Middleware appends one or more middleware handlers to the current group.
func (g *sourceRouteGroup) Middleware(handlers ...pluginhost.RouteMiddleware) {
	if g == nil || g.group == nil {
		return
	}
	g.group.Middleware(handlers...)
}

// Bind registers one or more handlers or controller objects.
func (g *sourceRouteGroup) Bind(handlerOrObject ...interface{}) {
	if g == nil || g.group == nil {
		return
	}
	for _, item := range handlerOrObject {
		bindings := captureRouteBindings(g.pluginID, g.prefix, "/", routeMethodAll, item)
		if err := validateSourceRouteBindings(g.pluginID, bindings); err != nil {
			g.setRegistrationError(err)
			return
		}
		if g.bindRoute != nil && len(bindings) > 0 {
			g.bindRoute(bindings...)
		}
	}
	g.group.Bind(handlerOrObject...)
}

// ALL registers one handler for all HTTP methods.
func (g *sourceRouteGroup) ALL(pattern string, object interface{}, params ...interface{}) {
	g.bindMethodRoute(pattern, routeMethodAll, object, params, func() {
		g.group.ALL(pattern, object, params...)
	})
}

// GET registers one GET handler.
func (g *sourceRouteGroup) GET(pattern string, object interface{}, params ...interface{}) {
	g.bindMethodRoute(pattern, "GET", object, params, func() {
		g.group.GET(pattern, object, params...)
	})
}

// PUT registers one PUT handler.
func (g *sourceRouteGroup) PUT(pattern string, object interface{}, params ...interface{}) {
	g.bindMethodRoute(pattern, "PUT", object, params, func() {
		g.group.PUT(pattern, object, params...)
	})
}

// POST registers one POST handler.
func (g *sourceRouteGroup) POST(pattern string, object interface{}, params ...interface{}) {
	g.bindMethodRoute(pattern, "POST", object, params, func() {
		g.group.POST(pattern, object, params...)
	})
}

// DELETE registers one DELETE handler.
func (g *sourceRouteGroup) DELETE(pattern string, object interface{}, params ...interface{}) {
	g.bindMethodRoute(pattern, "DELETE", object, params, func() {
		g.group.DELETE(pattern, object, params...)
	})
}

// PATCH registers one PATCH handler.
func (g *sourceRouteGroup) PATCH(pattern string, object interface{}, params ...interface{}) {
	g.bindMethodRoute(pattern, "PATCH", object, params, func() {
		g.group.PATCH(pattern, object, params...)
	})
}

// HEAD registers one HEAD handler.
func (g *sourceRouteGroup) HEAD(pattern string, object interface{}, params ...interface{}) {
	g.bindMethodRoute(pattern, "HEAD", object, params, func() {
		g.group.HEAD(pattern, object, params...)
	})
}

// CONNECT registers one CONNECT handler.
func (g *sourceRouteGroup) CONNECT(pattern string, object interface{}, params ...interface{}) {
	g.bindMethodRoute(pattern, "CONNECT", object, params, func() {
		g.group.CONNECT(pattern, object, params...)
	})
}

// OPTIONS registers one OPTIONS handler.
func (g *sourceRouteGroup) OPTIONS(pattern string, object interface{}, params ...interface{}) {
	g.bindMethodRoute(pattern, "OPTIONS", object, params, func() {
		g.group.OPTIONS(pattern, object, params...)
	})
}

// TRACE registers one TRACE handler.
func (g *sourceRouteGroup) TRACE(pattern string, object interface{}, params ...interface{}) {
	g.bindMethodRoute(pattern, "TRACE", object, params, func() {
		g.group.TRACE(pattern, object, params...)
	})
}

// bindMethodRoute captures one explicit method route before delegating to the
// underlying GoFrame router group.
func (g *sourceRouteGroup) bindMethodRoute(
	pattern string,
	method string,
	object interface{},
	_ []interface{},
	bind func(),
) {
	if g == nil || g.group == nil || bind == nil {
		return
	}
	bindings := captureRouteBindings(g.pluginID, g.prefix, pattern, method, object)
	if len(bindings) == 0 {
		finalPath := joinRoutePatterns(g.prefix, pattern)
		if err := validateSourceRoutePath(g.pluginID, finalPath); err != nil {
			g.setRegistrationError(err)
			return
		}
	}
	if err := validateSourceRouteBindings(g.pluginID, bindings); err != nil {
		g.setRegistrationError(err)
		return
	}
	if g.bindRoute != nil {
		g.bindRoute(bindings...)
	}
	bind()
}

// setRegistrationError records one invalid route registration on the owning registrar.
func (g *sourceRouteGroup) setRegistrationError(err error) {
	if g == nil || err == nil || g.setError == nil {
		return
	}
	g.setError(err)
}

// normalizeMiddlewareScope canonicalizes one raw GoFrame middleware pattern while
// keeping plugin-declared wildcard semantics intact.
func normalizeMiddlewareScope(scope pluginhost.MiddlewareScope) string {
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

// normalizeRoutePrefix canonicalizes plugin-owned route group prefixes so
// callers can register `api/v1`, `/api/v1/`, or `/` interchangeably.
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

// pluginAPIPrefix returns the mandatory plugin-owned namespace for one plugin ID.
func pluginAPIPrefix(pluginID string) string {
	return pluginhost.PluginAPINamespacePrefix + "/" + strings.Trim(strings.TrimSpace(pluginID), "/")
}

// validateSourceRoutePrefix rejects source-plugin registrations that try to use
// `/x` for anything outside the owning plugin namespace.
func validateSourceRoutePrefix(pluginID string, prefix string) error {
	normalizedPrefix := normalizeRoutePrefix(prefix)
	return validateSourceRoutePath(pluginID, normalizedPrefix)
}

// validateSourceRoutePath rejects final source-plugin route paths that try to
// use `/x` for anything outside the owning plugin namespace.
func validateSourceRoutePath(pluginID string, routePath string) error {
	if routePathHasTraversal(routePath) {
		return gerror.Newf("source plugin %s cannot register route with traversal segments: %s", strings.TrimSpace(pluginID), routePath)
	}
	normalizedPrefix := normalizeRoutePrefix(routePath)
	if normalizedPrefix == "/" {
		return nil
	}
	if normalizedPrefix != pluginhost.PluginAPINamespacePrefix &&
		!strings.HasPrefix(normalizedPrefix, pluginhost.PluginAPINamespacePrefix+"/") {
		return nil
	}
	apiPrefix := pluginAPIPrefix(pluginID)
	if normalizedPrefix == apiPrefix || strings.HasPrefix(normalizedPrefix, apiPrefix+"/") {
		return nil
	}
	return gerror.Newf(
		"source plugin %s cannot register route outside its /x namespace: %s; use %s for plugin routes",
		strings.TrimSpace(pluginID),
		normalizedPrefix,
		apiPrefix,
	)
}

// routePathHasTraversal reports whether a plugin-supplied route contains dot
// segments that could change the effective reserved namespace after cleaning.
func routePathHasTraversal(routePath string) bool {
	parts := strings.Split(strings.ReplaceAll(strings.TrimSpace(routePath), "\\", "/"), "/")
	for _, part := range parts {
		if part == ".." || part == "." {
			return true
		}
	}
	return false
}

// validateSourceRouteBindings checks documentable DTO routes after their
// GoFrame metadata has resolved to final public paths.
func validateSourceRouteBindings(pluginID string, bindings []pluginhost.SourceRouteBinding) error {
	for _, binding := range bindings {
		if err := validateSourceRoutePath(pluginID, binding.Path); err != nil {
			return err
		}
	}
	return nil
}
