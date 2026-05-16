// This file contains execution validation, execution context creation, and
// job-log persistence helpers for scheduled-job runs. It keeps database log
// updates together so registration and runner flow files stay focused on their
// own lifecycle responsibilities.

package scheduler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/jobhandler"
	"lina-core/internal/service/jobmeta"
	"lina-core/pkg/bizerr"
)

// Trigger starts one manual execution and returns the created log ID.
func (s *serviceImpl) Trigger(ctx context.Context, jobID int64) (int64, error) {
	job, err := s.getJob(ctx, jobID)
	if err != nil {
		return 0, err
	}
	if job == nil {
		return 0, bizerr.NewCode(jobmeta.CodeJobNotFound)
	}
	if jobmeta.NormalizeJobStatus(job.Status) == jobmeta.JobStatusPausedByPlugin {
		return 0, bizerr.NewCode(jobmeta.CodeJobHandlerUnavailable)
	}
	if err = s.validateExecutableJob(ctx, job); err != nil {
		return 0, err
	}

	logID, execution, err := s.createExecution(ctx, job, jobmeta.TriggerTypeManual)
	if err != nil {
		return 0, err
	}
	s.storeRunningExecution(logID, job.Id, execution.cancel, func() {})

	go s.executeJob(execution, job, logID)
	return logID, nil
}

// executionState keeps one execution context together with its cancellation and start time.
type executionState struct {
	ctx       context.Context    // ctx is passed into the actual task execution.
	cancel    context.CancelFunc // cancel aborts the running execution.
	startedAt time.Time          // startedAt records when the execution log began.
}

// createExecution inserts one running log and returns the derived execution context.
func (s *serviceImpl) createExecution(
	ctx context.Context,
	job *entity.SysJob,
	trigger jobmeta.TriggerType,
) (int64, executionState, error) {
	if job == nil {
		return 0, executionState{}, bizerr.NewCode(jobmeta.CodeJobNotFound)
	}

	startAt := time.Now()
	logID, err := s.createRunningLog(ctx, job, trigger, startAt)
	if err != nil {
		return 0, executionState{}, err
	}

	execCtx, cancel := context.WithTimeout(
		context.WithoutCancel(ctx),
		time.Duration(job.TimeoutSeconds)*time.Second,
	)
	return logID, executionState{
		ctx:       execCtx,
		cancel:    cancel,
		startedAt: startAt,
	}, nil
}

// validateExecutableJob validates the runtime prerequisites for one job definition.
func (s *serviceImpl) validateExecutableJob(ctx context.Context, job *entity.SysJob) error {
	if job == nil {
		return bizerr.NewCode(jobmeta.CodeJobNotFound)
	}
	switch jobmeta.NormalizeTaskType(job.TaskType) {
	case jobmeta.TaskTypeHandler:
		def, ok := s.registry.Lookup(job.HandlerRef)
		if !ok {
			return bizerr.NewCode(jobhandler.CodeJobHandlerNotFound)
		}
		return jobhandler.ValidateParams(def.ParamsSchema, json.RawMessage(job.Params))

	case jobmeta.TaskTypeShell:
		if s.shellExecutor == nil {
			return bizerr.NewCode(jobmeta.CodeJobShellExecutorUninitialized)
		}
		return nil
	}
	return bizerr.NewCode(jobmeta.CodeJobTaskTypeUnsupported)
}

// nodeID returns the stable execution node identifier.
func (s *serviceImpl) nodeID() string {
	if s == nil || s.clusterSvc == nil {
		return "local-node"
	}
	return s.clusterSvc.NodeID()
}

// isPrimary reports whether the current node should execute primary-only jobs.
func (s *serviceImpl) isPrimary() bool {
	if s == nil || s.clusterSvc == nil {
		return true
	}
	return s.clusterSvc.IsPrimary()
}

// createRunningLog inserts one running log row and returns its identifier.
func (s *serviceImpl) createRunningLog(
	ctx context.Context,
	job *entity.SysJob,
	trigger jobmeta.TriggerType,
	startedAt time.Time,
) (int64, error) {
	snapshot, err := json.Marshal(job)
	if err != nil {
		return 0, bizerr.WrapCode(err, jobmeta.CodeJobSnapshotMarshalFailed)
	}
	paramsSnapshot := ""
	if jobmeta.NormalizeTaskType(job.TaskType) == jobmeta.TaskTypeHandler {
		paramsSnapshot = job.Params
	}

	insertID, err := dao.SysJobLog.Ctx(ctx).Data(do.SysJobLog{
		JobId:          job.Id,
		JobSnapshot:    string(snapshot),
		NodeId:         s.nodeID(),
		Trigger:        string(trigger),
		ParamsSnapshot: paramsSnapshot,
		StartAt:        gtime.New(startedAt),
		Status:         string(jobmeta.LogStatusRunning),
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}
	return gconv.Int64(insertID), nil
}

// createTerminalLog inserts one already-finished log row for skipped executions.
func (s *serviceImpl) createTerminalLog(
	ctx context.Context,
	job *entity.SysJob,
	trigger jobmeta.TriggerType,
	status jobmeta.LogStatus,
	errMsg string,
) error {
	now := time.Now()
	snapshot, err := json.Marshal(job)
	if err != nil {
		return err
	}
	paramsSnapshot := ""
	if jobmeta.NormalizeTaskType(job.TaskType) == jobmeta.TaskTypeHandler {
		paramsSnapshot = job.Params
	}

	_, err = dao.SysJobLog.Ctx(ctx).Data(do.SysJobLog{
		JobId:          job.Id,
		JobSnapshot:    string(snapshot),
		NodeId:         s.nodeID(),
		Trigger:        string(trigger),
		ParamsSnapshot: paramsSnapshot,
		StartAt:        gtime.New(now),
		EndAt:          gtime.New(now),
		DurationMs:     0,
		Status:         string(status),
		ErrMsg:         errMsg,
	}).Insert()
	return err
}

// finishLog updates one running log row with its terminal result snapshot.
func (s *serviceImpl) finishLog(
	ctx context.Context,
	logID int64,
	startedAt time.Time,
	status jobmeta.LogStatus,
	errMsg string,
	resultJSON string,
) error {
	_, err := dao.SysJobLog.Ctx(ctx).
		Where(do.SysJobLog{Id: logID}).
		Data(do.SysJobLog{
			EndAt:      gtime.New(time.Now()),
			DurationMs: time.Since(startedAt).Milliseconds(),
			Status:     string(status),
			ErrMsg:     errMsg,
			ResultJson: resultJSON,
		}).
		Update()
	return err
}
