// Package backend wires the monitor-server source plugin into the host plugin registry.
package backend

import (
	"context"
	"time"

	"lina-core/pkg/pluginhost"
	configsvc "lina-core/pkg/pluginservice/config"
	monitorserverplugin "lina-plugin-monitor-server"
	servercontroller "lina-plugin-monitor-server/backend/internal/controller/monitor"
	monitorsvc "lina-plugin-monitor-server/backend/service/monitor"
)

// monitor-server plugin constants.
const (
	// pluginID is the immutable identifier published by the embedded source plugin.
	pluginID = "monitor-server"
)

// init registers the monitor-server source plugin and its host callbacks.
func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.UseEmbeddedFiles(monitorserverplugin.EmbeddedFiles)
	plugin.RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerRoutes,
	)
	plugin.RegisterCron(
		pluginhost.ExtensionPointCronRegister,
		pluginhost.CallbackExecutionModeBlocking,
		registerBuiltinCrons,
	)
	plugin.RegisterHook(
		pluginhost.ExtensionPointSystemStarted,
		pluginhost.CallbackExecutionModeAsync,
		collectOnSystemStarted,
	)
	pluginhost.RegisterSourcePlugin(plugin)
}

// registerRoutes binds server-monitor query routes through the published host middleware set.
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
			group.Bind(servercontroller.NewV1())
		})
	})
	return nil
}

// registerBuiltinCrons contributes managed cron definitions for server-monitor collection and cleanup.
func registerBuiltinCrons(ctx context.Context, registrar pluginhost.CronRegistrar) error {
	monitorCfg := configsvc.New().GetMonitor(ctx)
	interval := monitorCfg.Interval
	if interval <= 0 {
		interval = 30 * time.Second
	}

	if err := registrar.Add(ctx, "@every "+interval.String(), "服务监控采集", collectSnapshot); err != nil {
		return err
	}
	return registrar.Add(ctx, "# * * * * *", "服务监控清理", func(ctx context.Context) error {
		return cleanupSnapshots(ctx, registrar)
	})
}

// collectOnSystemStarted performs one eager collection after host startup so the page has an initial snapshot.
func collectOnSystemStarted(ctx context.Context, payload pluginhost.HookPayload) error {
	monitorsvc.New().CollectAndStore(ctx)
	return nil
}

// collectSnapshot writes one fresh monitoring snapshot.
func collectSnapshot(ctx context.Context) error {
	monitorsvc.New().CollectAndStore(ctx)
	return nil
}

// cleanupSnapshots removes expired monitoring snapshots.
func cleanupSnapshots(ctx context.Context, registrar pluginhost.CronRegistrar) error {
	if registrar != nil && !registrar.IsPrimaryNode() {
		return nil
	}

	monitorCfg := configsvc.New().GetMonitor(ctx)
	interval := monitorCfg.Interval
	if interval <= 0 {
		interval = 30 * time.Second
	}

	retentionMultiplier := monitorCfg.RetentionMultiplier
	if retentionMultiplier <= 0 {
		retentionMultiplier = 120
	}

	_, err := monitorsvc.New().CleanupStale(ctx, interval*time.Duration(retentionMultiplier))
	return err
}
