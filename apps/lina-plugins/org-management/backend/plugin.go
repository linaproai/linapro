// Package backend wires the org-management source plugin into the host plugin registry.
package backend

import (
	"context"

	deptcontroller "lina-core/pkg/plugincontroller/dept"
	postcontroller "lina-core/pkg/plugincontroller/post"
	"lina-core/pkg/pluginhost"
	orgmanagement "lina-plugin-org-management"
)

// org-management plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "org-management"
)

// init registers the org-management source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.UseEmbeddedFiles(orgmanagement.EmbeddedFiles)
	plugin.RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	)
	pluginhost.RegisterSourcePlugin(plugin)
}

// registerRoutes binds department and post management routes through the published host middleware set.
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
				deptcontroller.NewV1(),
				postcontroller.NewV1(),
			)
		})
	})
	return nil
}
