// Package backend wires the monitor-server source plugin into the host plugin registry.
package backend

import (
	"context"
	"time"

	"lina-core/pkg/pluginhost"
	monitorserverplugin "lina-plugin-monitor-server"
	servercontroller "lina-plugin-monitor-server/backend/internal/controller/monitor"
	monitorconfig "lina-plugin-monitor-server/backend/internal/service/config"
	monitorsvc "lina-plugin-monitor-server/backend/internal/service/monitor"
)

// monitor-server plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "monitor-server"
	// serviceMonitorCollectorName identifies the server metric collection cron.
	serviceMonitorCollectorName = "service-monitor-collector"
	// serviceMonitorCollectorDisplayName is the English source title for the collection cron.
	serviceMonitorCollectorDisplayName = "Server Monitor Collection"
	// serviceMonitorCollectorDescription is the English source description for the collection cron.
	serviceMonitorCollectorDescription = "Collects server runtime metrics for the monitor-server plugin."
	// serviceMonitorCleanupName identifies the server metric cleanup cron.
	serviceMonitorCleanupName = "service-monitor-cleanup"
	// serviceMonitorCleanupDisplayName is the English source title for the cleanup cron.
	serviceMonitorCleanupDisplayName = "Server Monitor Cleanup"
	// serviceMonitorCleanupDescription is the English source description for the cleanup cron.
	serviceMonitorCleanupDescription = "Cleans up expired server runtime metric snapshots for the monitor-server plugin."
)

// sharedMonitorSvc is the plugin-owned server monitor service shared by HTTP,
// Cron, and startup hooks so sampling state is not split across callbacks.
var sharedMonitorSvc = monitorsvc.New()

// init registers the monitor-server source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(monitorserverplugin.EmbeddedFiles)
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
	plugin.Hooks().RegisterHook(
		pluginhost.ExtensionPointSystemStarted,
		pluginhost.CallbackExecutionModeAsync,
		collectOnSystemStarted,
	)
	pluginhost.RegisterSourcePlugin(plugin)
}

// registerRoutes binds server-monitor query routes through the published host middleware set.
func registerRoutes(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
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
			group.Bind(servercontroller.NewV1(sharedMonitorSvc))
		})
	})
	return nil
}

// registerBuiltinCrons contributes managed cron definitions for server-monitor collection and cleanup.
func registerBuiltinCrons(ctx context.Context, registrar pluginhost.CronRegistrar) error {
	hostServices := registrar.HostServices()
	if hostServices == nil || hostServices.Config() == nil {
		panic("monitor-server cron requires host config service")
	}
	monitorCfg, err := monitorconfig.Load(ctx, hostServices.Config())
	if err != nil {
		return err
	}
	interval := monitorCfg.Interval

	if err := registrar.AddWithMetadata(
		ctx,
		"@every "+interval.String(),
		serviceMonitorCollectorName,
		serviceMonitorCollectorDisplayName,
		serviceMonitorCollectorDescription,
		func(ctx context.Context) error {
			return collectSnapshot(ctx, sharedMonitorSvc)
		},
	); err != nil {
		return err
	}
	return registrar.AddWithMetadata(
		ctx,
		"# * * * * *",
		serviceMonitorCleanupName,
		serviceMonitorCleanupDisplayName,
		serviceMonitorCleanupDescription,
		func(ctx context.Context) error {
			return cleanupSnapshots(ctx, registrar, sharedMonitorSvc)
		},
	)
}

// collectOnSystemStarted performs one eager collection after host startup so the page has an initial snapshot.
func collectOnSystemStarted(ctx context.Context, payload pluginhost.HookPayload) error {
	sharedMonitorSvc.CollectAndStore(ctx)
	return nil
}

// collectSnapshot writes one fresh monitoring snapshot.
func collectSnapshot(ctx context.Context, monitorSvc monitorsvc.Service) error {
	monitorSvc.CollectAndStore(ctx)
	return nil
}

// cleanupSnapshots removes expired monitoring snapshots.
func cleanupSnapshots(ctx context.Context, registrar pluginhost.CronRegistrar, monitorSvc monitorsvc.Service) error {
	if registrar != nil && !registrar.IsPrimaryNode() {
		return nil
	}

	hostServices := registrar.HostServices()
	if hostServices == nil || hostServices.Config() == nil {
		panic("monitor-server cleanup requires host config service")
	}
	monitorCfg, err := monitorconfig.Load(ctx, hostServices.Config())
	if err != nil {
		return err
	}

	_, err = monitorSvc.CleanupStale(ctx, monitorCfg.Interval*time.Duration(monitorCfg.RetentionMultiplier))
	return err
}
