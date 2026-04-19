// This file keeps manual cancellation support for running scheduled-job instances.

package scheduler

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
)

// CancelLog cancels one currently running job-log instance.
func (s *serviceImpl) CancelLog(_ context.Context, logID uint64) error {
	s.mu.Lock()
	execution, ok := s.runningInstances[logID]
	s.mu.Unlock()
	if !ok || execution == nil || execution.cancel == nil {
		return gerror.New("当前日志实例未在运行")
	}
	execution.cancel()
	return nil
}
