// This file registers the periodic online-session cleanup job.

package cron

import (
	"context"
	"fmt"
	"lina-core/pkg/logger"

	"github.com/gogf/gf/v2/os/gcron"
)

// startSessionCleanup registers the session cleanup cron job.
// This is a primary-only job, only executed on the primary node in clustered mode.
func (s *serviceImpl) startSessionCleanup(ctx context.Context) {
	cronPattern := fmt.Sprintf("@every %dns", s.sessionCfg.CleanupInterval.Nanoseconds())
	_, err := gcron.Add(ctx, cronPattern, func(ctx context.Context) {
		if !s.IsPrimary() {
			logger.Debug(ctx, "skipping session cleanup on non-primary node")
			return
		}
		cleaned, cleanErr := s.sessionStore.CleanupInactive(ctx, s.sessionCfg.Timeout)
		if cleanErr != nil {
			logger.Warningf(ctx, "session cleanup error: %v", cleanErr)
		} else if cleaned > 0 {
			logger.Infof(ctx, "session cleanup: removed %d inactive sessions", cleaned)
		}
	}, CronSessionCleanup)
	if err != nil {
		logger.Panicf(ctx, "failed to start session cleanup cron: %v", err)
	}
}
