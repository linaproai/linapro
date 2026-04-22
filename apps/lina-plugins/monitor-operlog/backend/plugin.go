// Package backend wires the monitor-operlog source plugin into the host plugin registry.
package backend

import (
	"context"

	"lina-core/pkg/pluginhost"
	monitoroperlogplugin "lina-plugin-monitor-operlog"
	operlogcontroller "lina-plugin-monitor-operlog/backend/internal/controller/operlog"
	operlogsvc "lina-plugin-monitor-operlog/backend/service/operlog"
)

// monitor-operlog plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "monitor-operlog"
)

// init registers the monitor-operlog source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.UseEmbeddedFiles(monitoroperlogplugin.EmbeddedFiles)
	plugin.RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	)
	plugin.RegisterHook(
		pluginhost.ExtensionPointAuditRecorded,
		pluginhost.CallbackExecutionModeAsync,
		handleAuditRecorded,
	)
	pluginhost.RegisterSourcePlugin(plugin)
}

// registerRoutes binds operation-log governance routes and audit middleware through the published host HTTP registrars.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	registrar.GlobalMiddlewares().Bind("/*", auditMiddleware)

	var (
		routes      = registrar.Routes()
		middlewares = routes.Middlewares()
	)
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
				middlewares.Permission(),
			)
			group.Bind(operlogcontroller.NewV1())
		})
	})
	return nil
}

// handleAuditRecorded persists one host audit event into the operation-log table owned by this plugin.
func handleAuditRecorded(ctx context.Context, payload pluginhost.HookPayload) error {
	values := payload.Values()

	operType, _ := pluginhost.HookPayloadIntValue(values, pluginhost.HookPayloadKeyOperType)
	status, _ := pluginhost.HookPayloadIntValue(values, pluginhost.HookPayloadKeyStatus)
	costTime, _ := pluginhost.HookPayloadIntValue(values, pluginhost.HookPayloadKeyCostTime)

	return operlogsvc.New().Create(ctx, operlogsvc.CreateInput{
		Title:         pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyTitle),
		OperSummary:   pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyOperSummary),
		OperType:      operType,
		Method:        pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyMethod),
		RequestMethod: pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyRequestMethod),
		OperName:      pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyOperName),
		OperUrl:       pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyOperURL),
		OperIp:        pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyIP),
		OperParam:     pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyOperParam),
		JsonResult:    pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyJSONResult),
		Status:        status,
		ErrorMsg:      pluginhost.HookPayloadStringValue(values, pluginhost.HookPayloadKeyErrorMsg),
		CostTime:      costTime,
	})
}
