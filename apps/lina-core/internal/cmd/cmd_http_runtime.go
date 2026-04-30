// This file maintains HTTP runtime services and process-level server settings.

package cmd

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/cluster"
	"lina-core/internal/service/config"
	"lina-core/internal/service/cron"
	jobhandlersvc "lina-core/internal/service/jobhandler"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
	"lina-core/internal/service/middleware"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/logger"
)

// httpRuntime groups long-lived services that must be shared across HTTP
// startup phases without re-constructing them in each route binding helper.
type httpRuntime struct {
	configSvc     config.Service                 // configSvc reads static and runtime host settings shared by startup helpers.
	clusterSvc    cluster.Service                // clusterSvc owns primary-election lifecycle for clustered deployments.
	pluginSvc     pluginsvc.Service              // pluginSvc owns plugin lifecycle, runtime assets, routes, and hooks.
	apiDocSvc     apidoc.Service                 // apiDocSvc builds the host-managed OpenAPI document.
	jobRegistry   jobhandlersvc.Registry         // jobRegistry stores host and plugin scheduled-job handlers.
	jobMgmtSvc    jobmgmtsvc.Service             // jobMgmtSvc backs scheduled-job management controllers and cron projection.
	middlewareSvc middleware.Service             // middlewareSvc publishes host middleware chains for static and plugin routes.
	cronSvc       cron.Service                   // cronSvc starts host-level and persistent scheduled jobs.
	serverCfg     *config.ServerExtensionsConfig // serverCfg contains host extension route settings such as API docs.
}

// configureHTTPServer applies process-level server configuration that must be
// in place before any route groups are bound.
func configureHTTPServer(
	ctx context.Context,
	server *ghttp.Server,
	configSvc config.Service,
) error {
	loggerCfg := configSvc.GetLogger(ctx)
	if err := logger.BindServer(server, logger.ServerOutputConfig{
		Path:   loggerCfg.Path,
		File:   loggerCfg.File,
		Stdout: loggerCfg.Stdout,
	}); err != nil {
		return err
	}

	shutdownCfg := configSvc.GetShutdown(ctx)
	if shutdownCfg != nil && shutdownCfg.Timeout > 0 {
		timeoutSeconds := durationSeconds(shutdownCfg.Timeout)
		server.SetGracefulTimeout(timeoutSeconds)
		server.SetGracefulShutdownTimeout(timeoutSeconds)
	}

	// Request-size limits are enforced by host middleware so multipart uploads
	// can follow the runtime-effective sys.upload.maxSize value per request
	// instead of being clipped by GoFrame's static 8MB default at server entry.
	server.SetClientMaxBodySize(0)
	return nil
}

// newHTTPRuntime constructs the shared services used by the HTTP server and
// keeps their startup dependencies in one place.
func newHTTPRuntime(ctx context.Context, configSvc config.Service) (*httpRuntime, error) {
	var (
		clusterSvc    = cluster.New(configSvc.GetCluster(ctx))
		pluginSvc     = pluginsvc.New(clusterSvc)
		apiDocSvc     = apidoc.New(configSvc, pluginSvc)
		jobRegistry   = jobhandlersvc.New()
		jobScheduler  = jobmgmtsvc.NewScheduler(clusterSvc, jobRegistry, configSvc)
		jobMgmtSvc    = jobmgmtsvc.New(configSvc, jobRegistry, jobScheduler)
		middlewareSvc = middleware.New()
	)

	// Host-owned handler definitions are registered before cron startup so the
	// persistent scheduler can project and validate code-owned jobs immediately.
	if err := jobhandlersvc.RegisterHostHandlers(jobRegistry, jobMgmtSvc); err != nil {
		return nil, err
	}

	sessionCfg, err := configSvc.GetSession(ctx)
	if err != nil {
		return nil, err
	}

	return &httpRuntime{
		configSvc:     configSvc,
		clusterSvc:    clusterSvc,
		pluginSvc:     pluginSvc,
		apiDocSvc:     apiDocSvc,
		jobRegistry:   jobRegistry,
		jobMgmtSvc:    jobMgmtSvc,
		middlewareSvc: middlewareSvc,
		cronSvc: cron.New(
			sessionCfg,
			middlewareSvc.SessionStore(),
			clusterSvc,
			jobRegistry,
			jobMgmtSvc,
			jobScheduler,
		),
		serverCfg: configSvc.GetServerExtensions(ctx),
	}, nil
}

// startHTTPRuntime starts cluster, plugin, and cron services in the order
// required for source-plugin handlers and dynamic runtime state to be ready.
func startHTTPRuntime(ctx context.Context, runtime *httpRuntime) error {
	runtime.clusterSvc.Start(ctx)

	// Auto-enable and source-upgrade validation run before plugin routes and
	// cron jobs are registered so stale plugin state fails the process early.
	if err := runtime.pluginSvc.BootstrapAutoEnable(ctx); err != nil {
		return err
	}
	if err := runtime.pluginSvc.ValidateSourcePluginUpgradeReadiness(ctx); err != nil {
		return err
	}
	if _, err := jobhandlersvc.AttachPluginLifecycle(
		ctx,
		runtime.jobRegistry,
		runtime.pluginSvc,
	); err != nil {
		return err
	}

	// Cron startup comes after plugin lifecycle wiring so plugin-owned scheduled
	// jobs are visible when the persistent scheduler loads enabled jobs.
	runtime.cronSvc.Start(ctx)
	return nil
}

// shutdownHTTPRuntime stops non-HTTP runtime components after GoFrame Server.Run
// has handled signal listening and HTTP graceful shutdown.
func shutdownHTTPRuntime(ctx context.Context, runtime *httpRuntime, configSvc config.Service) error {
	shutdownBaseCtx := context.WithoutCancel(ctx)
	shutdownTimeout := resolveShutdownTimeout(shutdownBaseCtx, configSvc)
	logger.Infof(shutdownBaseCtx, "runtime shutdown requested, timeout=%s", shutdownTimeout)

	shutdownCtx, cancel := context.WithTimeout(shutdownBaseCtx, shutdownTimeout)
	defer cancel()

	if runtime != nil && runtime.cronSvc != nil {
		if err := shutdownStep(shutdownCtx, "cron scheduler", func(stepCtx context.Context) error {
			runtime.cronSvc.Stop(stepCtx)
			return nil
		}); err != nil {
			logger.Warningf(shutdownBaseCtx, "runtime shutdown failed: %v", err)
			return err
		}
	}

	if runtime != nil && runtime.clusterSvc != nil {
		if err := shutdownStep(shutdownCtx, "cluster service", func(stepCtx context.Context) error {
			runtime.clusterSvc.Stop(stepCtx)
			return nil
		}); err != nil {
			logger.Warningf(shutdownBaseCtx, "runtime shutdown failed: %v", err)
			return err
		}
	}

	if err := shutdownStep(shutdownCtx, "database pool", func(stepCtx context.Context) error {
		return g.DB().Close(stepCtx)
	}); err != nil {
		logger.Warningf(shutdownBaseCtx, "runtime shutdown failed: %v", err)
		return err
	}

	logger.Info(shutdownBaseCtx, "runtime shutdown completed")
	return nil
}

// shutdownStep runs one shutdown operation under the shared deadline and
// returns a step-scoped error when it fails or times out.
func shutdownStep(ctx context.Context, name string, fn func(context.Context) error) error {
	if err := ctx.Err(); err != nil {
		return gerror.Wrapf(err, "%s shutdown skipped because the shutdown deadline is done", name)
	}

	done := make(chan error, 1)
	go func() {
		done <- fn(ctx)
	}()

	select {
	case err := <-done:
		if err != nil {
			return gerror.Wrapf(err, "%s shutdown failed", name)
		}
		return nil
	case <-ctx.Done():
		return gerror.Wrapf(ctx.Err(), "%s shutdown timed out", name)
	}
}

// resolveShutdownTimeout returns the configured full runtime-shutdown budget.
func resolveShutdownTimeout(ctx context.Context, configSvc config.Service) time.Duration {
	if configSvc == nil {
		return 30 * time.Second
	}
	cfg := configSvc.GetShutdown(ctx)
	if cfg == nil || cfg.Timeout <= 0 {
		return 30 * time.Second
	}
	return cfg.Timeout
}

// durationSeconds converts a validated duration into whole seconds for
// GoFrame server configuration.
func durationSeconds(value time.Duration) int {
	seconds := int(value / time.Second)
	if seconds < 1 {
		return 1
	}
	return seconds
}
