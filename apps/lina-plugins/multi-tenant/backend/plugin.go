// Package backend wires the multi-tenant source plugin into the host plugin registry.
package backend

import (
	"context"

	"lina-core/pkg/pluginhost"
	pkgtenantcap "lina-core/pkg/tenantcap"
	multitenant "lina-plugin-multi-tenant"
	authcontroller "lina-plugin-multi-tenant/backend/internal/controller/auth"
	platformcontroller "lina-plugin-multi-tenant/backend/internal/controller/platform"
	tenantcontroller "lina-plugin-multi-tenant/backend/internal/controller/tenant"
	"lina-plugin-multi-tenant/backend/internal/service/lifecycleguard"
	"lina-plugin-multi-tenant/backend/internal/service/provider"
)

// multi-tenant plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "multi-tenant"
)

// init registers the multi-tenant source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(multitenant.EmbeddedFiles)
	pkgtenantcap.RegisterProvider(provider.New())
	pluginhost.RegisterLifecycleGuard(pluginID, lifecycleguard.New())
	plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	)
	pluginhost.RegisterSourcePlugin(plugin)
}

// registerRoutes binds multi-tenant routes through the published host middleware set.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	routes := registrar.Routes()
	middlewares := routes.Middlewares()
	routes.Group("/api/v1", func(group pluginhost.RouteGroup) {
		authCtrl := authcontroller.NewControllerV1()
		group.Middleware(
			middlewares.NeverDoneCtx(),
			middlewares.HandlerResponse(),
			middlewares.CORS(),
			middlewares.RequestBodyLimit(),
			middlewares.Ctx(),
		)
		group.Group("/", func(group pluginhost.RouteGroup) {
			group.Bind(
				authCtrl.SelectTenant,
			)
		})
		group.Group("/", func(group pluginhost.RouteGroup) {
			group.Middleware(
				middlewares.Auth(),
				middlewares.Tenancy(),
				middlewares.Permission(),
			)
			group.Bind(
				authCtrl.LoginTenants,
				authCtrl.SwitchTenant,
				platformcontroller.NewV1(),
				tenantcontroller.NewV1(),
			)
		})
	})
	return nil
}
