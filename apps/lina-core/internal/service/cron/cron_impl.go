// This file contains the cron service lifecycle methods that wire startup
// jobs, persistent scheduled jobs, and graceful scheduler shutdown.

package cron

import (
	"context"

	"github.com/gogf/gf/v2/os/gcron"

	"lina-core/pkg/logger"
)

// Start registers and starts all cron jobs.
func (s *serviceImpl) Start(ctx context.Context) {
	s.startAccessTopologyRevisionSync(ctx)
	s.startRuntimeParamSnapshotSync(ctx)
	s.attachPluginLifecycleObserver()

	if err := s.syncBuiltinScheduledJobs(ctx); err != nil {
		logger.Warningf(ctx, "sync builtin scheduled jobs failed: %v", err)
	}
	if s.persistentScheduler != nil {
		if err := s.persistentScheduler.LoadAndRegister(ctx); err != nil {
			logger.Warningf(ctx, "register persistent cron jobs failed: %v", err)
		}
	}
}

// Stop gracefully stops cron scheduling and waits for in-flight jobs.
func (s *serviceImpl) Stop(ctx context.Context) {
	doneCtx := gcron.StopGracefullyNonBlocking()
	select {
	case <-doneCtx.Done():
		return
	case <-ctx.Done():
		logger.Warningf(ctx, "cron graceful stop timed out or was canceled: %v", ctx.Err())
	}
}

// IsPrimary reports whether the current node should execute primary-only jobs.
func (s *serviceImpl) IsPrimary() bool {
	if s == nil || s.clusterSvc == nil {
		return true
	}
	return s.clusterSvc.IsPrimary()
}
