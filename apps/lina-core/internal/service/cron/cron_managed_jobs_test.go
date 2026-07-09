// This file verifies declaration-driven registration for host and plugin
// managed scheduled jobs.

package cron

import (
	"context"
	"errors"
	jobv1 "lina-core/api/job/v1"
	"testing"
	"time"

	"lina-core/internal/model/entity"
	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/jobhandler"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
	pluginsvc "lina-core/internal/service/plugin"
)

// managedJobSyncerStub captures built-in reconciliation inputs and returns
// sys_job projection IDs without reading them back from persistence.
type managedJobSyncerStub struct {
	jobmgmtsvc.Service
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
			Status:         string(jobv1.StatusEnabled),
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
	pluginsvc.Service
	listExecutableCalled        bool
	listInstalledDeclaredCalled bool
	installedDeclarations       []pluginsvc.ManagedJob
}

// retryRegistryStub fails one registration attempt before accepting handlers.
type retryRegistryStub struct {
	registerErr error
	registered  []jobhandler.HandlerDef
}

// Register records handlers after the configured transient error is consumed.
func (s *retryRegistryStub) Register(def jobhandler.HandlerDef) error {
	if s.registerErr != nil {
		err := s.registerErr
		s.registerErr = nil
		return err
	}
	s.registered = append(s.registered, def)
	return nil
}

// Unregister is unused by managed-handler retry tests.
func (s *retryRegistryStub) Unregister(ref string) {}

// Lookup is unused by managed-handler retry tests.
func (s *retryRegistryStub) Lookup(ref string) (jobhandler.HandlerDef, bool) {
	return jobhandler.HandlerDef{}, false
}

// List is unused by managed-handler retry tests.
func (s *retryRegistryStub) List() []jobhandler.HandlerInfo { return nil }

// SubscribeChanges is unused by managed-handler retry tests.
func (s *retryRegistryStub) SubscribeChanges(callback jobhandler.ChangeCallback) func() {
	return func() {}
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

// ListManagedJobs returns installed plugin declaration snapshots.
func (s *managedPluginCronStub) ListManagedJobs(ctx context.Context, query pluginsvc.ManagedJobQuery) ([]pluginsvc.ManagedJob, error) {
	if query.ExecutableOnly {
		s.listExecutableCalled = true
		return nil, nil
	}
	if query.InstalledOnly {
		s.listInstalledDeclaredCalled = true
	}
	return s.installedDeclarations, nil
}

// TestSyncBuiltinScheduledJobsRegistersDeclarationSnapshots verifies cron
// registers built-ins from the reconciliation return value rather than a later
// persistent scheduler scan of sys_job.
func TestSyncBuiltinScheduledJobsRegistersDeclarationSnapshots(t *testing.T) {
	var (
		ctx       = context.Background()
		syncer    = &managedJobSyncerStub{}
		scheduler = &managedJobSchedulerStub{}
	)
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

// TestManagedHostJobsDoNotProjectKVCacheCleanup verifies built-in scheduled
// jobs no longer expose SQL table cache cleanup now that supported backends
// expire keys natively.
func TestManagedHostJobsDoNotProjectKVCacheCleanup(t *testing.T) {
	ctx := context.Background()
	registry := jobhandler.New()
	svc := &serviceImpl{
		configSvc: hostconfig.New(),
		registry:  registry,
	}

	if err := svc.ensureManagedHandlersRegistered(); err != nil {
		t.Fatalf("register managed handlers: %v", err)
	}
	if _, ok := registry.Lookup("host:kvcache-cleanup-expired"); ok {
		t.Fatal("expected KV cache cleanup handler to be absent")
	}
	jobs, err := svc.buildHostBuiltinJobs(ctx)
	if err != nil {
		t.Fatalf("build host builtin jobs: %v", err)
	}
	for _, job := range jobs {
		if job.HandlerRef == "host:kvcache-cleanup-expired" {
			t.Fatalf("expected no KV cache cleanup job projection, got %#v", job)
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
				Scope:          jobv1.ScopeAllNode,
				Concurrency:    jobv1.ConcurrencySingleton,
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

// TestEnsureManagedHandlersRegisteredRetriesAfterFailure verifies transient
// registry failures are not hidden by one-shot initialization state.
func TestEnsureManagedHandlersRegisteredRetriesAfterFailure(t *testing.T) {
	registry := &retryRegistryStub{registerErr: errors.New("temporary registry failure")}
	svc := &serviceImpl{registry: registry}

	if err := svc.ensureManagedHandlersRegistered(); err == nil {
		t.Fatal("expected first managed handler registration to fail")
	}
	if len(registry.registered) != 0 {
		t.Fatalf("expected failed registration not to record handlers, got %#v", registry.registered)
	}

	if err := svc.ensureManagedHandlersRegistered(); err != nil {
		t.Fatalf("expected managed handler registration retry to succeed, got error: %v", err)
	}
	if len(registry.registered) != 1 {
		t.Fatalf("expected retry to register one host handler, got %d", len(registry.registered))
	}
	if registry.registered[0].Ref != "host:session-cleanup" {
		t.Fatalf("unexpected registered handler ref: %s", registry.registered[0].Ref)
	}
}

// TestEffectiveSessionCleanupTimeoutUsesRuntimeSessionTimeout verifies online
// session cleanup honors the runtime session timeout before retention boundaries.
func TestEffectiveSessionCleanupTimeoutUsesRuntimeSessionTimeout(t *testing.T) {
	svc := &serviceImpl{
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
