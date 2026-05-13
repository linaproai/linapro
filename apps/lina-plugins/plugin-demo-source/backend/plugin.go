// Package backend wires the source demo plugin into the host plugin registry.
package backend

import (
	"context"

	"lina-core/pkg/pluginhost"
	plugindemosource "lina-plugin-demo-source"
	democtrl "lina-plugin-demo-source/backend/internal/controller/demo"
	demosvc "lina-plugin-demo-source/backend/internal/service/demo"
)

// Source demo plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded demo plugin.
	pluginID = "plugin-demo-source"
	// sourcePluginEchoInspectionName identifies the demo source-plugin cron.
	sourcePluginEchoInspectionName = "source-plugin-echo-inspection"
	// sourcePluginEchoInspectionDisplayName is the English source title for the demo cron.
	sourcePluginEchoInspectionDisplayName = "Source Plugin Echo Inspection"
	// sourcePluginEchoInspectionDescription is the English source description for the demo cron.
	sourcePluginEchoInspectionDescription = "Runs a lightweight source-plugin inspection task for scheduler integration validation."
)

// init registers the embedded source demo plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(plugindemosource.EmbeddedFiles)
	plugin.Lifecycle().RegisterUninstallHandler(func(ctx context.Context, input pluginhost.SourcePluginUninstallInput) error {
		if !input.PurgeStorageData() {
			return nil
		}
		return demosvc.PurgeStorageData(ctx)
	})
	plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	)
	plugin.Cron().RegisterCron(
		pluginhost.ExtensionPointCronRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerBuiltinCrons,
	)
	pluginhost.RegisterSourcePlugin(plugin)
}

// registerRoutes binds the demo plugin HTTP routes using the published host
// middleware directory so plugin traffic follows the same governance chain as
// host-owned APIs.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
	hostServices := registrar.HostServices()
	if hostServices == nil || hostServices.I18n() == nil || hostServices.TenantFilter() == nil {
		panic("plugin-demo-source routes require host i18n and tenant-filter services")
	}
	demoSvc := demosvc.New(hostServices.I18n(), hostServices.TenantFilter())
	var (
		routes         = registrar.Routes()
		middlewares    = routes.Middlewares()
		demoController = democtrl.NewV1(demoSvc)
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
			group.Bind(demoController.Ping)
		})

		group.Group("/", func(group pluginhost.RouteGroup) {
			group.Middleware(
				middlewares.Auth(),
				middlewares.Tenancy(),
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

// registerBuiltinCrons contributes one plugin-owned builtin scheduled job so
// the host can project source-plugin cron registrations into unified
// scheduled-job management.
func registerBuiltinCrons(ctx context.Context, registrar pluginhost.CronRegistrar) error {
	return registrar.AddWithMetadata(
		ctx,
		"# */15 * * * *",
		sourcePluginEchoInspectionName,
		sourcePluginEchoInspectionDisplayName,
		sourcePluginEchoInspectionDescription,
		func(ctx context.Context) error {
			return nil
		},
	)
}
