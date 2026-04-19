// Package cron implements host-level scheduled jobs such as cleanup,
// monitoring, and local runtime-cache synchronization.
package cron

import (
	"context"

	"lina-core/internal/service/cluster"
	"lina-core/internal/service/config"
	pluginsvc "lina-core/internal/service/plugin"
	rolesvc "lina-core/internal/service/role"
	"lina-core/internal/service/servermon"
	"lina-core/internal/service/session"
	"lina-core/pkg/logger"
)

// Cron job name constants.
const (
	CronSessionCleanup         = "session-cleanup"          // Session cleanup job name
	CronAccessTopologySync     = "access-topology-sync"     // Access topology sync job name
	CronRuntimeParamSync       = "runtime-param-sync"       // Runtime-parameter snapshot sync job name
	CronServerMonitorCollector = "server-monitor-collector" // Server monitor collector job name
	CronServerMonitorCleanup   = "server-monitor-cleanup"   // Server monitor cleanup job name
)

// Service defines the cron service contract.
type Service interface {
	// Start registers and starts all cron jobs.
	Start(ctx context.Context)
	// IsPrimary reports whether the current node should execute primary-only jobs.
	IsPrimary() bool
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
	sessionCfg            *config.SessionConfig // Session configuration
	monCfg                *config.MonitorConfig // Monitor configuration
	configSvc             config.Service        // Config service
	roleSvc               rolesvc.Service       // Role service
	serverMonSvc          servermon.Service     // Server monitor service
	sessionStore          session.Store         // Session store
	clusterSvc            cluster.Service       // Cluster topology service
	pluginSvc             pluginsvc.Service     // Plugin service
	runtimeParamSyncJob   startupJob            // Runtime-parameter sync startup job
	accessTopologySyncJob startupJob            // Permission-topology sync startup job
}

// New creates and returns a new Service instance.
func New(
	sessionCfg *config.SessionConfig,
	monCfg *config.MonitorConfig,
	sessionStore session.Store,
	clusterSvc cluster.Service,
) Service {
	var (
		configSvc      = config.New()
		roleSvc        = rolesvc.New()
		serverMonSvc   = servermon.New()
		pluginSvc      = pluginsvc.New(clusterSvc)
		clusterEnabled = clusterSvc != nil && clusterSvc.IsEnabled()
	)

	return &serviceImpl{
		sessionCfg:   sessionCfg,
		monCfg:       monCfg,
		configSvc:    configSvc,
		roleSvc:      roleSvc,
		serverMonSvc: serverMonSvc,
		sessionStore: sessionStore,
		clusterSvc:   clusterSvc,
		pluginSvc:    pluginSvc,
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

// Start registers and starts all cron jobs.
func (s *serviceImpl) Start(ctx context.Context) {
	// All-Node Jobs: executed on every node
	s.startServerMonitor(ctx)
	s.startAccessTopologyRevisionSync(ctx)
	s.startRuntimeParamSnapshotSync(ctx)

	// Master-Only Jobs: only executed on the leader node
	s.startSessionCleanup(ctx)
	s.startServerMonitorCleanup(ctx)
	if err := s.pluginSvc.RegisterCrons(ctx); err != nil {
		logger.Warningf(ctx, "register plugin cron jobs failed: %v", err)
	}
}

// IsPrimary reports whether the current node should execute primary-only jobs.
func (s *serviceImpl) IsPrimary() bool {
	if s == nil || s.clusterSvc == nil {
		return true
	}
	return s.clusterSvc.IsPrimary()
}
