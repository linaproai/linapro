// This file keeps gcron registration and concurrency bookkeeping helpers for
// the persistent scheduled-job scheduler.

package scheduler

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/os/gcron"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/jobmeta"
	"lina-core/internal/service/startupstats"
	"lina-core/pkg/bizerr"
)

// LoadAndRegister registers all currently enabled persistent jobs at startup.
func (s *serviceImpl) LoadAndRegister(ctx context.Context) error {
	jobs, err := s.listEnabledJobs(ctx)
	if err != nil {
		return err
	}
	startupstats.Add(ctx, startupstats.CounterPersistentJobStartupLoaded, len(jobs))
	for _, job := range jobs {
		if err = s.registerJob(ctx, job); err != nil {
			if handled, handleErr := s.handleLoadRegisterError(ctx, job, err); handleErr != nil {
				return handleErr
			} else if handled {
				continue
			}
			return err
		}
	}
	return nil
}

// Refresh removes and re-registers one job according to its latest persisted state.
func (s *serviceImpl) Refresh(ctx context.Context, jobID int64) error {
	s.Remove(jobID)

	job, err := s.getJob(ctx, jobID)
	if err != nil {
		return err
	}
	if job == nil || jobmeta.NormalizeJobStatus(job.Status) != jobmeta.JobStatusEnabled {
		return nil
	}
	return s.registerJob(ctx, job)
}

// RegisterJobSnapshot removes and registers one provided job snapshot without
// reloading it from sys_job. Code-owned built-ins use this path after their
// declaration snapshot has been projected into sys_job for display and logs.
func (s *serviceImpl) RegisterJobSnapshot(ctx context.Context, job *entity.SysJob) error {
	if job == nil || job.Id == 0 {
		return nil
	}
	// Refresh the gcron entry from the declaration snapshot, not from sys_job.
	// This removes only the in-memory scheduler entry so changed cron metadata
	// takes effect and paused built-ins cannot keep running from a stale entry.
	s.Remove(job.Id)
	if jobmeta.NormalizeJobStatus(job.Status) != jobmeta.JobStatusEnabled {
		return nil
	}
	return s.registerJob(ctx, job)
}

// Remove unregisters one persistent job from gcron.
func (s *serviceImpl) Remove(jobID int64) {
	gcron.Remove(jobEntryName(jobID))
}

// jobEntryName builds the stable gcron entry name for one persistent job.
func jobEntryName(jobID int64) string {
	return fmt.Sprintf("scheduled-job:%d", jobID)
}

// normalizeGcronPattern converts stored cron expressions into the format accepted by gcron.
func normalizeGcronPattern(expr string) (string, error) {
	trimmedExpr := strings.TrimSpace(expr)
	if strings.HasPrefix(trimmedExpr, "@") {
		return trimmedExpr, nil
	}
	fields := strings.Fields(trimmedExpr)
	switch len(fields) {
	case 5:
		return "# " + strings.Join(fields, " "), nil
	case 6:
		return strings.Join(fields, " "), nil
	}
	return "", bizerr.NewCode(jobmeta.CodeJobCronFieldCountUnsupported)
}

// registerJob validates and registers one persistent job with gcron.
func (s *serviceImpl) registerJob(ctx context.Context, job *entity.SysJob) error {
	if job == nil {
		return nil
	}
	if err := s.validateExecutableJob(ctx, job); err != nil {
		return err
	}
	pattern, err := normalizeGcronPattern(job.CronExpr)
	if err != nil {
		return err
	}
	_, err = gcron.Add(context.Background(), pattern, func(jobCtx context.Context) {
		s.runCronJob(jobCtx, job.Id)
	}, jobEntryName(job.Id))
	return err
}

// handleLoadRegisterError downgrades enabled plugin-handler jobs to
// paused_by_plugin when the handler is no longer available during startup load.
func (s *serviceImpl) handleLoadRegisterError(
	ctx context.Context,
	job *entity.SysJob,
	registerErr error,
) (bool, error) {
	if job == nil || registerErr == nil {
		return false, nil
	}
	if job.IsBuiltin == 1 {
		return false, nil
	}
	if jobmeta.NormalizeTaskType(job.TaskType) != jobmeta.TaskTypeHandler {
		return false, nil
	}
	if !strings.HasPrefix(strings.TrimSpace(job.HandlerRef), "plugin:") {
		return false, nil
	}
	if _, ok := s.registry.Lookup(job.HandlerRef); ok {
		return false, nil
	}

	_, err := dao.SysJob.Ctx(ctx).
		Where(do.SysJob{
			Id:     job.Id,
			Status: string(jobmeta.JobStatusEnabled),
		}).
		Data(do.SysJob{
			Status:     string(jobmeta.JobStatusPausedByPlugin),
			StopReason: string(jobmeta.StopReasonPluginUnavailable),
		}).
		Update()
	if err != nil {
		return true, err
	}
	return true, nil
}

// listEnabledJobs queries all enabled persistent jobs.
func (s *serviceImpl) listEnabledJobs(ctx context.Context) ([]*entity.SysJob, error) {
	var jobs []*entity.SysJob
	err := dao.SysJob.Ctx(ctx).
		Where(do.SysJob{
			IsBuiltin: 0,
			Status:    string(jobmeta.JobStatusEnabled),
		}).
		Scan(&jobs)
	return jobs, err
}

// getJob queries one job by ID.
func (s *serviceImpl) getJob(ctx context.Context, jobID int64) (*entity.SysJob, error) {
	var job *entity.SysJob
	err := dao.SysJob.Ctx(ctx).
		Where(do.SysJob{Id: jobID}).
		Scan(&job)
	return job, err
}

// acquireSlot applies the per-job concurrency guard for cron-triggered executions.
func (s *serviceImpl) acquireSlot(job *entity.SysJob) (func(), jobmeta.LogStatus, error) {
	if job == nil {
		return func() {}, "", bizerr.NewCode(jobmeta.CodeJobNotFound)
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
	logID int64,
	jobID int64,
	cancel context.CancelFunc,
	release func(),
) {
	s.mu.Lock()
	s.runningInstances[logID] = &runningExecution{
		jobID:   jobID,
		cancel:  cancel,
		release: release,
	}
	s.mu.Unlock()
}

// finishRunningExecution removes one running instance and releases its slot.
func (s *serviceImpl) finishRunningExecution(logID int64) {
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
