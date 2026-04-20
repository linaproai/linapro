// This file projects host and plugin code-owned scheduled jobs into sys_job
// and publishes the host-side handler callbacks they execute through.

package cron

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	jobhandlersvc "lina-core/internal/service/jobhandler"
	"lina-core/internal/service/jobmeta"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
	"lina-core/pkg/pluginbridge"
)

const (
	defaultManagedJobTimezone = "Asia/Shanghai"
	defaultManagedJobTimeout  = 5 * time.Minute
)

// warmServerMonitor performs the immediate startup collection before the
// periodic collector is handed over to the persistent scheduler.
func (s *serviceImpl) warmServerMonitor(ctx context.Context) {
	if s == nil || s.serverMonSvc == nil {
		return
	}
	s.serverMonSvc.CollectAndStore(ctx)
}

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
	return s.builtinSyncer.SyncBuiltinJobs(ctx, jobs)
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
			DisplayName:  "在线会话清理",
			Description:  "按会话超时策略清理宿主中的失活在线会话。",
			ParamsSchema: `{"type":"object","properties":{}}`,
			Source:       jobmeta.HandlerSourceHost,
			Invoke:       s.invokeSessionCleanup,
		},
		{
			Ref:          "host:server-monitor-collector",
			DisplayName:  "服务监控采集",
			Description:  "采集当前节点的服务监控指标并写入监控快照。",
			ParamsSchema: `{"type":"object","properties":{}}`,
			Source:       jobmeta.HandlerSourceHost,
			Invoke:       s.invokeServerMonitorCollector,
		},
		{
			Ref:          "host:server-monitor-cleanup",
			DisplayName:  "服务监控清理",
			Description:  "按监控保留窗口清理过期的服务监控数据。",
			ParamsSchema: `{"type":"object","properties":{}}`,
			Source:       jobmeta.HandlerSourceHost,
			Invoke:       s.invokeServerMonitorCleanup,
		},
	}

	if s.clusterSvc != nil && s.clusterSvc.IsEnabled() {
		handlers = append(handlers,
			jobhandlersvc.HandlerDef{
				Ref:          "host:access-topology-sync",
				DisplayName:  "权限拓扑同步",
				Description:  "同步集群内权限拓扑版本快照，保持各节点鉴权缓存一致。",
				ParamsSchema: `{"type":"object","properties":{}}`,
				Source:       jobmeta.HandlerSourceHost,
				Invoke:       s.invokeAccessTopologySync,
			},
			jobhandlersvc.HandlerDef{
				Ref:          "host:runtime-param-sync",
				DisplayName:  "运行时参数同步",
				Description:  "同步集群内受保护运行时参数快照，保持各节点本地缓存一致。",
				ParamsSchema: `{"type":"object","properties":{}}`,
				Source:       jobmeta.HandlerSourceHost,
				Invoke:       s.invokeRuntimeParamSync,
			},
		)
	}

	for _, definition := range handlers {
		if err := s.registry.Register(definition); err != nil &&
			!strings.Contains(err.Error(), "已存在") {
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
	monitorInterval := 30 * time.Second
	if s.monCfg != nil {
		if s.monCfg.Interval > 0 {
			monitorInterval = s.monCfg.Interval
		}
	}

	jobs := []jobmgmtsvc.BuiltinJobDef{
		{
			GroupCode:      "default",
			Name:           "任务日志清理",
			Description:    "按全局与任务级日志保留策略清理定时任务执行日志。",
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
			Name:           "在线会话清理",
			Description:    "按会话超时策略清理宿主中的失活在线会话。",
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
		{
			GroupCode:      "default",
			Name:           "服务监控采集",
			Description:    "采集当前节点的服务监控指标并写入监控快照。",
			TaskType:       jobmeta.TaskTypeHandler,
			HandlerRef:     "host:server-monitor-collector",
			Params:         map[string]any{},
			Timeout:        defaultManagedJobTimeout,
			Pattern:        formatEveryPattern(monitorInterval),
			Timezone:       defaultManagedJobTimezone,
			Scope:          jobmeta.JobScopeAllNode,
			Concurrency:    jobmeta.JobConcurrencySingleton,
			MaxConcurrency: 1,
			MaxExecutions:  0,
			Status:         jobmeta.JobStatusEnabled,
		},
		{
			GroupCode:      "default",
			Name:           "服务监控清理",
			Description:    "按监控保留窗口清理过期的服务监控数据。",
			TaskType:       jobmeta.TaskTypeHandler,
			HandlerRef:     "host:server-monitor-cleanup",
			Params:         map[string]any{},
			Timeout:        defaultManagedJobTimeout,
			Pattern:        "# * * * * *",
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
				Name:           "权限拓扑同步",
				Description:    "同步集群内权限拓扑版本快照，保持各节点鉴权缓存一致。",
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
				Name:           "运行时参数同步",
				Description:    "同步集群内受保护运行时参数快照，保持各节点本地缓存一致。",
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
			description = fmt.Sprintf("插件 %s 注册的内置定时任务。", strings.TrimSpace(item.PluginID))
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
		return nil, gerror.New("在线会话清理依赖未初始化")
	}
	cleaned, err := s.sessionStore.CleanupInactive(ctx, s.sessionCfg.Timeout)
	if err != nil {
		return nil, err
	}
	return map[string]any{"cleanedCount": cleaned}, nil
}

// invokeServerMonitorCollector runs the monitor collector built-in handler.
func (s *serviceImpl) invokeServerMonitorCollector(ctx context.Context, _ json.RawMessage) (any, error) {
	if s == nil || s.serverMonSvc == nil {
		return nil, gerror.New("服务监控采集依赖未初始化")
	}
	s.serverMonSvc.CollectAndStore(ctx)
	return map[string]any{"collected": true}, nil
}

// invokeServerMonitorCleanup runs the monitor cleanup built-in handler.
func (s *serviceImpl) invokeServerMonitorCleanup(ctx context.Context, _ json.RawMessage) (any, error) {
	if s == nil || s.serverMonSvc == nil || s.monCfg == nil {
		return nil, gerror.New("服务监控清理依赖未初始化")
	}
	staleThreshold := s.monCfg.Interval * time.Duration(s.monCfg.RetentionMultiplier)
	cleaned, err := s.serverMonSvc.CleanupStale(ctx, staleThreshold)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"cleanedCount":   cleaned,
		"staleThreshold": staleThreshold.String(),
	}, nil
}

// invokeAccessTopologySync runs the access-topology watcher handler.
func (s *serviceImpl) invokeAccessTopologySync(ctx context.Context, _ json.RawMessage) (any, error) {
	if s == nil || s.roleSvc == nil {
		return nil, gerror.New("权限拓扑同步依赖未初始化")
	}
	if err := s.roleSvc.SyncAccessTopologyRevision(ctx); err != nil {
		return nil, err
	}
	return map[string]any{"synced": true}, nil
}

// invokeRuntimeParamSync runs the runtime-parameter watcher handler.
func (s *serviceImpl) invokeRuntimeParamSync(ctx context.Context, _ json.RawMessage) (any, error) {
	if s == nil || s.configSvc == nil {
		return nil, gerror.New("运行时参数同步依赖未初始化")
	}
	if err := s.configSvc.SyncRuntimeParamSnapshot(ctx); err != nil {
		return nil, err
	}
	return map[string]any{"synced": true}, nil
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
