// This file registers the stale server-monitor cleanup job.

package cron

import (
	"context"
	"lina-core/pkg/logger"
	"time"

	"github.com/gogf/gf/v2/os/gcron"
)

// startServerMonitorCleanup starts the stale monitor records cleanup job.
// It runs every hour to delete records where updated_at is older than
// (collection_interval * retention_multiplier) seconds.
// This is a primary-only job, only executed on the primary node in clustered mode.
func (s *serviceImpl) startServerMonitorCleanup(ctx context.Context) {
	// Calculate stale threshold: interval * multiplier
	staleThreshold := s.monCfg.Interval * time.Duration(s.monCfg.RetentionMultiplier)

	_, err := gcron.Add(ctx, "# * * * * *", func(ctx context.Context) {
		if !s.IsPrimary() {
			logger.Debug(ctx, "skipping server monitor cleanup on non-primary node")
			return
		}
		cleaned, cleanErr := s.serverMonSvc.CleanupStale(ctx, staleThreshold)
		if cleanErr != nil {
			logger.Errorf(ctx, "failed to cleanup stale monitor records: %v", cleanErr)
			return
		}
		if cleaned > 0 {
			logger.Infof(
				ctx,
				"cleaned up %d stale monitor records (older than %v)",
				cleaned, time.Now().Add(-staleThreshold).Format("2006-01-02 15:04:05"),
			)
		}
	}, CronServerMonitorCleanup)
	if err != nil {
		logger.Panicf(ctx, "failed to start server monitor cleanup cron: %v", err)
	}
}
