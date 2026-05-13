// Package backend wires the content-notice source plugin into the host plugin registry.
package backend

import (
	"context"

	"lina-core/pkg/pluginhost"
	contentnotice "lina-plugin-content-notice"
	noticecontroller "lina-plugin-content-notice/backend/internal/controller/notice"
	noticesvc "lina-plugin-content-notice/backend/internal/service/notice"
)

// content-notice plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "content-notice"
)

// init registers the content-notice source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(contentnotice.EmbeddedFiles)
	plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	)
	pluginhost.RegisterSourcePlugin(plugin)
}

// registerRoutes binds notice-management routes through the published host middleware set.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	routes := registrar.Routes()
	middlewares := routes.Middlewares()
	hostServices := registrar.HostServices()
	if hostServices == nil ||
		hostServices.BizCtx() == nil ||
		hostServices.Notify() == nil ||
		hostServices.TenantFilter() == nil {
		panic("content-notice routes require host bizctx, notify, and tenant-filter services")
	}
	noticeSvc := noticesvc.New(hostServices.BizCtx(), hostServices.Notify(), hostServices.TenantFilter())
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
			group.Bind(noticecontroller.NewV1(noticeSvc))
		})
	})
	return nil
}
