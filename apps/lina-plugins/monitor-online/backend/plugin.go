// Package backend wires the monitor-online source plugin into the host plugin registry.
package backend

import (
	"context"

	onlinecontroller "lina-core/pkg/plugincontroller/monitoronline"
	"lina-core/pkg/pluginhost"
	monitoronlineplugin "lina-plugin-monitor-online"
)

// monitor-online plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "monitor-online"
)

// init registers the monitor-online source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.UseEmbeddedFiles(monitoronlineplugin.EmbeddedFiles)
	plugin.RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	)
	pluginhost.RegisterSourcePlugin(plugin)
}

// registerRoutes binds online-user governance routes through the published host middleware set.
func registerRoutes(ctx context.Context, registrar pluginhost.RouteRegistrar) error {
	middlewares := registrar.Middlewares()
	registrar.Group("/api/v1", func(group pluginhost.RouteGroup) {
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
				middlewares.OperLog(),
				middlewares.Permission(),
			)
			group.Bind(
				onlinecontroller.OnlineList(),
				onlinecontroller.OnlineForceLogout(),
			)
		})
	})
	return nil
}
