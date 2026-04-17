// This file registers the periodic runtime-parameter snapshot sync job.

package cron

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/os/gcron"

	hostconfig "lina-core/internal/service/config"
	"lina-core/pkg/logger"
)

// startRuntimeParamSnapshotSync registers the all-node runtime-parameter cache
// sync job and performs one eager sync during service startup.
func (s *serviceImpl) startRuntimeParamSnapshotSync(ctx context.Context) {
	if s == nil || s.configSvc == nil {
		return
	}

	// Warm the local snapshot during startup so the first protected request does
	// not need to cold-load sys_config unless startup sync itself fails.
	if err := s.configSvc.SyncRuntimeParamSnapshot(ctx); err != nil {
		logger.Warningf(ctx, "initial runtime param snapshot sync failed: %v", err)
	}

	// Every node runs this watcher because runtime params are consumed locally on
	// every instance and should converge without waiting for request traffic.
	cronPattern := fmt.Sprintf("@every %fs", hostconfig.RuntimeParamSnapshotSyncInterval().Seconds())
	_, err := gcron.Add(ctx, cronPattern, func(ctx context.Context) {
		if syncErr := s.configSvc.SyncRuntimeParamSnapshot(ctx); syncErr != nil {
			logger.Warningf(ctx, "runtime param snapshot sync failed: %v", syncErr)
		}
	}, CronRuntimeParamSync)
	if err != nil {
		logger.Panicf(ctx, "failed to start runtime param snapshot sync cron: %v", err)
	}
}
