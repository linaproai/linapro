// Package scheduler implements persistent scheduled-job registration and
// execution on top of GoFrame's gcron runner.
package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/os/gcron"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/cluster"
	"lina-core/internal/service/config"
	"lina-core/internal/service/jobhandler"
	"lina-core/internal/service/jobmeta"
	"lina-core/internal/service/jobmgmt/internal/shellexec"
	"lina-core/internal/service/startupstats"
	"lina-core/pkg/bizerr"
)

// Scheduler defines the persistent scheduled-job runner contract.
type Scheduler interface {
	// LoadAndRegister registers all currently enabled persistent jobs at startup.
	LoadAndRegister(ctx context.Context) error
	// Refresh removes and re-registers one job according to its latest persisted state.
	Refresh(ctx context.Context, jobID int64) error
	// RegisterJobSnapshot removes and registers one provided job snapshot without
	// reloading it from sys_job.
	RegisterJobSnapshot(ctx context.Context, job *entity.SysJob) error
	// Remove unregisters one persistent job from gcron.
	Remove(jobID int64)
	// Trigger starts one manual execution and returns the created log ID.
	Trigger(ctx context.Context, jobID int64) (int64, error)
	// CancelLog cancels one currently running job-log instance.
	CancelLog(ctx context.Context, logID int64) error
}

// runningExecution stores one cancellable execution instance.
type runningExecution struct {
	jobID   int64              // jobID identifies the owning job definition.
	cancel  context.CancelFunc // cancel stops the execution context.
	release func()             // release decrements concurrency bookkeeping.
}

// serviceImpl implements Scheduler.
type serviceImpl struct {
	clusterSvc    cluster.Service     // clusterSvc exposes primary-node and node-ID state.
	registry      jobhandler.Registry // registry resolves handler callbacks.
	shellExecutor shellexec.Executor  // shellExecutor runs shell-type jobs.

	mu               sync.Mutex                  // mu protects running instance bookkeeping.
	runningCounts    map[int64]int               // runningCounts tracks concurrent in-flight runs per job.
	runningInstances map[int64]*runningExecution // runningInstances tracks cancellable log instances.
}

// Ensure serviceImpl implements Scheduler.
var _ Scheduler = (*serviceImpl)(nil)

var instance Scheduler
var once sync.Once

// Instance returns the singleton scheduler instance.
// It initializes the instance exactly once, using default dependencies.
func Instance() Scheduler {
	once.Do(func() {
		instance = &serviceImpl{
			clusterSvc:       cluster.Instance(),
			registry:         jobhandler.Instance(),
			shellExecutor:    shellexec.New(config.Instance()),
			runningCounts:    make(map[int64]int),
			runningInstances: make(map[int64]*runningExecution),
		}
	})
	return instance
}

// New creates and returns one persistent scheduler.
// Deprecated: Use Instance() for singleton access.
func New(
	clusterSvc cluster.Service,
	registry jobhandler.Registry,
	shellExecutor shellexec.Executor,
) Scheduler {
	return &serviceImpl{
		clusterSvc:       clusterSvc,
		registry:         registry,
		shellExecutor:    shellExecutor,
		runningCounts:    make(map[int64]int),
		runningInstances: make(map[int64]*runningExecution),
	}
}

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

// executionState keeps one execution context together with its cancellation and start time.
type executionState struct {
	ctx       context.Context    // ctx is passed into the actual task execution.
	cancel    context.CancelFunc // cancel aborts the running execution.
	startedAt time.Time          // startedAt records when the execution log began.
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
