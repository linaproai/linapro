package backend

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/pluginhost"
	plugindemosource "lina-plugin-demo-source"
	democtrl "lina-plugin-demo-source/backend/internal/controller/demo"
)

const (
	pluginID = "plugin-demo-source"
)

func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.UseEmbeddedFiles(plugindemosource.EmbeddedFiles)
	plugin.RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	)
	pluginhost.RegisterSourcePlugin(plugin)
}

func registerRoutes(ctx context.Context, registrar pluginhost.RouteRegistrar) error {
	var (
		middlewares    = registrar.Middlewares()
		demoController = democtrl.NewV1()
	)
	registrar.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.Middleware(
			middlewares.NeverDoneCtx(),
			middlewares.HandlerResponse(),
			middlewares.CORS(),
			middlewares.Ctx(),
		)

		group.Group("/", func(group *ghttp.RouterGroup) {
			group.Bind(demoController.Ping)
		})

		group.Group("/", func(group *ghttp.RouterGroup) {
			group.Middleware(
				middlewares.Auth(),
				middlewares.OperLog(),
				middlewares.Permission(),
			)
			group.Bind(demoController.Summary)
		})
	})
	return nil
}
