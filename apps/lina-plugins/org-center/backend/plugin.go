// Package backend wires the org-center source plugin into the host plugin registry.
package backend

import (
	"context"

	hostorgcap "lina-core/pkg/orgcap"
	"lina-core/pkg/pluginhost"
	orgcenter "lina-plugin-org-center"
	deptcontroller "lina-plugin-org-center/backend/internal/controller/dept"
	postcontroller "lina-plugin-org-center/backend/internal/controller/post"
	"lina-plugin-org-center/backend/provider/orgcapadapter"
)

// org-center plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "org-center"
)

// init registers the org-center source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.UseEmbeddedFiles(orgcenter.EmbeddedFiles)
	hostorgcap.RegisterProvider(orgcapadapter.New())
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
