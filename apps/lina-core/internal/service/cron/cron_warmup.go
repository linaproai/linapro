// This file performs startup warming for local runtime snapshots that are
// later kept fresh by managed scheduled jobs in clustered deployments.

package cron

import (
	"context"

	"lina-core/pkg/logger"
)

// startAccessTopologyRevisionSync warms the local revision snapshot during
// startup so most protected requests can stay on process memory instead of
// reading cachecoord on first access.
func (s *serviceImpl) startAccessTopologyRevisionSync(ctx context.Context) {
	if s == nil || s.roleSvc == nil {
		return
	}

	if err := s.roleSvc.SyncAccessTopologyRevision(ctx); err != nil {
		logger.Warningf(ctx, "initial access topology sync failed: %v", err)
	}
}

// startRuntimeParamSnapshotSync warms the runtime-parameter snapshot during
// startup so the first protected request does not need to cold-load sys_config
// unless startup sync itself fails.
func (s *serviceImpl) startRuntimeParamSnapshotSync(ctx context.Context) {
	if s == nil || s.configSvc == nil {
		return
	}

	if err := s.configSvc.SyncRuntimeParamSnapshot(ctx); err != nil {
		logger.Warningf(ctx, "initial runtime param snapshot sync failed: %v", err)
	}
}
