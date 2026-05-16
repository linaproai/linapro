// Package scheduler implements persistent scheduled-job registration and
// execution on top of GoFrame's gcron runner.
package scheduler

import (
	"context"
	"sync"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/cluster"
	"lina-core/internal/service/jobhandler"
	"lina-core/internal/service/jobmgmt/internal/shellexec"
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

// New creates and returns one persistent scheduler.
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
