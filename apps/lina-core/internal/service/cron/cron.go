package cron

import (
	"context"

	"lina-core/internal/service/cluster"
	"lina-core/internal/service/config"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/servermon"
	"lina-core/internal/service/session"
	"lina-core/pkg/logger"
)

// Cron job name constants.
const (
	CronSessionCleanup         = "session-cleanup"          // Session cleanup job name
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

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	sessionCfg   *config.SessionConfig // Session configuration
	monCfg       *config.MonitorConfig // Monitor configuration
	serverMonSvc servermon.Service     // Server monitor service
	sessionStore session.Store         // Session store
	clusterSvc   cluster.Service       // Cluster topology service
	pluginSvc    pluginsvc.Service     // Plugin service
}

// New creates and returns a new Service instance.
func New(
	sessionCfg *config.SessionConfig,
	monCfg *config.MonitorConfig,
	sessionStore session.Store,
	clusterSvc cluster.Service,
) Service {
	return &serviceImpl{
		sessionCfg:   sessionCfg,
		monCfg:       monCfg,
		serverMonSvc: servermon.New(),
		sessionStore: sessionStore,
		clusterSvc:   clusterSvc,
		pluginSvc:    pluginsvc.New(clusterSvc),
	}
}

// Start registers and starts all cron jobs.
func (s *serviceImpl) Start(ctx context.Context) {
	// All-Node Jobs: executed on every node
	s.startServerMonitor(ctx)

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
