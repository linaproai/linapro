// Package cron implements host-level scheduled jobs such as session cleanup
// and local runtime-cache synchronization.
package cron

import (
	"context"
	"sync"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/cluster"
	"lina-core/internal/service/config"
	jobhandlersvc "lina-core/internal/service/jobhandler"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
	"lina-core/internal/service/kvcache"
	pluginsvc "lina-core/internal/service/plugin"
	rolesvc "lina-core/internal/service/role"
	"lina-core/internal/service/session"
)

// Cron job name constants.
const (
	CronSessionCleanup     = "session-cleanup"      // Session cleanup job name
	CronAccessTopologySync = "access-topology-sync" // Access topology sync job name
	CronRuntimeParamSync   = "runtime-param-sync"   // Runtime-parameter snapshot sync job name
)

// Service defines the cron service contract.
type Service interface {
	// Start registers and starts all cron jobs for the current node. The method
	// also reconciles code-owned persistent jobs and logs startup failures rather
	// than returning them because cron startup is invoked from host bootstrapping.
	Start(ctx context.Context)
	// Stop gracefully stops cron scheduling and waits for in-flight jobs. The
	// provided context bounds the wait; timeout or cancellation is logged.
	Stop(ctx context.Context)
	// IsPrimary reports whether the current node should execute primary-only jobs.
	// Standalone deployments and services without a cluster dependency are
	// treated as primary to preserve single-node behavior.
	IsPrimary() bool
}

// builtinJobSyncer syncs code-owned scheduled-job definitions into sys_job.
type builtinJobSyncer interface {
	// ReconcileBuiltinJobs refreshes code-owned scheduled-job projections and
	// returns declaration-derived snapshots keyed with sys_job IDs.
	ReconcileBuiltinJobs(ctx context.Context, jobs []jobmgmtsvc.BuiltinJobDef) ([]*entity.SysJob, error)
}

// startupJob abstracts warm-up and watcher registration logic selected during
// service construction for single-node or clustered deployments.
type startupJob interface {
	// Start performs eager warmup and registers any required background watchers.
	Start(ctx context.Context)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	sessionCfg            *config.SessionConfig  // Session configuration
	configSvc             config.Service         // Config service
	roleSvc               rolesvc.Service        // Role service
	kvCacheSvc            kvcache.Service        // kvCacheSvc cleans backend-owned expired cache rows.
	sessionStore          session.Store          // Session store
	clusterSvc            cluster.Service        // Cluster topology service
	registry              jobhandlersvc.Registry // registry stores managed host and plugin handlers.
	pluginSvc             pluginsvc.Service      // Plugin service
	builtinSyncer         builtinJobSyncer       // builtinSyncer persists code-owned job definitions.
	persistentScheduler   jobmgmtsvc.Scheduler   // persistentScheduler loads and registers persisted jobs.
	runtimeParamSyncJob   startupJob             // Runtime-parameter sync startup job
	accessTopologySyncJob startupJob             // Permission-topology sync startup job
	managedHandlersOnce   sync.Once              // managedHandlersOnce avoids duplicate handler registration.
	pluginObserverOnce    sync.Once              // pluginObserverOnce avoids duplicate lifecycle subscriptions.
}

// New creates and returns a new Service instance.
func New(
	configSvc config.Service,
	roleSvc rolesvc.Service,
	kvCacheSvc kvcache.Service,
	pluginSvc pluginsvc.Service,
	sessionCfg *config.SessionConfig,
	sessionStore session.Store,
	clusterSvc cluster.Service,
	registry jobhandlersvc.Registry,
	builtinSyncer builtinJobSyncer,
	persistentScheduler jobmgmtsvc.Scheduler,
) Service {
	clusterEnabled := clusterSvc != nil && clusterSvc.IsEnabled()
	return &serviceImpl{
		sessionCfg:          sessionCfg,
		configSvc:           configSvc,
		roleSvc:             roleSvc,
		kvCacheSvc:          kvCacheSvc,
		sessionStore:        sessionStore,
		clusterSvc:          clusterSvc,
		registry:            registry,
		pluginSvc:           pluginSvc,
		builtinSyncer:       builtinSyncer,
		persistentScheduler: persistentScheduler,
		runtimeParamSyncJob: newRuntimeParamSnapshotSyncJob(
			clusterEnabled,
			configSvc,
		),
		accessTopologySyncJob: newAccessTopologyRevisionSyncJob(
			clusterEnabled,
			roleSvc,
		),
	}
}
