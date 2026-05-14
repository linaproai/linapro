// This file maintains HTTP runtime services and process-level server settings.

package cmd

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/cluster"
	"lina-core/internal/service/config"
	"lina-core/internal/service/coordination"
	"lina-core/internal/service/cron"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/dict"
	"lina-core/internal/service/file"
	"lina-core/internal/service/hostlock"
	i18nsvc "lina-core/internal/service/i18n"
	jobhandlersvc "lina-core/internal/service/jobhandler"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/locker"
	"lina-core/internal/service/menu"
	"lina-core/internal/service/middleware"
	"lina-core/internal/service/notify"
	"lina-core/internal/service/orgcap"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/pluginhostservices"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	"lina-core/internal/service/startupstats"
	"lina-core/internal/service/sysconfig"
	sysinfosvc "lina-core/internal/service/sysinfo"
	tenantcapsvc "lina-core/internal/service/tenantcap"
	"lina-core/internal/service/user"
	"lina-core/internal/service/usermsg"
	"lina-core/pkg/dialect"
	"lina-core/pkg/logger"
	"lina-core/pkg/pluginhost"
)

// httpRuntime groups long-lived services that must be shared across HTTP
// startup phases without re-constructing them in each route binding helper.
type httpRuntime struct {
	configSvc       config.Service       // configSvc reads static and runtime host settings shared by startup helpers.
	coordinationSvc coordination.Service // coordinationSvc owns Redis-backed distributed coordination resources.
	clusterSvc      cluster.Service      // clusterSvc owns primary-election lifecycle for clustered deployments.
	pluginSvc       pluginsvc.Service    // pluginSvc owns plugin lifecycle, runtime assets, routes, and hooks.
	authSvc         auth.Service         // authSvc owns JWT, session, and token-state flows.
	authTokenIssuer auth.TenantTokenIssuer
	bizCtxSvc       bizctx.Service          // bizCtxSvc owns request-scoped business context mutation.
	i18nSvc         i18nsvc.Service         // i18nSvc owns runtime language bundles and localization.
	orgCapSvc       orgcap.Service          // orgCapSvc exposes optional organization capability.
	roleSvc         role.Service            // roleSvc owns permission and access snapshot state.
	sessionStore    session.Store           // sessionStore owns online-session persistence and hot state.
	tenantSvc       tenantcapsvc.Service    // tenantSvc exposes optional multi-tenant capability.
	kvCacheSvc      kvcache.Service         // kvCacheSvc owns runtime-selected KV backend.
	hostServices    pluginhost.HostServices // hostServices publishes runtime-owned adapters to source plugins.
	dictSvc         dict.Service            // dictSvc owns dictionary lookup and maintenance.
	fileSvc         file.Service            // fileSvc owns file metadata and storage operations.
	menuSvc         menu.Service            // menuSvc owns menu tree and permission menu lookup.
	notifySvc       notify.Service          // notifySvc owns unified notification delivery.
	sysConfigSvc    sysconfig.Service       // sysConfigSvc owns mutable runtime configuration records.
	sysInfoSvc      sysinfosvc.Service      // sysInfoSvc owns runtime diagnostics projection.
	userSvc         user.Service            // userSvc owns host user management operations.
	userMsgSvc      usermsg.Service         // userMsgSvc owns current-user inbox operations.
	apiDocSvc       apidoc.Service          // apiDocSvc builds the host-managed OpenAPI document.
	jobRegistry     jobhandlersvc.Registry
	jobMgmtSvc      jobmgmtsvc.Service
	middlewareSvc   middleware.Service
	cronSvc         cron.Service
	serverCfg       *config.ServerExtensionsConfig // serverCfg contains host extension route settings such as API docs.
}

// pluginStartupConsistencyValidator is the narrow startup contract required to
// fail fast before HTTP route publication.
type pluginStartupConsistencyValidator interface {
	// ValidateStartupConsistency verifies persisted plugin and tenant governance state.
	ValidateStartupConsistency(ctx context.Context) error
}

// newHTTPStartupContext creates the context shared by one HTTP startup
// orchestration pass. It carries only short-lived snapshots and statistics.
func newHTTPStartupContext(ctx context.Context, runtime *httpRuntime) (context.Context, *startupstats.Collector, error) {
	collector := startupstats.New()
	startupCtx := startupstats.WithCollector(ctx, collector)

	var err error
	if runtime != nil && runtime.pluginSvc != nil {
		startupCtx, err = runtime.pluginSvc.WithStartupDataSnapshot(startupCtx)
		if err != nil {
			return startupCtx, collector, err
		}
	}
	if runtime != nil && runtime.jobMgmtSvc != nil {
		startupCtx, err = runtime.jobMgmtSvc.WithStartupDataSnapshot(startupCtx)
		if err != nil {
			return startupCtx, collector, err
		}
	}
	return startupCtx, collector, nil
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
	link, err := currentDatabaseLink(ctx)
	if err != nil {
		return nil, err
	}
	dbDialect, err := dialect.From(link)
	if err != nil {
		return nil, err
	}
	if err = dbDialect.OnStartup(ctx, configSvc); err != nil {
		return nil, err
	}

	clusterCfg := configSvc.GetCluster(ctx)
	coordinationSvc, err := newHTTPCoordinationService(ctx, clusterCfg, configSvc)
	if err != nil {
		return nil, err
	}
	clusterSvc := cluster.NewWithCoordination(clusterCfg, coordinationSvc)
	if clusterCfg != nil && clusterCfg.Enabled {
		cachecoord.DefaultWithCoordination(clusterSvc, coordinationSvc)
		configureDistributedKVCache(coordinationSvc)
		locker.ConfigureCoordination(coordinationSvc)
		session.ConfigureCoordination(coordinationSvc)
	} else {
		configureLocalKVCache()
		locker.ConfigureCoordination(nil)
		session.ConfigureCoordination(nil)
	}

	var (
		bizCtxSvc     = bizctx.New()
		sessionStore  = session.NewDBStore()
		cacheCoordSvc = cachecoord.Default(clusterSvc)
		i18nSvc       = i18nsvc.New(bizCtxSvc, configSvc, cacheCoordSvc)
		pluginSvc     = pluginsvc.New(clusterSvc, configSvc, bizCtxSvc, cacheCoordSvc, i18nSvc, sessionStore)
		orgCapSvc     = orgcap.New(pluginSvc)
		tenantSvc     = tenantcapsvc.New(pluginSvc, bizCtxSvc)
		kvCacheSvc    = kvcache.New()
		roleSvc       = role.New(pluginSvc, bizCtxSvc, configSvc, i18nSvc, nil, orgCapSvc, tenantSvc)
		scopeSvc      = datascope.New(bizCtxSvc, roleSvc, orgCapSvc)
		dictSvc       = dict.New(i18nSvc)
		menuSvc       = menu.New(pluginSvc, i18nSvc, roleSvc)
		notifySvc     = notify.New(tenantSvc)
		authSvc       = auth.New(configSvc, pluginSvc, orgCapSvc, roleSvc, tenantSvc, sessionStore, kvCacheSvc)
		fileSvc       = file.New(configSvc, file.NewLocalStorage(configSvc.GetUploadPath(ctx)), bizCtxSvc, dictSvc, scopeSvc)
		sysConfigSvc  = sysconfig.New(configSvc, i18nSvc)
		sysInfoSvc    = sysinfosvc.New(configSvc, clusterSvc, coordinationSvc, cacheCoordSvc)
		userSvc       = user.New(authSvc, bizCtxSvc, i18nSvc, orgCapSvc, roleSvc, scopeSvc, tenantSvc)
		userMsgSvc    = usermsg.New(bizCtxSvc, notifySvc, i18nSvc)
		apiDocSvc     = apidoc.New(configSvc, bizCtxSvc, i18nSvc, pluginSvc)
		authTokenSvc  = authSvc.(auth.TenantTokenIssuer)
		jobRegistry   = jobhandlersvc.New()
		jobScheduler  = jobmgmtsvc.NewScheduler(clusterSvc, jobRegistry, configSvc)
		jobMgmtSvc    = jobmgmtsvc.New(bizCtxSvc, configSvc, i18nSvc, jobRegistry, jobScheduler, scopeSvc)
		middlewareSvc = middleware.New(authSvc, bizCtxSvc, configSvc, i18nSvc, pluginSvc, roleSvc, tenantSvc)
		hostServices  = pluginhostservices.New(
			apiDocSvc,
			authSvc,
			authTokenSvc,
			bizCtxSvc,
			scopeSvc,
			i18nSvc,
			pluginSvc,
			sessionStore,
			tenantSvc,
			notifySvc,
		)
	)
	roleSvc.SetDataScopeService(scopeSvc)
	pluginSvc.SetHostServices(hostServices)
	pluginSvc.SetTenantCapability(tenantSvc)
	pluginsvc.ConfigureWasmHostServices(
		kvCacheSvc,
		hostlock.New(locker.New()),
		notifySvc,
		configSvc,
		hostServices.Config(),
	)

	// Host-owned handler definitions are registered before cron startup so the
	// persistent scheduler can project and validate code-owned jobs immediately.
	if err := jobhandlersvc.RegisterHostHandlers(jobRegistry, jobMgmtSvc); err != nil {
		closeHTTPCoordinationAfterInitError(ctx, coordinationSvc)
		return nil, err
	}

	sessionCfg, err := configSvc.GetSession(ctx)
	if err != nil {
		closeHTTPCoordinationAfterInitError(ctx, coordinationSvc)
		return nil, err
	}

	return &httpRuntime{
		configSvc:       configSvc,
		coordinationSvc: coordinationSvc,
		clusterSvc:      clusterSvc,
		pluginSvc:       pluginSvc,
		authSvc:         authSvc,
		authTokenIssuer: authTokenSvc,
		bizCtxSvc:       bizCtxSvc,
		i18nSvc:         i18nSvc,
		orgCapSvc:       orgCapSvc,
		roleSvc:         roleSvc,
		sessionStore:    sessionStore,
		tenantSvc:       tenantSvc,
		kvCacheSvc:      kvCacheSvc,
		hostServices:    hostServices,
		dictSvc:         dictSvc,
		fileSvc:         fileSvc,
		menuSvc:         menuSvc,
		notifySvc:       notifySvc,
		sysConfigSvc:    sysConfigSvc,
		sysInfoSvc:      sysInfoSvc,
		userSvc:         userSvc,
		userMsgSvc:      userMsgSvc,
		apiDocSvc:       apiDocSvc,
		jobRegistry:     jobRegistry,
		jobMgmtSvc:      jobMgmtSvc,
		middlewareSvc:   middlewareSvc,
		cronSvc:         cron.New(configSvc, roleSvc, kvCacheSvc, pluginSvc, sessionCfg, sessionStore, clusterSvc, jobRegistry, jobMgmtSvc, jobScheduler),
		serverCfg:       configSvc.GetServerExtensions(ctx),
	}, nil
}

// configureDistributedKVCache switches process-default short-lived KV cache
// state to the shared coordination KV backend.
func configureDistributedKVCache(coordinationSvc coordination.Service) {
	kvcache.SetDefaultProvider(kvcache.NewCoordinationKVProvider(coordinationSvc))
}

// configureLocalKVCache restores the SQL table backend used by single-node
// deployments and tests.
func configureLocalKVCache() {
	kvcache.SetDefaultProvider(kvcache.NewSQLTableProvider())
}

// newHTTPCoordinationService creates the distributed coordination provider for
// cluster mode and intentionally returns nil in single-node deployments.
func newHTTPCoordinationService(
	ctx context.Context,
	clusterCfg *config.ClusterConfig,
	configSvc config.Service,
) (coordination.Service, error) {
	if clusterCfg == nil || !clusterCfg.Enabled {
		return nil, nil
	}
	if clusterCfg.Coordination != config.ClusterCoordinationRedis {
		return nil, gerror.Newf("cluster.coordination=%s is unsupported; only redis is supported", clusterCfg.Coordination)
	}
	redisCfg := configSvc.GetClusterRedis(ctx)
	if redisCfg == nil {
		return nil, gerror.New("cluster.redis is required when cluster.coordination=redis")
	}
	return coordination.NewRedis(ctx, coordination.RedisOptions{
		Address:        redisCfg.Address,
		DB:             redisCfg.DB,
		Password:       redisCfg.Password,
		ConnectTimeout: redisCfg.ConnectTimeout,
		ReadTimeout:    redisCfg.ReadTimeout,
		WriteTimeout:   redisCfg.WriteTimeout,
		KeyBuilder:     coordination.DefaultKeyBuilder(),
	})
}

// closeHTTPCoordinationAfterInitError best-effort closes Redis coordination
// resources when later HTTP runtime construction fails.
func closeHTTPCoordinationAfterInitError(ctx context.Context, coordinationSvc coordination.Service) {
	if coordinationSvc == nil {
		return
	}
	if closeErr := coordinationSvc.Close(ctx); closeErr != nil {
		logger.Warningf(ctx, "close coordination after runtime init failure: %v", closeErr)
	}
}

// startHTTPRuntime starts cluster, plugin, and cron services in the order
// required for source-plugin handlers and dynamic runtime state to be ready.
func startHTTPRuntime(ctx context.Context, runtime *httpRuntime) error {
	runtime.clusterSvc.Start(ctx)

	// Auto-enable and source-upgrade validation run before plugin routes and
	// cron jobs are registered so stale plugin state fails the process early.
	if err := startupstats.Observe(ctx, startupstats.PhasePluginBootstrapAutoEnable, func() error {
		return runtime.pluginSvc.BootstrapAutoEnable(ctx)
	}); err != nil {
		return err
	}
	if err := startupstats.Observe(ctx, startupstats.PhasePluginSourceUpgradeReadiness, func() error {
		return runtime.pluginSvc.ValidateSourcePluginUpgradeReadiness(ctx)
	}); err != nil {
		return err
	}
	if err := validateHTTPStartupPluginConsistency(ctx, runtime.pluginSvc); err != nil {
		return err
	}
	if err := startupstats.Observe(ctx, startupstats.PhasePluginLifecycleAttach, func() error {
		_, attachErr := jobhandlersvc.AttachPluginLifecycle(
			ctx,
			runtime.jobRegistry,
			runtime.pluginSvc,
		)
		return attachErr
	}); err != nil {
		return err
	}

	// Cron startup comes after plugin lifecycle wiring so plugin-owned scheduled
	// jobs are visible when the persistent scheduler loads enabled jobs.
	if err := startupstats.Observe(ctx, startupstats.PhaseCronStart, func() error {
		runtime.cronSvc.Start(ctx)
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// validateHTTPStartupPluginConsistency fails fast before route publication when
// persisted plugin or tenant-governance state is incoherent.
func validateHTTPStartupPluginConsistency(ctx context.Context, pluginSvc pluginStartupConsistencyValidator) error {
	if pluginSvc == nil {
		return nil
	}
	err := startupstats.Observe(ctx, startupstats.PhasePluginStartupConsistency, func() error {
		return pluginSvc.ValidateStartupConsistency(ctx)
	})
	if err != nil {
		logger.Panicf(ctx, "plugin startup consistency validation failed: %v", err)
	}
	return err
}

// logHTTPStartupSummary emits the startup metric summary without ORM SQL text.
func logHTTPStartupSummary(ctx context.Context, collector *startupstats.Collector) {
	if collector == nil {
		return
	}
	snapshot := collector.Snapshot()
	logger.Infof(
		ctx,
		"startup summary elapsed=%s catalogSnapshots=%d integrationSnapshots=%d jobSnapshots=%d pluginScans=%d pluginItems=%d pluginChanged=%d pluginNoop=%d menuChanged=%d menuNoop=%d resourceChanged=%d resourceNoop=%d builtinJobs=%d builtinNoop=%d persistentJobs=%d",
		snapshot.Elapsed.Round(time.Millisecond),
		snapshot.CounterValue(startupstats.CounterCatalogSnapshotBuilds),
		snapshot.CounterValue(startupstats.CounterIntegrationSnapshotBuilds),
		snapshot.CounterValue(startupstats.CounterJobSnapshotBuilds),
		snapshot.CounterValue(startupstats.CounterPluginScans),
		snapshot.CounterValue(startupstats.CounterPluginScanItems),
		snapshot.CounterValue(startupstats.CounterPluginSyncChanged),
		snapshot.CounterValue(startupstats.CounterPluginSyncNoop),
		snapshot.CounterValue(startupstats.CounterPluginMenuSyncChanged),
		snapshot.CounterValue(startupstats.CounterPluginMenuSyncNoop),
		snapshot.CounterValue(startupstats.CounterPluginResourceSyncChanged),
		snapshot.CounterValue(startupstats.CounterPluginResourceSyncNoop),
		snapshot.CounterValue(startupstats.CounterBuiltinJobProjections),
		snapshot.CounterValue(startupstats.CounterBuiltinJobProjectionNoop),
		snapshot.CounterValue(startupstats.CounterPersistentJobStartupLoaded),
	)
	for _, phase := range snapshot.PhaseNames() {
		logger.Debugf(ctx, "startup phase duration phase=%s duration=%s", phase, snapshot.Phases[phase].Round(time.Millisecond))
	}
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

	if runtime != nil && runtime.coordinationSvc != nil {
		if err := shutdownStep(shutdownCtx, "coordination service", func(stepCtx context.Context) error {
			return runtime.coordinationSvc.Close(stepCtx)
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
