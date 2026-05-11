// Package backend wires the monitor-operlog source plugin into the host plugin registry.
package backend

import (
	"context"

	"lina-core/pkg/pluginhost"
	monitoroperlogplugin "lina-plugin-monitor-operlog"
	operlogcontroller "lina-plugin-monitor-operlog/backend/internal/controller/operlog"
	middlewaresvc "lina-plugin-monitor-operlog/backend/internal/service/middleware"
)

// monitor-operlog plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "monitor-operlog"
)

// init registers the monitor-operlog source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(monitoroperlogplugin.EmbeddedFiles)
	plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	)
	pluginhost.RegisterSourcePlugin(plugin)
}

// registerRoutes binds operation-log governance routes and audit middleware through the published host HTTP registrars.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	auditMiddlewareSvc := middlewaresvc.New()
	registrar.GlobalMiddlewares().Bind("/*", auditMiddlewareSvc.Audit)

	var (
		routes      = registrar.Routes()
		middlewares = routes.Middlewares()
	)
	routes.Group("/api/v1", func(group pluginhost.RouteGroup) {
		group.Middleware(
			middlewares.NeverDoneCtx(),
			middlewares.HandlerResponse(),
			middlewares.CORS(),
			middlewares.RequestBodyLimit(),
			middlewares.Ctx(),
		)
		group.Group("/", func(group pluginhost.RouteGroup) {
			group.Middleware(
				middlewares.Auth(),
				middlewares.Tenancy(),
				middlewares.Permission(),
			)
			group.Bind(operlogcontroller.NewV1())
		})
	})
	return nil
}
