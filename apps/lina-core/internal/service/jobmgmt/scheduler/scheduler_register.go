// This file keeps gcron registration and concurrency bookkeeping helpers for
// the persistent scheduled-job scheduler.

package scheduler

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/jobmeta"
)

// acquireSlot applies the per-job concurrency guard for cron-triggered executions.
func (s *serviceImpl) acquireSlot(job *entity.SysJob) (func(), jobmeta.LogStatus, error) {
	if job == nil {
		return func() {}, "", gerror.New("定时任务不存在")
	}

	concurrency := jobmeta.NormalizeJobConcurrency(job.Concurrency)
	maxConcurrency := job.MaxConcurrency
	if concurrency == jobmeta.JobConcurrencySingleton || maxConcurrency <= 0 {
		maxConcurrency = 1
	}

	s.mu.Lock()
	current := s.runningCounts[job.Id]
	if current >= maxConcurrency {
		s.mu.Unlock()
		if concurrency == jobmeta.JobConcurrencySingleton {
			return func() {}, jobmeta.LogStatusSkippedSingleton, nil
		}
		return func() {}, jobmeta.LogStatusSkippedMaxConcurrency, nil
	}
	s.runningCounts[job.Id] = current + 1
	s.mu.Unlock()

	released := false
	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if released {
			return
		}
		released = true
		if s.runningCounts[job.Id] <= 1 {
			delete(s.runningCounts, job.Id)
			return
		}
		s.runningCounts[job.Id]--
	}, "", nil
}

// storeRunningExecution stores one cancellable running instance.
func (s *serviceImpl) storeRunningExecution(
	logID uint64,
	jobID uint64,
	cancel context.CancelFunc,
	release func(),
) {
	s.mu.Lock()
	s.runningInstances[logID] = &runningExecution{
		jobID:  jobID,
		cancel: cancel,
		release: release,
	}
	s.mu.Unlock()
}

// finishRunningExecution removes one running instance and releases its slot.
func (s *serviceImpl) finishRunningExecution(logID uint64) {
	s.mu.Lock()
	execution, ok := s.runningInstances[logID]
	if ok {
		delete(s.runningInstances, logID)
	}
	s.mu.Unlock()

	if ok && execution.release != nil {
		execution.release()
	}
}
