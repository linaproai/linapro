// Package backend wires the monitor-loginlog source plugin into the host plugin registry.
package backend

import (
	"context"

	"lina-core/pkg/pluginhost"
	monitorloginlogplugin "lina-plugin-monitor-loginlog"
	loginlogcontroller "lina-plugin-monitor-loginlog/backend/internal/controller/loginlog"
	loginlogsvc "lina-plugin-monitor-loginlog/backend/service/loginlog"
)

// monitor-loginlog plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "monitor-loginlog"
)

// init registers the monitor-loginlog source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.UseEmbeddedFiles(monitorloginlogplugin.EmbeddedFiles)
	plugin.RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	)
	plugin.RegisterHook(
		pluginhost.ExtensionPointAuthLoginSucceeded,
		pluginhost.CallbackExecutionModeAsync,
		handleAuthEvent,
	)
	plugin.RegisterHook(
		pluginhost.ExtensionPointAuthLoginFailed,
		pluginhost.CallbackExecutionModeAsync,
		handleAuthEvent,
	)
	plugin.RegisterHook(
		pluginhost.ExtensionPointAuthLogoutSucceeded,
		pluginhost.CallbackExecutionModeAsync,
		handleAuthEvent,
	)
	pluginhost.RegisterSourcePlugin(plugin)
}

// registerRoutes binds login-log governance routes through the published host middleware set.
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
			group.Bind(loginlogcontroller.NewV1())
		})
	})
	return nil
}

// handleAuthEvent persists one host authentication lifecycle event into the login-log table owned by this plugin.
func handleAuthEvent(ctx context.Context, payload pluginhost.HookPayload) error {
	values := payload.Values()
	status, _ := pluginhost.HookPayloadIntValue(values, pluginhost.HookPayloadKeyStatus)

	return loginlogsvc.New().Create(ctx, loginlogsvc.CreateInput{
		UserName: pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyUserName),
		Status:   status,
		Ip:       pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyIP),
		Browser:  pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyBrowser),
		Os:       pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyOS),
		Msg:      pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyMessage),
	})
}
