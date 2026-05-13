// Package backend wires the org-center source plugin into the host plugin registry.
package backend

import (
	"context"

	hostorgcap "lina-core/pkg/orgcap"
	"lina-core/pkg/pluginhost"
	orgcenter "lina-plugin-org-center"
	deptcontroller "lina-plugin-org-center/backend/internal/controller/dept"
	postcontroller "lina-plugin-org-center/backend/internal/controller/post"
	deptsvc "lina-plugin-org-center/backend/internal/service/dept"
	postsvc "lina-plugin-org-center/backend/internal/service/post"
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
	plugin.Assets().UseEmbeddedFiles(orgcenter.EmbeddedFiles)
	plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	)
	pluginhost.RegisterSourcePlugin(plugin)
}

// registerRoutes binds department and post management routes through the published host middleware set.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	hostServices := registrar.HostServices()
	if hostServices == nil || hostServices.I18n() == nil || hostServices.TenantFilter() == nil {
		panic("org-center routes require host i18n and tenant-filter services")
	}
	hostorgcap.RegisterProvider(orgcapadapter.New(hostServices.TenantFilter()))
	deptSvc := deptsvc.New(hostServices.TenantFilter())
	postSvc := postsvc.New(hostServices.I18n(), hostServices.TenantFilter())
	routes := registrar.Routes()
	middlewares := routes.Middlewares()
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
			group.Bind(
				deptcontroller.NewV1(deptSvc),
				postcontroller.NewV1(postSvc),
			)
		})
	})
	return nil
}
