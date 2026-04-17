// This file registers the periodic permission-topology revision sync job.

package cron

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/os/gcron"

	rolesvc "lina-core/internal/service/role"
	"lina-core/pkg/logger"
)

// startAccessTopologyRevisionSync registers the all-node permission-topology
// sync job and performs one eager sync during service startup.
func (s *serviceImpl) startAccessTopologyRevisionSync(ctx context.Context) {
	if s == nil || s.roleSvc == nil {
		return
	}

	// Warm the local revision snapshot during startup so most protected requests
	// can stay on process memory instead of reading shared KV on first access.
	if err := s.roleSvc.SyncAccessTopologyRevision(ctx); err != nil {
		logger.Warningf(ctx, "initial access topology sync failed: %v", err)
	}

	// Every node runs this watcher because permission checks happen locally on
	// every instance and should converge without waiting for request traffic.
	// The role service only stores the shared revision here; token-scoped access
	// snapshots are then evicted lazily when the watcher detects a new revision.
	cronPattern := fmt.Sprintf("@every %fs", rolesvc.AccessRevisionSyncInterval().Seconds())
	_, err := gcron.Add(ctx, cronPattern, func(ctx context.Context) {
		if syncErr := s.roleSvc.SyncAccessTopologyRevision(ctx); syncErr != nil {
			logger.Warningf(ctx, "access topology sync failed: %v", syncErr)
		}
	}, CronAccessTopologySync)
	if err != nil {
		logger.Panicf(ctx, "failed to start access topology sync cron: %v", err)
	}
}
