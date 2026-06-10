// This file verifies declaration-driven registration for host and plugin
// managed scheduled jobs.

package cron

import (
	"context"
	"testing"
	"time"

	"lina-core/internal/model/entity"
	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/jobhandler"
	"lina-core/internal/service/jobmeta"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
	pluginsvc "lina-core/internal/service/plugin"
)

// managedJobSyncerStub captures built-in reconciliation inputs and returns
// sys_job projection IDs without reading them back from persistence.
type managedJobSyncerStub struct {
	received []jobmgmtsvc.BuiltinJobDef
}

// ReconcileBuiltinJobs captures declarations and returns matching projections.
func (s *managedJobSyncerStub) ReconcileBuiltinJobs(
	ctx context.Context,
	jobs []jobmgmtsvc.BuiltinJobDef,
) ([]*entity.SysJob, error) {
	s.received = append([]jobmgmtsvc.BuiltinJobDef(nil), jobs...)

	projections := make([]*entity.SysJob, 0, len(jobs))
	for index, job := range jobs {
		projections = append(projections, &entity.SysJob{
			Id:             int64(1000 + index),
			Name:           job.Name,
			Description:    job.Description,
			TaskType:       string(job.TaskType),
			HandlerRef:     job.HandlerRef,
			CronExpr:       job.Pattern,
			Timezone:       job.Timezone,
			Scope:          string(job.Scope),
			Concurrency:    string(job.Concurrency),
			MaxConcurrency: job.MaxConcurrency,
			MaxExecutions:  job.MaxExecutions,
			Status:         string(jobmeta.JobStatusEnabled),
			IsBuiltin:      1,
		})
	}
	return projections, nil
}

// managedJobSchedulerStub captures declaration-derived snapshots registered by
// the cron service.
type managedJobSchedulerStub struct {
	snapshots []*entity.SysJob
}

// LoadAndRegister is unused by declaration registration tests.
func (s *managedJobSchedulerStub) LoadAndRegister(ctx context.Context) error { return nil }

// Refresh is unused by declaration registration tests.
func (s *managedJobSchedulerStub) Refresh(ctx context.Context, jobID int64) error { return nil }

// RegisterJobSnapshot records one declaration-derived registration snapshot.
func (s *managedJobSchedulerStub) RegisterJobSnapshot(ctx context.Context, job *entity.SysJob) error {
	s.snapshots = append(s.snapshots, job)
	return nil
}

// Remove is unused by declaration registration tests.
func (s *managedJobSchedulerStub) Remove(jobID int64) {}

// Trigger is unused by declaration registration tests.
func (s *managedJobSchedulerStub) Trigger(ctx context.Context, jobID int64) (int64, error) {
	return 0, nil
}

// CancelLog is unused by declaration registration tests.
func (s *managedJobSchedulerStub) CancelLog(ctx context.Context, logID int64) error { return nil }

// managedPluginCronStub records which plugin cron surface the cron service
// consumes while building code-owned scheduled-job projections.
type managedPluginCronStub struct {
	listExecutableCalled        bool
	listInstalledDeclaredCalled bool
	installedDeclarations       []pluginsvc.ManagedJob
}

// managedJobConfigStub overrides runtime settings while inheriting the rest
// of the host config service contract from a real service instance.
type managedJobConfigStub struct {
	hostconfig.Service
	sessionTimeout  time.Duration
	logRetentionDay int64
}

// GetSessionTimeout returns the runtime-effective session timeout for tests.
func (s managedJobConfigStub) GetSessionTimeout(context.Context) (time.Duration, error) {
	return s.sessionTimeout, nil
}

// GetLogRetentionDays returns the runtime-effective log retention in days for tests.
func (s managedJobConfigStub) GetLogRetentionDays(context.Context) (int64, error) {
	return s.logRetentionDay, nil
}

// ListExecutableJobs records unexpected executable-list usage.
func (s *managedPluginCronStub) ListExecutableJobs(ctx context.Context) ([]pluginsvc.ManagedJob, error) {
	s.listExecutableCalled = true
	return nil, nil
}

// ListInstalledJobDeclarations returns installed plugin declaration snapshots.
func (s *managedPluginCronStub) ListInstalledJobDeclarations(ctx context.Context) ([]pluginsvc.ManagedJob, error) {
	s.listInstalledDeclaredCalled = true
	return s.installedDeclarations, nil
}

// TestSyncBuiltinScheduledJobsRegistersDeclarationSnapshots verifies cron
// registers built-ins from the reconciliation return value rather than a later
// persistent scheduler scan of sys_job.
func TestSyncBuiltinScheduledJobsRegistersDeclarationSnapshots(t *testing.T) {
	ctx := context.Background()
	syncer := &managedJobSyncerStub{}
	scheduler := &managedJobSchedulerStub{}
	svc := &serviceImpl{
		configSvc:           hostconfig.New(),
		registry:            jobhandler.New(),
		builtinSyncer:       syncer,
		persistentScheduler: scheduler,
	}

	if err := svc.syncBuiltinScheduledJobs(ctx); err != nil {
		t.Fatalf("expected builtin sync to succeed, got error: %v", err)
	}
	if len(syncer.received) == 0 {
		t.Fatal("expected host built-in declarations to be reconciled")
	}
	if len(scheduler.snapshots) != len(syncer.received) {
		t.Fatalf("expected one scheduler snapshot per declaration, got %d want %d", len(scheduler.snapshots), len(syncer.received))
	}
	for index, declaration := range syncer.received {
		snapshot := scheduler.snapshots[index]
		if snapshot == nil {
			t.Fatalf("expected scheduler snapshot %d to be present", index)
		}
		if snapshot.Id == 0 {
			t.Fatalf("expected scheduler snapshot %d to carry a sys_job projection ID", index)
		}
		if snapshot.HandlerRef != declaration.HandlerRef {
			t.Fatalf("expected snapshot handler_ref=%s, got %s", declaration.HandlerRef, snapshot.HandlerRef)
		}
	}
}

// TestBuildPluginBuiltinJobsUsesInstalledDeclarations verifies disabled but
// installed plugin job declarations remain visible to scheduled-job management.
func TestBuildPluginBuiltinJobsUsesInstalledDeclarations(t *testing.T) {
	ctx := context.Background()
	pluginSvc := &managedPluginCronStub{
		installedDeclarations: []pluginsvc.ManagedJob{
			{
				PluginID:       "plugin-jobs-installed",
				Name:           "heartbeat",
				DisplayName:    "Plugin Heartbeat",
				Description:    "Installed plugin heartbeat.",
				Pattern:        "# */10 * * * *",
				Timezone:       "Asia/Shanghai",
				Scope:          jobmeta.JobScopeAllNode,
				Concurrency:    jobmeta.JobConcurrencySingleton,
				MaxConcurrency: 1,
			},
		},
	}
	svc := &serviceImpl{
		configSvc: hostconfig.New(),
		pluginSvc: pluginSvc,
	}

	jobs, err := svc.buildPluginBuiltinJobs(ctx)
	if err != nil {
		t.Fatalf("expected plugin builtin projection to succeed, got error: %v", err)
	}
	if pluginSvc.listExecutableCalled {
		t.Fatal("expected plugin builtin projection not to use executable cron list")
	}
	if !pluginSvc.listInstalledDeclaredCalled {
		t.Fatal("expected plugin builtin projection to use installed declaration cron list")
	}
	if len(jobs) != 1 {
		t.Fatalf("expected one plugin builtin job, got %#v", jobs)
	}
	if jobs[0].HandlerRef != "plugin:plugin-jobs-installed/jobs:heartbeat" {
		t.Fatalf("unexpected plugin handler ref: %s", jobs[0].HandlerRef)
	}
}

// TestEffectiveSessionCleanupTimeoutUsesRuntimeSessionTimeout verifies online
// session cleanup honors the runtime session timeout before retention boundaries.
func TestEffectiveSessionCleanupTimeoutUsesRuntimeSessionTimeout(t *testing.T) {
	svc := &serviceImpl{
		sessionCfg: &hostconfig.SessionConfig{Timeout: 30 * 24 * time.Hour},
		configSvc: managedJobConfigStub{
			Service:         hostconfig.New(),
			sessionTimeout:  6 * time.Hour,
			logRetentionDay: 90,
		},
	}

	timeout, err := svc.effectiveSessionCleanupTimeout(context.Background())
	if err != nil {
		t.Fatalf("resolve effective session cleanup timeout: %v", err)
	}
	if timeout != 6*time.Hour {
		t.Fatalf("expected runtime session timeout 6h, got %s", timeout)
	}
}

// TestEffectiveSessionCleanupTimeoutUsesStricterLogRetention verifies the
// global log-retention maximum can tighten online-session cleanup.
func TestEffectiveSessionCleanupTimeoutUsesStricterLogRetention(t *testing.T) {
	svc := &serviceImpl{
		sessionCfg: &hostconfig.SessionConfig{Timeout: 30 * 24 * time.Hour},
		configSvc: managedJobConfigStub{
			Service:         hostconfig.New(),
			sessionTimeout:  20 * 24 * time.Hour,
			logRetentionDay: 7,
		},
	}

	timeout, err := svc.effectiveSessionCleanupTimeout(context.Background())
	if err != nil {
		t.Fatalf("resolve effective session cleanup timeout: %v", err)
	}
	if timeout != 7*24*time.Hour {
		t.Fatalf("expected stricter log retention timeout 168h, got %s", timeout)
	}
}
