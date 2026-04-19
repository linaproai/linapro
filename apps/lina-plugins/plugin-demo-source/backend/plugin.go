// Package backend wires the source demo plugin into the host plugin registry.
package backend

import (
	"context"
	"encoding/json"

	"lina-core/pkg/pluginhost"
	plugindemosource "lina-plugin-demo-source"
	democtrl "lina-plugin-demo-source/backend/internal/controller/demo"
	demosvc "lina-plugin-demo-source/backend/service/demo"
)

// Source demo plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded demo plugin.
	pluginID = "plugin-demo-source"
)

// init registers the embedded source demo plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.UseEmbeddedFiles(plugindemosource.EmbeddedFiles)
	plugin.RegisterUninstallHandler(func(ctx context.Context, input pluginhost.SourcePluginUninstallInput) error {
		if !input.PurgeStorageData() {
			return nil
		}
		return demosvc.New().PurgeStorageData(ctx)
	})
	plugin.RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	)
	plugin.RegisterJobHandler(pluginhost.JobHandlerRegistration{
		Name:         "echo",
		DisplayName:  "源码插件示例回显",
		Description:  "返回源码插件示例任务参数，用于验证插件处理器生命周期与定时任务调度链路",
		ParamsSchema: `{"type":"object","properties":{"message":{"type":"string","description":"回显消息"}},"required":["message"]}`,
		Handler:      invokeEchoJobHandler,
	})
	pluginhost.RegisterSourcePlugin(plugin)
}

// registerRoutes binds the demo plugin HTTP routes using the published host
// middleware directory so plugin traffic follows the same governance chain as
// host-owned APIs.
func registerRoutes(ctx context.Context, registrar pluginhost.RouteRegistrar) error {
	var (
		middlewares    = registrar.Middlewares()
		demoController = democtrl.NewControllerV1()
	)
	registrar.Group("/api/v1", func(group pluginhost.RouteGroup) {
		group.Middleware(
			middlewares.NeverDoneCtx(),
			middlewares.HandlerResponse(),
			middlewares.CORS(),
			middlewares.RequestBodyLimit(),
			middlewares.Ctx(),
		)

		group.Group("/", func(group pluginhost.RouteGroup) {
			group.Bind(demoController.Ping)
		})

		group.Group("/", func(group pluginhost.RouteGroup) {
			group.Middleware(
				middlewares.Auth(),
				middlewares.OperLog(),
				middlewares.Permission(),
			)
			group.Bind(
				demoController.Summary,
				demoController.ListRecords,
				demoController.GetRecord,
				demoController.CreateRecord,
				demoController.UpdateRecord,
				demoController.DeleteRecord,
				demoController.DownloadAttachment,
			)
		})
	})
	return nil
}

// echoJobHandlerParams stores the supported payload for the demo source-plugin job handler.
type echoJobHandlerParams struct {
	Message string `json:"message"`
}

// invokeEchoJobHandler returns the plugin identity and payload so tests can
// verify plugin handler registration and execution end to end.
func invokeEchoJobHandler(ctx context.Context, params json.RawMessage) (any, error) {
	var input echoJobHandlerParams
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, err
	}
	return map[string]any{
		"pluginId": pluginID,
		"message":  input.Message,
	}, nil
}
