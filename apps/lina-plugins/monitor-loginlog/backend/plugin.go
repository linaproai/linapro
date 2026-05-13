// Package backend wires the monitor-loginlog source plugin into the host plugin registry.
package backend

import (
	"context"

	"lina-core/pkg/pluginhost"
	monitorloginlogplugin "lina-plugin-monitor-loginlog"
	loginlogcontroller "lina-plugin-monitor-loginlog/backend/internal/controller/loginlog"
	loginlogsvc "lina-plugin-monitor-loginlog/backend/internal/service/loginlog"
)

// monitor-loginlog plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "monitor-loginlog"
)

// init registers the monitor-loginlog source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(monitorloginlogplugin.EmbeddedFiles)
	plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	)
	plugin.Hooks().RegisterHook(
		pluginhost.ExtensionPointAuthLoginSucceeded,
		pluginhost.CallbackExecutionModeAsync,
		handleAuthEvent,
	)
	plugin.Hooks().RegisterHook(
		pluginhost.ExtensionPointAuthLoginFailed,
		pluginhost.CallbackExecutionModeAsync,
		handleAuthEvent,
	)
	plugin.Hooks().RegisterHook(
		pluginhost.ExtensionPointAuthLogoutSucceeded,
		pluginhost.CallbackExecutionModeAsync,
		handleAuthEvent,
	)
	pluginhost.RegisterSourcePlugin(plugin)
}

// registerRoutes binds login-log governance routes through the published host middleware set.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	hostServices := registrar.HostServices()
	if hostServices == nil || hostServices.I18n() == nil || hostServices.TenantFilter() == nil {
		panic("monitor-loginlog routes require host i18n and tenant-filter services")
	}
	loginLogSvc := loginlogsvc.New(hostServices.I18n(), hostServices.TenantFilter())
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
			group.Bind(loginlogcontroller.NewV1(loginLogSvc))
		})
	})
	return nil
}

// handleAuthEvent persists one host authentication lifecycle event into the login-log table owned by this plugin.
func handleAuthEvent(ctx context.Context, payload pluginhost.HookPayload) error {
	hostServices := payload.HostServices()
	if hostServices == nil || hostServices.I18n() == nil || hostServices.TenantFilter() == nil {
		panic("monitor-loginlog hook requires host i18n and tenant-filter services")
	}
	values := payload.Values()
	status, _ := pluginhost.HookPayloadIntValue(values, pluginhost.HookPayloadKeyStatus)
	message := pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyReason)
	if message == "" {
		message = pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyMessage)
	}

	return loginlogsvc.New(hostServices.I18n(), hostServices.TenantFilter()).Create(ctx, loginlogsvc.CreateInput{
		UserName: pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyUserName),
		Status:   status,
		Ip:       pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyIP),
		Browser:  pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyBrowser),
		Os:       pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyOS),
		Msg:      message,
	})
}
