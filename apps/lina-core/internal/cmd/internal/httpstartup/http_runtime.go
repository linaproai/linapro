// This file maintains HTTP runtime services and process-level server settings.

package httpstartup

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/cmd/internal/dbconfig"
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
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	"lina-core/internal/service/startupstats"
	"lina-core/internal/service/sysconfig"
	sysinfosvc "lina-core/internal/service/sysinfo"
	"lina-core/internal/service/user"
	"lina-core/internal/service/usermsg"
	"lina-core/pkg/dialect"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability"
	"lina-core/pkg/plugin/capability/aicap/aitext"
	pluginservicehostconfig "lina-core/pkg/plugin/capability/hostconfigcap"
	pluginservicemanifest "lina-core/pkg/plugin/capability/manifestcap"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	pluginserviceconfig "lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
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
	bizCtxSvc       bizctx.Service           // bizCtxSvc owns request-scoped business context mutation.
	i18nSvc         i18nsvc.Service          // i18nSvc owns runtime language bundles and localization.
	aiTextSvc       aitext.Service           // aiTextSvc exposes optional text AI capability.
	orgCapSvc       orgcap.Service           // orgCapSvc exposes optional organization capability.
	orgProjection   orgspi.ProjectionService // orgProjection exposes host user-management organization projections.
	roleSvc         role.Service             // roleSvc owns permission and access snapshot state.
	sessionStore    session.Store            // sessionStore owns online-session persistence and hot state.
	tenantSvc       tenantspi.RuntimeService // tenantSvc exposes optional linapro-tenant-core capability.
	kvCacheSvc      kvcache.Service          // kvCacheSvc owns runtime-selected KV backend.
	capabilities    capability.Services      // capabilities publishes runtime-owned adapters to plugins.
	dictSvc         dict.Service             // dictSvc owns dictionary lookup and maintenance.
	fileSvc         file.Service             // fileSvc owns file metadata and storage operations.
	menuSvc         menu.Service             // menuSvc owns menu tree and permission menu lookup.
	notifySvc       notify.Service           // notifySvc owns unified notification delivery.
	sysConfigSvc    sysconfig.Service        // sysConfigSvc owns mutable runtime configuration records.
	sysInfoSvc      sysinfosvc.Service       // sysInfoSvc owns runtime diagnostics projection.
	userSvc         user.Service             // userSvc owns host user management operations.
	userMsgSvc      usermsg.Service          // userMsgSvc owns current-user inbox operations.
	apiDocSvc       apidoc.Service           // apiDocSvc builds the host-managed OpenAPI document.
	jobRegistry     jobhandlersvc.Registry
	jobMgmtSvc      jobmgmtsvc.Service
	middlewareSvc   middleware.Service
	cronSvc         cron.Service
	serverCfg       *config.ServerExtensionsConfig // serverCfg contains host extension route settings such as API docs.
}

// pluginStartupConsistencyValidator is the narrow startup contract required to
// fail fast before the HTTP server starts serving requests.
type pluginStartupConsistencyValidator interface {
	// ValidateStartupConsistency verifies persisted plugin and tenant governance state.
	ValidateStartupConsistency(ctx context.Context) error
}

// pluginStartupTenantProvisioner is the narrow startup contract required to
// apply plugin.autoEnable tenant-scoped policies after source providers register.
type pluginStartupTenantProvisioner interface {
	// ReconcileAutoEnabledTenantPlugins provisions startup auto-enabled tenant plugins.
	ReconcileAutoEnabledTenantPlugins(ctx context.Context) error
}

// pluginManagementListPrewarmer is the startup-only contract for warming the
// plugin management read model after runtime plugin state has converged.
type pluginManagementListPrewarmer interface {
	// PrewarmManagementList builds the plugin management summary list read model and returns build errors to the caller.
	PrewarmManagementList(ctx context.Context) error
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

	// Request-size limits are enforced by host middleware so multipart uploads
	// can follow the runtime-effective sys.upload.maxSize value per request
	// instead of being clipped by GoFrame's static 8MB default at server entry.
	server.SetClientMaxBodySize(0)
	return nil
}

// newHTTPRuntime constructs the shared services used by the HTTP server and
// keeps their startup dependencies in one place.
func newHTTPRuntime(ctx context.Context, configSvc config.Service) (*httpRuntime, error) {
	link, err := dbconfig.CurrentDatabaseLink(ctx)
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
		locker.ConfigureCoordination(coordinationSvc)
		session.ConfigureCoordination(coordinationSvc)
	} else {
		locker.ConfigureCoordination(nil)
		session.ConfigureCoordination(nil)
	}
	kvCacheProvider, err := newHTTPKVCacheProvider(clusterCfg, coordinationSvc)
	if err != nil {
		closeHTTPCoordinationAfterInitError(ctx, coordinationSvc)
		return nil, err
	}

	// ========================================================================
	// Dependence injections.
	// ========================================================================

	var (
		bizCtxSvc        = bizctx.New()
		sessionStore     = session.NewDBStore()
		cacheCoordSvc    = cachecoord.Default(clusterSvc)
		i18nSvc          = i18nsvc.New(bizCtxSvc, configSvc, cacheCoordSvc)
		lockerSvc        = locker.New()
		lockStore        = runtimeUpgradeLockStore(coordinationSvc)
		pluginRuntime    = pluginsvc.NewRuntimeDelegate()
		kvCacheSvc       = kvcache.New(kvcache.WithProvider(kvCacheProvider))
		fileStorage      = file.NewLocalStorage(configSvc.GetUploadPath(ctx))
		jobRegistry      = jobhandlersvc.New()
		hostConfigReader pluginservicehostconfig.RawConfigReader
	)
	var (
		tenantProviderManager = tenantspi.NewManager()
		orgProviderManager    = orgspi.NewManager()
		aiTextProviderManager = aitext.NewManager()
	)
	var (
		orgCapSvc     = orgspi.New(orgProviderManager, pluginRuntime)
		aiTextSvc     = aitext.New(aiTextProviderManager, pluginRuntime)
		orgProjection = orgCapSvc
		tenantSvc     = tenantspi.New(tenantProviderManager, pluginRuntime, bizCtxSvc)
		roleSvc       = role.New(pluginRuntime, bizCtxSvc, configSvc, i18nSvc, orgCapSvc, tenantSvc)
		scopeSvc      = datascope.New(bizCtxSvc, roleSvc, orgCapSvc)
		dictSvc       = dict.New(i18nSvc)
		menuSvc       = menu.New(pluginRuntime, i18nSvc, roleSvc, tenantSvc)
		notifySvc     = notify.New(tenantSvc)
		authSvc       = auth.New(configSvc, pluginRuntime, orgCapSvc, roleSvc, tenantSvc, sessionStore, kvCacheSvc)
		fileSvc       = file.New(configSvc, fileStorage, bizCtxSvc, dictSvc, scopeSvc)
		sysConfigSvc  = sysconfig.New(configSvc, i18nSvc)
		userSvc       = user.New(authSvc, bizCtxSvc, i18nSvc, orgCapSvc, orgCapSvc, orgCapSvc, roleSvc, scopeSvc, tenantSvc, tenantSvc, tenantSvc)
		userMsgSvc    = usermsg.New(bizCtxSvc, notifySvc, i18nSvc)
		apiDocSvc     = apidoc.New(configSvc, bizCtxSvc, i18nSvc, pluginRuntime)
		authTokenSvc  = authSvc.(auth.TenantTokenIssuer)
	)
	sysInfoSvc, err := sysinfosvc.New(configSvc, clusterSvc, coordinationSvc, cacheCoordSvc)
	if err != nil {
		closeHTTPCoordinationAfterInitError(ctx, coordinationSvc)
		return nil, err
	}
	jobScheduler, err := jobmgmtsvc.NewScheduler(clusterSvc, jobRegistry, configSvc)
	if err != nil {
		closeHTTPCoordinationAfterInitError(ctx, coordinationSvc)
		return nil, err
	}
	var ok bool
	hostConfigReader, ok = configSvc.(pluginservicehostconfig.RawConfigReader)
	if !ok {
		closeHTTPCoordinationAfterInitError(ctx, coordinationSvc)
		return nil, gerror.New("host config service does not support raw reads")
	}
	var (
		jobMgmtSvc            = jobmgmtsvc.New(bizCtxSvc, configSvc, i18nSvc, jobRegistry, jobScheduler, scopeSvc)
		middlewareSvc         = middleware.New(authSvc, bizCtxSvc, configSvc, i18nSvc, roleSvc, tenantSvc)
		hostConfigSvc         = pluginservicehostconfig.New(hostConfigReader)
		pluginConfigFactory   = pluginserviceconfig.NewConfigFactory("", "")
		pluginManifestFactory = pluginservicemanifest.NewFactory("")
	)
	hostLockSvc, err := hostlock.New(lockerSvc)
	if err != nil {
		closeHTTPCoordinationAfterInitError(ctx, coordinationSvc)
		return nil, err
	}
	storageRuntime := pluginsvc.NewStorageProviderRuntime(pluginRuntime)
	localStorageProvider := pluginsvc.NewLocalStorageProvider(configSvc.GetPluginDynamicStoragePath(ctx))
	capabilities, err := pluginsvc.NewHostServices(
		apiDocSvc,
		authSvc,
		authTokenSvc,
		bizCtxSvc,
		hostConfigSvc,
		scopeSvc,
		i18nSvc,
		pluginRuntime,
		pluginRuntime,
		sessionStore,
		aiTextSvc,
		orgCapSvc,
		tenantSvc,
		notifySvc,
		kvCacheSvc,
		hostLockSvc,
		storageRuntime,
		localStorageProvider,
	)
	if err != nil {
		closeHTTPCoordinationAfterInitError(ctx, coordinationSvc)
		return nil, err
	}
	roleSvc.SetDataScopeService(scopeSvc)
	pluginSvc, err := pluginsvc.New(
		clusterSvc,
		configSvc,
		bizCtxSvc,
		cacheCoordSvc,
		i18nSvc,
		sessionStore,
		roleSvc,
		lockerSvc,
		lockStore,
		capabilities,
		orgCapSvc,
		tenantSvc,
		tenantSvc,
		tenantSvc,
		pluginConfigFactory,
		hostConfigSvc,
		pluginManifestFactory,
	)
	if err != nil {
		closeHTTPCoordinationAfterInitError(ctx, coordinationSvc)
		return nil, err
	}
	if err = pluginRuntime.BindService(pluginSvc); err != nil {
		closeHTTPCoordinationAfterInitError(ctx, coordinationSvc)
		return nil, err
	}
	if err = pluginSvc.RegisterSourcePluginProviderFactories(
		tenantProviderManager,
		orgProviderManager,
		aiTextProviderManager,
	); err != nil {
		closeHTTPCoordinationAfterInitError(ctx, coordinationSvc)
		return nil, err
	}

	// Host-owned handler definitions are registered before cron startup so the
	// persistent scheduler can project and validate code-owned jobs immediately.
	if err = jobhandlersvc.RegisterHostHandlers(jobRegistry, jobMgmtSvc); err != nil {
		closeHTTPCoordinationAfterInitError(ctx, coordinationSvc)
		return nil, err
	}

	sessionCfg, err := configSvc.GetSession(ctx)
	if err != nil {
		closeHTTPCoordinationAfterInitError(ctx, coordinationSvc)
		return nil, err
	}
	var cronSvc = cron.New(
		configSvc, roleSvc, kvCacheSvc,
		pluginSvc, sessionCfg, sessionStore,
		clusterSvc, jobRegistry, jobMgmtSvc, jobScheduler,
	)
	return &httpRuntime{
		configSvc:       configSvc,
		coordinationSvc: coordinationSvc,
		clusterSvc:      clusterSvc,
		pluginSvc:       pluginSvc,
		authSvc:         authSvc,
		authTokenIssuer: authTokenSvc,
		bizCtxSvc:       bizCtxSvc,
		i18nSvc:         i18nSvc,
		aiTextSvc:       aiTextSvc,
		orgCapSvc:       orgCapSvc,
		orgProjection:   orgProjection,
		roleSvc:         roleSvc,
		sessionStore:    sessionStore,
		tenantSvc:       tenantSvc,
		kvCacheSvc:      kvCacheSvc,
		capabilities:    capabilities,
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
		cronSvc:         cronSvc,
		serverCfg:       configSvc.GetServerExtensions(ctx),
	}, nil
}

// newHTTPKVCacheProvider selects the shared KV cache backend for the current
// HTTP runtime topology without mutating process-wide defaults.
func newHTTPKVCacheProvider(
	clusterCfg *config.ClusterConfig,
	coordinationSvc coordination.Service,
) (kvcache.Provider, error) {
	if clusterCfg != nil && clusterCfg.Enabled {
		if coordinationSvc == nil {
			return nil, gerror.New("cluster kvcache backend requires coordination service")
		}
		return kvcache.NewCoordinationKVProvider(coordinationSvc), nil
	}
	return kvcache.NewSQLTableProvider(), nil
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

// runtimeUpgradeLockStore extracts the cluster coordination lock store used by
// plugin runtime upgrades. Single-node deployments pass nil explicitly.
func runtimeUpgradeLockStore(coordinationSvc coordination.Service) coordination.LockStore {
	if coordinationSvc == nil {
		return nil
	}
	return coordinationSvc.Lock()
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

// startHTTPRuntime starts the complete runtime in the default order used by
// tests and non-HTTP callers that do not need to insert source route binding.
func startHTTPRuntime(ctx context.Context, runtime *httpRuntime) error {
	if err := startHTTPRuntimeBeforeSourceRoutes(ctx, runtime); err != nil {
		return err
	}
	return finishHTTPRuntimeAfterSourceRoutes(ctx, runtime)
}

// startHTTPRuntimeBeforeSourceRoutes starts cluster coordination and plugin
// bootstrap work that must finish before source plugins publish HTTP routes.
func startHTTPRuntimeBeforeSourceRoutes(ctx context.Context, runtime *httpRuntime) error {
	runtime.clusterSvc.Start(ctx)

	// Auto-enable and source-upgrade drift scanning run before plugin routes and
	// cron jobs are registered so plugin management can surface runtime state.
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
	return nil
}

// finishHTTPRuntimeAfterSourceRoutes validates startup consistency and starts
// runtime work that depends on source-plugin provider and route registration.
func finishHTTPRuntimeAfterSourceRoutes(ctx context.Context, runtime *httpRuntime) error {
	if err := reconcileHTTPStartupAutoEnabledTenantPlugins(ctx, runtime.pluginSvc); err != nil {
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
	startHTTPPluginManagementListPrewarm(ctx, runtime.pluginSvc)
	return nil
}

// reconcileHTTPStartupAutoEnabledTenantPlugins provisions tenant-scoped
// auto-enabled plugins after source plugin callbacks have registered providers.
func reconcileHTTPStartupAutoEnabledTenantPlugins(ctx context.Context, pluginSvc pluginStartupTenantProvisioner) error {
	if pluginSvc == nil {
		return nil
	}
	return startupstats.Observe(ctx, startupstats.PhasePluginTenantAutoProvisioning, func() error {
		return pluginSvc.ReconcileAutoEnabledTenantPlugins(ctx)
	})
}

// validateHTTPStartupPluginConsistency fails fast before the HTTP server starts
// when persisted plugin or tenant-governance state is incoherent.
func validateHTTPStartupPluginConsistency(ctx context.Context, pluginSvc pluginStartupConsistencyValidator) error {
	if pluginSvc == nil {
		return nil
	}
	err := startupstats.Observe(ctx, startupstats.PhasePluginStartupConsistency, func() error {
		return pluginSvc.ValidateStartupConsistency(ctx)
	})
	if err != nil {
		logger.Errorf(ctx, "plugin startup consistency validation failed: %v", err)
	}
	return err
}

// startHTTPPluginManagementListPrewarm warms the plugin management read model
// after startup convergence without delaying HTTP availability.
func startHTTPPluginManagementListPrewarm(ctx context.Context, pluginSvc pluginManagementListPrewarmer) {
	if pluginSvc == nil {
		return
	}
	prewarmCtx := context.WithoutCancel(ctx)
	go func() {
		startedAt := time.Now()
		if err := pluginSvc.PrewarmManagementList(prewarmCtx); err != nil {
			logger.Debugf(
				prewarmCtx,
				"prewarm plugin management list finished status=failed duration=%s",
				time.Since(startedAt).Round(time.Millisecond),
			)
			logger.Warningf(prewarmCtx, "prewarm plugin management list failed: %v", err)
			return
		}
		logger.Debugf(
			prewarmCtx,
			"prewarm plugin management list finished status=succeeded duration=%s",
			time.Since(startedAt).Round(time.Millisecond),
		)
	}()
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
func shutdownHTTPRuntime(ctx context.Context, runtime *httpRuntime, server *ghttp.Server) error {
	shutdownBaseCtx := context.WithoutCancel(ctx)
	shutdownTimeout := resolveRuntimeShutdownTimeout(server)
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

// resolveRuntimeShutdownTimeout returns the host-owned cleanup budget from the
// GoFrame HTTP server graceful shutdown configuration already active at startup.
func resolveRuntimeShutdownTimeout(server *ghttp.Server) time.Duration {
	if server == nil {
		return time.Duration(ghttp.NewConfig().GracefulShutdownTimeout) * time.Second
	}
	return time.Duration(server.GetGracefulShutdownTimeout()) * time.Second
}
