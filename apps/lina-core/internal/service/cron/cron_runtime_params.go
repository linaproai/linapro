// This file assembles startup jobs for runtime-parameter snapshot warming and
// optional cluster watcher registration.

package cron

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/os/gcron"

	hostconfig "lina-core/internal/service/config"
	"lina-core/pkg/logger"
)

type localRuntimeParamSnapshotSyncJob struct {
	configSvc hostconfig.Service
}

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

// Start reuses the local warmup first and then registers the periodic watcher
// required for clustered deployments to converge after cross-node writes.
func (j *clusterRuntimeParamSnapshotSyncJob) Start(ctx context.Context) {
	if j == nil || j.configSvc == nil {
		return
	}

	(&localRuntimeParamSnapshotSyncJob{configSvc: j.configSvc}).Start(ctx)

	// Every node runs this watcher because runtime params are consumed locally on
	// every instance and should converge without waiting for request traffic.
	cronPattern := fmt.Sprintf("@every %fs", hostconfig.RuntimeParamSnapshotSyncInterval().Seconds())
	_, err := gcron.Add(ctx, cronPattern, func(ctx context.Context) {
		if syncErr := j.configSvc.SyncRuntimeParamSnapshot(ctx); syncErr != nil {
			logger.Warningf(ctx, "runtime param snapshot sync failed: %v", syncErr)
		}
	}, CronRuntimeParamSync)
	if err != nil {
		logger.Panicf(ctx, "failed to start runtime param snapshot sync cron: %v", err)
	}
}
