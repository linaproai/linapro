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
	"lina-plugin-multi-tenant/backend/internal/service/impersonate"
	"lina-plugin-multi-tenant/backend/internal/service/lifecycleguard"
	"lina-plugin-multi-tenant/backend/internal/service/membership"
	"lina-plugin-multi-tenant/backend/internal/service/provider"
	"lina-plugin-multi-tenant/backend/internal/service/resolver"
	"lina-plugin-multi-tenant/backend/internal/service/resolverconfig"
	tenantsvc "lina-plugin-multi-tenant/backend/internal/service/tenant"
	"lina-plugin-multi-tenant/backend/internal/service/tenantplugin"
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
	plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	)
	pluginhost.RegisterSourcePlugin(plugin)
}

// registerRoutes binds multi-tenant routes through the published host middleware set.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	hostServices := registrar.HostServices()
	if hostServices == nil ||
		hostServices.Auth() == nil ||
		hostServices.BizCtx() == nil ||
		hostServices.Config() == nil {
		panic("multi-tenant routes require host auth, bizctx, and config services")
	}
	var (
		membershipSvc     = membership.New(hostServices.BizCtx())
		tenantPluginSvc   = tenantplugin.New(hostServices.BizCtx())
		resolverConfigSvc = resolverconfig.New()
		tenantSvc         = tenantsvc.New(hostServices.BizCtx(), resolverConfigSvc, tenantPluginSvc)
		resolverSvc       = resolver.New(hostServices.BizCtx(), membershipSvc)
		providerSvc       = provider.New(membershipSvc, resolverSvc, resolverConfigSvc)
	)
	pkgtenantcap.RegisterProvider(providerSvc)
	pluginhost.RegisterLifecycleGuard(pluginID, lifecycleguard.New(tenantSvc))
	var (
		impersonateSvc = impersonate.New(hostServices.BizCtx(), hostServices.Config(), tenantSvc)
		routes         = registrar.Routes()
		middlewares    = routes.Middlewares()
	)
	routes.Group("/api/v1", func(group pluginhost.RouteGroup) {
		authCtrl := authcontroller.NewV1(hostServices.Auth(), membershipSvc, providerSvc)
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
				platformcontroller.NewV1(tenantSvc, impersonateSvc),
				tenantcontroller.NewV1(tenantPluginSvc),
			)
		})
	})
	return nil
}
