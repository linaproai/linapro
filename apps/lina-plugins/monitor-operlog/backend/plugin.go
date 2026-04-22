// Package backend wires the monitor-operlog source plugin into the host plugin registry.
package backend

import (
	"context"

	"lina-core/pkg/audittype"
	"lina-core/pkg/pluginhost"
	monitoroperlogplugin "lina-plugin-monitor-operlog"
	operlogcontroller "lina-plugin-monitor-operlog/backend/internal/controller/operlog"
	middlewaresvc "lina-plugin-monitor-operlog/backend/internal/service/middleware"
	operlogsvc "lina-plugin-monitor-operlog/backend/internal/service/operlog"
)

// monitor-operlog plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "monitor-operlog"
)

// init registers the monitor-operlog source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(monitoroperlogplugin.EmbeddedFiles)
	plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	)
	plugin.Hooks().RegisterHook(
		pluginhost.ExtensionPointAuditRecorded,
		pluginhost.CallbackExecutionModeAsync,
		handleAuditRecorded,
	)
	pluginhost.RegisterSourcePlugin(plugin)
}

// registerRoutes binds operation-log governance routes and audit middleware through the published host HTTP registrars.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	auditMiddlewareSvc := middlewaresvc.New()
	registrar.GlobalMiddlewares().Bind("/*", auditMiddlewareSvc.Audit)

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
	var (
		values       = payload.Values()
		operType, ok = pluginhost.HookPayloadOperTypeValue(values, pluginhost.HookPayloadKeyOperType)
		status, _    = pluginhost.HookPayloadIntValue(values, pluginhost.HookPayloadKeyStatus)
		costTime, _  = pluginhost.HookPayloadIntValue(values, pluginhost.HookPayloadKeyCostTime)
	)
	if !ok {
		operType = audittype.OperTypeOther
	}
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
