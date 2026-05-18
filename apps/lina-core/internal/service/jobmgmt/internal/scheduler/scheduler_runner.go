// This file contains cron-trigger execution flow, quota handling, and result
// persistence for the persistent scheduled-job scheduler.

package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/jobhandler"
	"lina-core/internal/service/jobmeta"
	"lina-core/internal/service/jobmgmt/internal/shellexec"
	"lina-core/pkg/logger"
)

// runCronJob handles one gcron callback for the target persistent job.
func (s *serviceImpl) runCronJob(ctx context.Context, jobID int64) {
	job, err := s.getJob(ctx, jobID)
	if err != nil {
		logger.Warningf(ctx, "load scheduled job failed job_id=%d err=%v", jobID, err)
		return
	}
	if job == nil || jobmeta.NormalizeJobStatus(job.Status) != jobmeta.JobStatusEnabled {
		return
	}

	if jobmeta.NormalizeJobScope(job.Scope) == jobmeta.JobScopeMasterOnly && !s.isPrimary() {
		if err = s.createTerminalLog(
			ctx,
			job,
			jobmeta.TriggerTypeCron,
			jobmeta.LogStatusSkippedNotPrimary,
			"Current node is not primary, skipped execution",
		); err != nil {
			logger.Warningf(ctx, "create not-primary job log failed job_id=%d err=%v", jobID, err)
		}
		return
	}

	release, skippedStatus, err := s.acquireSlot(job)
	if err != nil {
		logger.Warningf(ctx, "acquire job slot failed job_id=%d err=%v", jobID, err)
		return
	}
	if skippedStatus.IsValid() {
		if err = s.createTerminalLog(
			ctx,
			job,
			jobmeta.TriggerTypeCron,
			skippedStatus,
			"Current scheduled job has reached its concurrency limit, skipped execution",
		); err != nil {
			logger.Warningf(ctx, "create skipped-concurrency log failed job_id=%d err=%v", jobID, err)
		}
		return
	}

	job, shouldRemove, err := s.prepareCronQuota(ctx, job)
	if err != nil {
		release()
		logger.Warningf(ctx, "prepare cron quota failed job_id=%d err=%v", jobID, err)
		return
	}
	if job == nil {
		release()
		if shouldRemove {
			s.Remove(jobID)
		}
		return
	}

	logID, execution, err := s.createExecution(ctx, job, jobmeta.TriggerTypeCron)
	if err != nil {
		release()
		logger.Warningf(ctx, "create running job log failed job_id=%d err=%v", jobID, err)
		return
	}
	s.storeRunningExecution(logID, jobID, execution.cancel, release)
	if shouldRemove {
		s.Remove(jobID)
	}

	go s.executeJob(execution, job, logID)
}

// prepareCronQuota increments cron execution counters and disables exhausted jobs.
func (s *serviceImpl) prepareCronQuota(
	ctx context.Context,
	job *entity.SysJob,
) (*entity.SysJob, bool, error) {
	if job == nil {
		return nil, false, nil
	}

	currentCount := job.ExecutedCount
	if job.MaxExecutions > 0 && currentCount >= int64(job.MaxExecutions) {
		removed, err := s.disableForMaxExecutions(ctx, job.Id)
		return nil, removed, err
	}

	cols := dao.SysJob.Columns()
	data := do.SysJob{
		ExecutedCount: gdb.Raw(cols.ExecutedCount + " + 1"),
		StopReason:    "",
	}
	nextCount := currentCount + 1
	shouldRemove := false
	if job.MaxExecutions > 0 && nextCount >= int64(job.MaxExecutions) {
		data.Status = string(jobmeta.JobStatusDisabled)
		data.StopReason = string(jobmeta.StopReasonMaxExecutionsReached)
		shouldRemove = true
	}

	result, err := dao.SysJob.Ctx(ctx).
		Where(do.SysJob{
			Id:            job.Id,
			Status:        string(jobmeta.JobStatusEnabled),
			ExecutedCount: currentCount,
		}).
		Data(data).
		Update()
	if err != nil {
		return nil, false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, false, err
	}
	if affected == 0 {
		return nil, false, nil
	}

	job.ExecutedCount = nextCount
	if shouldRemove {
		job.Status = string(jobmeta.JobStatusDisabled)
		job.StopReason = string(jobmeta.StopReasonMaxExecutionsReached)
	}
	return job, shouldRemove, nil
}

// disableForMaxExecutions disables one exhausted job and returns whether the scheduler should unregister it.
func (s *serviceImpl) disableForMaxExecutions(ctx context.Context, jobID int64) (bool, error) {
	result, err := dao.SysJob.Ctx(ctx).
		Where(do.SysJob{Id: jobID, Status: string(jobmeta.JobStatusEnabled)}).
		Data(do.SysJob{
			Status:     string(jobmeta.JobStatusDisabled),
			StopReason: string(jobmeta.StopReasonMaxExecutionsReached),
		}).
		Update()
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

// executeJob dispatches one running job instance and persists its final log state.
func (s *serviceImpl) executeJob(
	execution executionState,
	job *entity.SysJob,
	logID int64,
) {
	defer s.finishRunningExecution(logID)

	status, errMsg, resultJSON := s.dispatchExecution(execution.ctx, job)
	// Persist the terminal log state with a detached context so timeout or
	// cancellation does not prevent the final status from being recorded.
	finishCtx := context.WithoutCancel(execution.ctx)
	if err := s.finishLog(finishCtx, logID, execution.startedAt, status, errMsg, resultJSON); err != nil {
		logger.Warningf(execution.ctx, "finish scheduled job log failed log_id=%d err=%v", logID, err)
	}
}

// dispatchExecution runs one handler or shell task and maps the result to job-log fields.
func (s *serviceImpl) dispatchExecution(
	ctx context.Context,
	job *entity.SysJob,
) (jobmeta.LogStatus, string, string) {
	switch jobmeta.NormalizeTaskType(job.TaskType) {
	case jobmeta.TaskTypeHandler:
		return s.dispatchHandlerExecution(ctx, job)
	case jobmeta.TaskTypeShell:
		return s.dispatchShellExecution(ctx, job)
	default:
		return jobmeta.LogStatusFailed, "Scheduled-job task type is not supported", ""
	}
}

// dispatchHandlerExecution runs one registered handler callback.
func (s *serviceImpl) dispatchHandlerExecution(
	ctx context.Context,
	job *entity.SysJob,
) (jobmeta.LogStatus, string, string) {
	def, ok := s.registry.Lookup(job.HandlerRef)
	if !ok {
		return jobmeta.LogStatusFailed, "Scheduled-job handler does not exist", ""
	}
	if err := jobhandler.ValidateParams(def.ParamsSchema, json.RawMessage(job.Params)); err != nil {
		return jobmeta.LogStatusFailed, err.Error(), ""
	}

	result, err := def.Invoke(ctx, json.RawMessage(job.Params))
	if err != nil {
		return mapContextErrorToLogStatus(err, ctx), buildExecutionErrMsg(
			err,
			ctx,
			time.Duration(job.TimeoutSeconds)*time.Second,
		), marshalResultJSON(nil)
	}
	return jobmeta.LogStatusSuccess, "", marshalResultJSON(result)
}

// dispatchShellExecution runs one shell task through the guarded shell executor.
func (s *serviceImpl) dispatchShellExecution(
	ctx context.Context,
	job *entity.SysJob,
) (jobmeta.LogStatus, string, string) {
	if s.shellExecutor == nil {
		return jobmeta.LogStatusFailed, "Shell executor is not initialized", ""
	}

	envMap := make(map[string]string)
	if strings.TrimSpace(job.Env) != "" {
		if err := json.Unmarshal([]byte(job.Env), &envMap); err != nil {
			return jobmeta.LogStatusFailed, "Failed to parse Shell environment variables", ""
		}
	}

	output, err := s.shellExecutor.Execute(ctx, shellexec.ExecuteInput{
		ShellCmd: job.ShellCmd,
		WorkDir:  job.WorkDir,
		Env:      envMap,
		Timeout:  time.Duration(job.TimeoutSeconds) * time.Second,
	})
	resultJSON := marshalResultJSON(buildShellResult(output))
	if err != nil {
		return mapContextErrorToLogStatus(err, ctx), buildExecutionErrMsg(
			err,
			ctx,
			time.Duration(job.TimeoutSeconds)*time.Second,
		), resultJSON
	}
	return jobmeta.LogStatusSuccess, "", resultJSON
}

// buildShellResult converts one executor output snapshot to the stable
// jobmeta.ShellResult payload persisted in execution logs.
func buildShellResult(output *shellexec.ExecuteOutput) *jobmeta.ShellResult {
	if output == nil {
		return nil
	}
	return &jobmeta.ShellResult{
		Stdout:    output.Stdout,
		Stderr:    output.Stderr,
		ExitCode:  output.ExitCode,
		Cancelled: output.Cancelled,
		TimedOut:  output.TimedOut,
	}
}

// marshalResultJSON serializes one optional execution result payload.
func marshalResultJSON(result any) string {
	if result == nil {
		return ""
	}
	data, err := json.Marshal(result)
	if err != nil {
		return ""
	}
	return string(data)
}

// mapContextErrorToLogStatus maps cancellation and timeout failures to specific log statuses.
func mapContextErrorToLogStatus(err error, ctx context.Context) jobmeta.LogStatus {
	if err == nil {
		return jobmeta.LogStatusSuccess
	}
	if errors.Is(err, context.Canceled) || ctx.Err() == context.Canceled {
		return jobmeta.LogStatusCancelled
	}
	if errors.Is(err, context.DeadlineExceeded) || ctx.Err() == context.DeadlineExceeded {
		return jobmeta.LogStatusTimeout
	}
	return jobmeta.LogStatusFailed
}

// buildExecutionErrMsg normalizes timeout and cancellation failures into
// stable, user-facing execution-log messages.
func buildExecutionErrMsg(err error, ctx context.Context, timeout time.Duration) string {
	if err == nil {
		return ""
	}
	if mapContextErrorToLogStatus(err, ctx) != jobmeta.LogStatusTimeout {
		return err.Error()
	}

	message := "Scheduled-job execution timed out"
	if timeout > 0 {
		message += " (" + timeout.String() + ")"
	}
	return message + ": " + err.Error()
}
