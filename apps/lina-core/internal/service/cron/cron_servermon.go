// This file registers the periodic server monitor collector job.

package cron

import (
	"context"
	"fmt"
	"lina-core/pkg/logger"

	"github.com/gogf/gf/v2/os/gcron"
)

// startServerMonitor starts the server monitor metrics collector.
func (s *serviceImpl) startServerMonitor(ctx context.Context) {
	// Collect immediately on startup
	s.serverMonSvc.CollectAndStore(ctx)

	// Then collect periodically via gcron
	cronPattern := fmt.Sprintf("@every %dns", s.monCfg.Interval.Nanoseconds())
	_, err := gcron.Add(ctx, cronPattern, func(ctx context.Context) {
		s.serverMonSvc.CollectAndStore(ctx)
	}, CronServerMonitorCollector)
	if err != nil {
		logger.Panicf(ctx, "failed to start server monitor cron: %v", err)
	}
}
