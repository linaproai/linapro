// This file assembles startup jobs for permission-topology revision warming
// and optional cluster watcher registration.

package cron

import (
	"context"

	rolesvc "lina-core/internal/service/role"
	"lina-core/pkg/logger"
)

// localAccessTopologyRevisionSyncJob warms the local permission-topology
// revision snapshot without registering a distributed watcher.
type localAccessTopologyRevisionSyncJob struct {
	roleSvc rolesvc.Service
}

// clusterAccessTopologyRevisionSyncJob warms the local revision snapshot and
// registers the cross-node watcher required in clustered deployments.
type clusterAccessTopologyRevisionSyncJob struct {
	roleSvc rolesvc.Service
}

// newAccessTopologyRevisionSyncJob hides the deployment split behind one
// startup job so cron.Start does not need permission-topology mode branches.
func newAccessTopologyRevisionSyncJob(
	clusterEnabled bool,
	roleSvc rolesvc.Service,
) startupJob {
	if clusterEnabled {
		return &clusterAccessTopologyRevisionSyncJob{roleSvc: roleSvc}
	}
	return &localAccessTopologyRevisionSyncJob{roleSvc: roleSvc}
}

// startAccessTopologyRevisionSync forwards to the preselected startup job so
// startup orchestration remains linear after the strategy refactor.
func (s *serviceImpl) startAccessTopologyRevisionSync(ctx context.Context) {
	if s == nil || s.accessTopologySyncJob == nil {
		return
	}
	s.accessTopologySyncJob.Start(ctx)
}

// Start only warms the local revision snapshot because single-node permission
// topology cannot diverge across instances.
func (j *localAccessTopologyRevisionSyncJob) Start(ctx context.Context) {
	if j == nil || j.roleSvc == nil {
		return
	}

	// Warm the local revision snapshot during startup so most protected requests
	// can stay on process memory instead of reading cachecoord on first access.
	if err := j.roleSvc.SyncAccessTopologyRevision(ctx); err != nil {
		logger.Warningf(ctx, "initial access topology sync failed: %v", err)
	}
}

// Start performs the same eager warmup as single-node mode. The periodic
// watcher itself is projected into sys_job and executed by the persistent
// scheduled-job scheduler.
func (j *clusterAccessTopologyRevisionSyncJob) Start(ctx context.Context) {
	if j == nil || j.roleSvc == nil {
		return
	}

	(&localAccessTopologyRevisionSyncJob{roleSvc: j.roleSvc}).Start(ctx)
}
