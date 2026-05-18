// This file keeps manual cancellation support for running scheduled-job instances.

package scheduler

import (
	"context"

	"lina-core/internal/service/jobmeta"
	"lina-core/pkg/bizerr"
)

// CancelLog cancels one currently running job-log instance.
func (s *serviceImpl) CancelLog(_ context.Context, logID int64) error {
	s.mu.Lock()
	execution, ok := s.runningInstances[logID]
	s.mu.Unlock()
	if !ok || execution == nil || execution.cancel == nil {
		return bizerr.NewCode(jobmeta.CodeJobLogNotRunning)
	}
	execution.cancel()
	return nil
}
