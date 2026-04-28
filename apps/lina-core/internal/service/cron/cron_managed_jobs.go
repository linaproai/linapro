// This file projects host and plugin code-owned scheduled jobs into sys_job
// and publishes the host-side handler callbacks they execute through.

package cron

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	jobhandlersvc "lina-core/internal/service/jobhandler"
	"lina-core/internal/service/jobmeta"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/pluginbridge"
)

const (
	defaultManagedJobTimezone = "Asia/Shanghai"
	defaultManagedJobTimeout  = 5 * time.Minute
)

// syncBuiltinScheduledJobs ensures code-owned host and plugin jobs are synced
// into sys_job before the persistent scheduler loads enabled rows.
func (s *serviceImpl) syncBuiltinScheduledJobs(ctx context.Context) error {
	if s == nil || s.builtinSyncer == nil {
		return nil
	}
	if err := s.ensureManagedHandlersRegistered(); err != nil {
		return err
	}

	jobs := s.buildHostBuiltinJobs()
	pluginJobs, err := s.buildPluginBuiltinJobs(ctx)
	if err != nil {
		return err
	}
	jobs = append(jobs, pluginJobs...)
	return s.builtinSyncer.ReconcileBuiltinJobs(ctx, jobs)
}

// ensureManagedHandlersRegistered registers host-owned handlers exactly once so
// projected sys_job rows always resolve through the shared handler registry.
func (s *serviceImpl) ensureManagedHandlersRegistered() error {
	if s == nil || s.registry == nil {
		return nil
	}

	var registerErr error
	s.managedHandlersOnce.Do(func() {
		registerErr = s.registerManagedHandlers()
	})
	return registerErr
}

// registerManagedHandlers publishes the host-owned built-in scheduled-job callbacks.
func (s *serviceImpl) registerManagedHandlers() error {
	handlers := []jobhandlersvc.HandlerDef{
		{
			Ref:          "host:session-cleanup",
			DisplayName:  "Online Session Cleanup",
			Description:  "Cleans up inactive online sessions in the host according to the session-timeout policy.",
			ParamsSchema: `{"type":"object","properties":{}}`,
			Source:       jobmeta.HandlerSourceHost,
			Invoke:       s.invokeSessionCleanup,
		},
	}

	if s.clusterSvc != nil && s.clusterSvc.IsEnabled() {
		handlers = append(handlers,
			jobhandlersvc.HandlerDef{
				Ref:          "host:access-topology-sync",
				DisplayName:  "Access Topology Sync",
				Description:  "Synchronizes permission-topology revision snapshots across the cluster so authorization caches stay consistent on every node.",
				ParamsSchema: `{"type":"object","properties":{}}`,
				Source:       jobmeta.HandlerSourceHost,
				Invoke:       s.invokeAccessTopologySync,
			},
			jobhandlersvc.HandlerDef{
				Ref:          "host:runtime-param-sync",
				DisplayName:  "Runtime Parameter Sync",
				Description:  "Synchronizes protected runtime parameter snapshots across the cluster so each node keeps a fresh local cache.",
				ParamsSchema: `{"type":"object","properties":{}}`,
				Source:       jobmeta.HandlerSourceHost,
				Invoke:       s.invokeRuntimeParamSync,
			},
		)
	}

	for _, definition := range handlers {
		if err := s.registry.Register(definition); err != nil && !isDuplicateHandlerError(err) {
			return err
		}
	}
	return nil
}

// buildHostBuiltinJobs returns host-owned scheduled-job definitions that should
// always appear in unified scheduled-job management.
func (s *serviceImpl) buildHostBuiltinJobs() []jobmgmtsvc.BuiltinJobDef {
	if s == nil {
		return nil
	}

	sessionCleanupInterval := 5 * time.Minute
	if s.sessionCfg != nil && s.sessionCfg.CleanupInterval > 0 {
		sessionCleanupInterval = s.sessionCfg.CleanupInterval
	}

	jobs := []jobmgmtsvc.BuiltinJobDef{
		{
			GroupCode:      "default",
			Name:           "Job Log Cleanup",
			Description:    "Cleans up scheduled-job execution logs according to global and job-level retention policies.",
			TaskType:       jobmeta.TaskTypeHandler,
			HandlerRef:     "host:cleanup-job-logs",
			Params:         map[string]any{},
			Timeout:        defaultManagedJobTimeout,
			Pattern:        "# 17 3 * * *",
			Timezone:       defaultManagedJobTimezone,
			Scope:          jobmeta.JobScopeMasterOnly,
			Concurrency:    jobmeta.JobConcurrencySingleton,
			MaxConcurrency: 1,
			MaxExecutions:  0,
			Status:         jobmeta.JobStatusEnabled,
		},
		{
			GroupCode:      "default",
			Name:           "Online Session Cleanup",
			Description:    "Cleans up inactive online sessions in the host according to the session-timeout policy.",
			TaskType:       jobmeta.TaskTypeHandler,
			HandlerRef:     "host:session-cleanup",
			Params:         map[string]any{},
			Timeout:        defaultManagedJobTimeout,
			Pattern:        formatEveryPattern(sessionCleanupInterval),
			Timezone:       defaultManagedJobTimezone,
			Scope:          jobmeta.JobScopeMasterOnly,
			Concurrency:    jobmeta.JobConcurrencySingleton,
			MaxConcurrency: 1,
			MaxExecutions:  0,
			Status:         jobmeta.JobStatusEnabled,
		},
	}

	if s.clusterSvc != nil && s.clusterSvc.IsEnabled() {
		jobs = append(jobs,
			jobmgmtsvc.BuiltinJobDef{
				GroupCode:      "default",
				Name:           "Access Topology Sync",
				Description:    "Synchronizes permission-topology revision snapshots across the cluster so authorization caches stay consistent on every node.",
				TaskType:       jobmeta.TaskTypeHandler,
				HandlerRef:     "host:access-topology-sync",
				Params:         map[string]any{},
				Timeout:        defaultManagedJobTimeout,
				Pattern:        formatEveryPattern(10 * time.Second),
				Timezone:       defaultManagedJobTimezone,
				Scope:          jobmeta.JobScopeAllNode,
				Concurrency:    jobmeta.JobConcurrencySingleton,
				MaxConcurrency: 1,
				MaxExecutions:  0,
				Status:         jobmeta.JobStatusEnabled,
			},
			jobmgmtsvc.BuiltinJobDef{
				GroupCode:      "default",
				Name:           "Runtime Parameter Sync",
				Description:    "Synchronizes protected runtime parameter snapshots across the cluster so each node keeps a fresh local cache.",
				TaskType:       jobmeta.TaskTypeHandler,
				HandlerRef:     "host:runtime-param-sync",
				Params:         map[string]any{},
				Timeout:        defaultManagedJobTimeout,
				Pattern:        formatEveryPattern(10 * time.Second),
				Timezone:       defaultManagedJobTimezone,
				Scope:          jobmeta.JobScopeAllNode,
				Concurrency:    jobmeta.JobConcurrencySingleton,
				MaxConcurrency: 1,
				MaxExecutions:  0,
				Status:         jobmeta.JobStatusEnabled,
			},
		)
	}

	return jobs
}

// buildPluginBuiltinJobs converts plugin-owned cron definitions into sys_job projections.
func (s *serviceImpl) buildPluginBuiltinJobs(ctx context.Context) ([]jobmgmtsvc.BuiltinJobDef, error) {
	if s == nil || s.pluginSvc == nil {
		return nil, nil
	}

	items, err := s.pluginSvc.ListManagedCronJobs(ctx)
	if err != nil {
		return nil, err
	}
	jobs := make([]jobmgmtsvc.BuiltinJobDef, 0, len(items))
	for _, item := range items {
		handlerRef, refErr := pluginbridge.BuildPluginCronHandlerRef(item.PluginID, item.Name)
		if refErr != nil {
			return nil, refErr
		}

		scope := item.Scope
		if !scope.IsValid() {
			scope = jobmeta.JobScopeAllNode
		}
		concurrency := item.Concurrency
		if !concurrency.IsValid() {
			concurrency = jobmeta.JobConcurrencySingleton
		}
		timeout := item.Timeout
		if timeout <= 0 {
			timeout = defaultManagedJobTimeout
		}
		timezone := strings.TrimSpace(item.Timezone)
		if timezone == "" {
			timezone = defaultManagedJobTimezone
		}
		name := strings.TrimSpace(item.DisplayName)
		if name == "" {
			name = strings.TrimSpace(item.Name)
		}
		description := strings.TrimSpace(item.Description)
		if description == "" {
			description = fmt.Sprintf("Built-in scheduled job registered by plugin %s.", strings.TrimSpace(item.PluginID))
		}

		jobs = append(jobs, jobmgmtsvc.BuiltinJobDef{
			GroupCode:      "default",
			Name:           name,
			Description:    description,
			TaskType:       jobmeta.TaskTypeHandler,
			HandlerRef:     handlerRef,
			Params:         map[string]any{},
			Timeout:        timeout,
			Pattern:        strings.TrimSpace(item.Pattern),
			Timezone:       timezone,
			Scope:          scope,
			Concurrency:    concurrency,
			MaxConcurrency: maxInt(item.MaxConcurrency, 1),
			MaxExecutions:  0,
			Status:         jobmeta.JobStatusEnabled,
		})
	}
	return jobs, nil
}

// invokeSessionCleanup runs the session cleanup built-in handler.
func (s *serviceImpl) invokeSessionCleanup(ctx context.Context, _ json.RawMessage) (any, error) {
	if s == nil || s.sessionStore == nil || s.sessionCfg == nil {
		return nil, bizerr.NewCode(CodeCronSessionCleanupDependencyMissing)
	}
	cleaned, err := s.sessionStore.CleanupInactive(ctx, s.sessionCfg.Timeout)
	if err != nil {
		return nil, err
	}
	return map[string]any{"cleanedCount": cleaned}, nil
}

// invokeAccessTopologySync runs the access-topology watcher handler.
func (s *serviceImpl) invokeAccessTopologySync(ctx context.Context, _ json.RawMessage) (any, error) {
	if s == nil || s.roleSvc == nil {
		return nil, bizerr.NewCode(CodeCronAccessTopologyDependencyMissing)
	}
	if err := s.roleSvc.SyncAccessTopologyRevision(ctx); err != nil {
		return nil, err
	}
	return map[string]any{"synced": true}, nil
}

// invokeRuntimeParamSync runs the runtime-parameter watcher handler.
func (s *serviceImpl) invokeRuntimeParamSync(ctx context.Context, _ json.RawMessage) (any, error) {
	if s == nil || s.configSvc == nil {
		return nil, bizerr.NewCode(CodeCronRuntimeParamDependencyMissing)
	}
	if err := s.configSvc.SyncRuntimeParamSnapshot(ctx); err != nil {
		return nil, err
	}
	return map[string]any{"synced": true}, nil
}

// isDuplicateHandlerError reports whether handler registration failed because
// the same built-in ref was already registered by an earlier startup path.
func isDuplicateHandlerError(err error) bool {
	return bizerr.Is(err, jobhandlersvc.CodeJobHandlerExists)
}

// formatEveryPattern converts one duration to the stable `@every` form stored
// for code-owned interval-based jobs.
func formatEveryPattern(duration time.Duration) string {
	if duration <= 0 {
		duration = time.Minute
	}
	return "@every " + duration.String()
}

// maxInt returns the larger of the provided integers.
func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}
