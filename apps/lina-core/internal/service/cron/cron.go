// Package cron implements host-level scheduled jobs such as session cleanup
// and local runtime-cache synchronization.
package cron

import (
	"context"
	"sync"

	"lina-core/internal/service/cluster"
	"lina-core/internal/service/config"
	jobhandlersvc "lina-core/internal/service/jobhandler"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
	pluginsvc "lina-core/internal/service/plugin"
	rolesvc "lina-core/internal/service/role"
	"lina-core/internal/service/session"
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

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	configSvc           config.Service         // Config service
	roleSvc             rolesvc.Service        // Role service
	sessionStore        session.Store          // Session store
	clusterSvc          cluster.Service        // Cluster topology service
	registry            jobhandlersvc.Registry // registry stores managed host and plugin handlers.
	pluginSvc           pluginsvc.Service      // pluginSvc exposes plugin lifecycle and job declarations.
	builtinSyncer       jobmgmtsvc.Service     // builtinSyncer persists code-owned job definitions.
	persistentScheduler jobmgmtsvc.Scheduler   // persistentScheduler loads and registers persisted jobs.
	pluginObserverOnce  sync.Once              // pluginObserverOnce avoids duplicate lifecycle subscriptions.
}

// New creates and returns a new Service instance.
func New(
	configSvc config.Service,
	roleSvc rolesvc.Service,
	pluginSvc pluginsvc.Service,
	sessionStore session.Store,
	clusterSvc cluster.Service,
	registry jobhandlersvc.Registry,
	builtinSyncer jobmgmtsvc.Service,
	persistentScheduler jobmgmtsvc.Scheduler,
) Service {
	return &serviceImpl{
		configSvc:           configSvc,
		roleSvc:             roleSvc,
		sessionStore:        sessionStore,
		clusterSvc:          clusterSvc,
		registry:            registry,
		pluginSvc:           pluginSvc,
		builtinSyncer:       builtinSyncer,
		persistentScheduler: persistentScheduler,
	}
}
