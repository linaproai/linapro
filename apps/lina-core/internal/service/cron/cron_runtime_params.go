// This file assembles startup jobs for runtime-parameter snapshot warming and
// optional cluster watcher registration.

package cron

import (
	"context"

	hostconfig "lina-core/internal/service/config"
	"lina-core/pkg/logger"
)

// localRuntimeParamSnapshotSyncJob performs startup warm-up for protected
// runtime-parameter snapshots in single-node mode.
type localRuntimeParamSnapshotSyncJob struct {
	configSvc hostconfig.Service
}

// clusterRuntimeParamSnapshotSyncJob performs warm-up and registers the
// periodic cluster watcher for runtime-parameter revisions.
type clusterRuntimeParamSnapshotSyncJob struct {
	configSvc hostconfig.Service
}

// newRuntimeParamSnapshotSyncJob hides the deployment split behind one startup
// job so cron.Start can stay linear and free of mode branches.
func newRuntimeParamSnapshotSyncJob(
	clusterEnabled bool,
	configSvc hostconfig.Service,
) startupJob {
	if clusterEnabled {
		return &clusterRuntimeParamSnapshotSyncJob{configSvc: configSvc}
	}
	return &localRuntimeParamSnapshotSyncJob{configSvc: configSvc}
}

// startRuntimeParamSnapshotSync forwards to the preselected startup job so all
// mode-specific decisions stay in service construction instead of startup flow.
func (s *serviceImpl) startRuntimeParamSnapshotSync(ctx context.Context) {
	if s == nil || s.runtimeParamSyncJob == nil {
		return
	}
	s.runtimeParamSyncJob.Start(ctx)
}

// Start only performs eager warmup because single-node mode has no cross-node
// writer that requires a background watcher to reconcile local cache state.
func (j *localRuntimeParamSnapshotSyncJob) Start(ctx context.Context) {
	if j == nil || j.configSvc == nil {
		return
	}

	// Warm the local snapshot during startup so the first protected request does
	// not need to cold-load sys_config unless startup sync itself fails.
	if err := j.configSvc.SyncRuntimeParamSnapshot(ctx); err != nil {
		logger.Warningf(ctx, "initial runtime param snapshot sync failed: %v", err)
	}
}

// Start reuses the local warmup first. The periodic watcher is projected into
// sys_job and executed by the persistent scheduled-job scheduler.
func (j *clusterRuntimeParamSnapshotSyncJob) Start(ctx context.Context) {
	if j == nil || j.configSvc == nil {
		return
	}

	(&localRuntimeParamSnapshotSyncJob{configSvc: j.configSvc}).Start(ctx)
}
