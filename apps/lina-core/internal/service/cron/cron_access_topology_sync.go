// This file assembles startup jobs for permission-topology revision warming
// and optional cluster watcher registration.

package cron

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/os/gcron"

	rolesvc "lina-core/internal/service/role"
	"lina-core/pkg/logger"
)

type localAccessTopologyRevisionSyncJob struct {
	roleSvc rolesvc.Service
}

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
	// can stay on process memory instead of reading shared KV on first access.
	if err := j.roleSvc.SyncAccessTopologyRevision(ctx); err != nil {
		logger.Warningf(ctx, "initial access topology sync failed: %v", err)
	}
}

// Start performs the same eager warmup as single-node mode and then adds the
// all-node watcher required to detect topology mutations from other instances.
func (j *clusterAccessTopologyRevisionSyncJob) Start(ctx context.Context) {
	if j == nil || j.roleSvc == nil {
		return
	}

	(&localAccessTopologyRevisionSyncJob{roleSvc: j.roleSvc}).Start(ctx)

	// Every node runs this watcher because permission checks happen locally on
	// every instance and should converge without waiting for request traffic.
	// The role service only stores the shared revision here; token-scoped access
	// snapshots are then evicted lazily when the watcher detects a new revision.
	cronPattern := fmt.Sprintf("@every %fs", rolesvc.AccessRevisionSyncInterval().Seconds())
	_, err := gcron.Add(ctx, cronPattern, func(ctx context.Context) {
		if syncErr := j.roleSvc.SyncAccessTopologyRevision(ctx); syncErr != nil {
			logger.Warningf(ctx, "access topology sync failed: %v", syncErr)
		}
	}, CronAccessTopologySync)
	if err != nil {
		logger.Panicf(ctx, "failed to start access topology sync cron: %v", err)
	}
}
