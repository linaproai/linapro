// This file synchronizes code-owned scheduled-job definitions into sys_job so
// host and plugin built-ins share the same management view as user-created jobs.

package jobmgmt

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/jobmeta"
)

const (
	defaultBuiltinGroupCode = "default"
)

// SyncBuiltinJobs upserts code-owned scheduled-job definitions into sys_job.
func (s *serviceImpl) SyncBuiltinJobs(ctx context.Context, jobs []BuiltinJobDef) error {
	if len(jobs) == 0 {
		return nil
	}

	for _, job := range jobs {
		if err := s.syncBuiltinJob(ctx, job); err != nil {
			return err
		}
	}
	return nil
}

// syncBuiltinJob upserts one code-owned scheduled-job definition.
func (s *serviceImpl) syncBuiltinJob(ctx context.Context, job BuiltinJobDef) error {
	record, existing, err := s.buildBuiltinJobRecord(ctx, job)
	if err != nil {
		return err
	}

	if existing == nil {
		_, err = dao.SysJob.Ctx(ctx).Data(record).Insert()
		return err
	}

	_, err = dao.SysJob.Ctx(ctx).
		Where(do.SysJob{Id: existing.Id}).
		Data(record).
		Update()
	return err
}

// buildBuiltinJobRecord validates one code-owned job definition and converts it
// to a persistent DO snapshot together with any existing row.
func (s *serviceImpl) buildBuiltinJobRecord(
	ctx context.Context,
	job BuiltinJobDef,
) (do.SysJob, *entity.SysJob, error) {
	groupCode := strings.TrimSpace(job.GroupCode)
	if groupCode == "" {
		groupCode = defaultBuiltinGroupCode
	}
	group, err := s.groupByCode(ctx, groupCode)
	if err != nil {
		return do.SysJob{}, nil, err
	}
	if group == nil {
		return do.SysJob{}, nil, gerror.Newf("定时任务分组不存在: %s", groupCode)
	}

	taskType := job.TaskType
	if !taskType.IsValid() {
		taskType = jobmeta.TaskTypeHandler
	}
	if taskType != jobmeta.TaskTypeHandler && taskType != jobmeta.TaskTypeShell {
		return do.SysJob{}, nil, gerror.New("源码注册任务类型不受支持")
	}

	name := strings.TrimSpace(job.Name)
	if name == "" {
		return do.SysJob{}, nil, gerror.New("源码注册任务名称不能为空")
	}

	timeout := job.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	if timeout%time.Second != 0 {
		return do.SysJob{}, nil, gerror.New("源码注册任务超时时间必须按秒配置")
	}

	pattern := strings.TrimSpace(job.Pattern)
	if pattern == "" {
		return do.SysJob{}, nil, gerror.New("源码注册任务表达式不能为空")
	}
	if len(pattern) > 128 {
		return do.SysJob{}, nil, gerror.New("源码注册任务表达式长度不能超过128个字符")
	}

	timezone := strings.TrimSpace(job.Timezone)
	if timezone == "" {
		timezone = "Asia/Shanghai"
	}
	if _, _, err = normalizeJobTimezone(timezone); err != nil {
		return do.SysJob{}, nil, err
	}

	scope := job.Scope
	if !scope.IsValid() {
		scope = jobmeta.JobScopeAllNode
	}
	concurrency := job.Concurrency
	if !concurrency.IsValid() {
		concurrency = jobmeta.JobConcurrencySingleton
	}
	maxConcurrency := job.MaxConcurrency
	if concurrency == jobmeta.JobConcurrencySingleton {
		maxConcurrency = 1
	}
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}
	if job.MaxExecutions < 0 {
		return do.SysJob{}, nil, gerror.New("源码注册任务最大执行次数不能小于0")
	}

	status := job.Status
	if status != jobmeta.JobStatusEnabled && status != jobmeta.JobStatusDisabled {
		status = jobmeta.JobStatusEnabled
	}

	paramsJSON := ""
	handlerRef := strings.TrimSpace(job.HandlerRef)
	shellCmd := ""
	workDir := ""
	envJSON := ""

	switch taskType {
	case jobmeta.TaskTypeHandler:
		if handlerRef == "" {
			return do.SysJob{}, nil, gerror.New("源码注册任务处理器引用不能为空")
		}
		paramsData, marshalErr := json.Marshal(job.Params)
		if marshalErr != nil {
			return do.SysJob{}, nil, gerror.Wrap(marshalErr, "序列化源码注册任务参数失败")
		}
		paramsJSON = string(paramsData)
	case jobmeta.TaskTypeShell:
		shellCmd = strings.TrimSpace(shellCmd)
	}

	existing, err := s.builtinJobByHandlerRef(ctx, handlerRef)
	if err != nil {
		return do.SysJob{}, nil, err
	}

	stopReason := ""
	effectiveStatus := status
	if taskType == jobmeta.TaskTypeHandler && strings.HasPrefix(handlerRef, "plugin:") {
		if _, ok := s.registry.Lookup(handlerRef); !ok {
			effectiveStatus = jobmeta.JobStatusPausedByPlugin
			stopReason = string(jobmeta.StopReasonPluginUnavailable)
		}
	}
	if effectiveStatus == jobmeta.JobStatusDisabled {
		stopReason = string(jobmeta.StopReasonManual)
	}

	executedCount := int64(0)
	createdBy := int64(0)
	if existing != nil {
		executedCount = existing.ExecutedCount
		createdBy = existing.CreatedBy
	}

	record := do.SysJob{
		GroupId:              group.Id,
		Name:                 name,
		Description:          strings.TrimSpace(job.Description),
		TaskType:             string(taskType),
		HandlerRef:           handlerRef,
		Params:               paramsJSON,
		TimeoutSeconds:       int(timeout.Seconds()),
		ShellCmd:             shellCmd,
		WorkDir:              workDir,
		Env:                  envJSON,
		CronExpr:             pattern,
		Timezone:             timezone,
		Scope:                string(scope),
		Concurrency:          string(concurrency),
		MaxConcurrency:       maxConcurrency,
		MaxExecutions:        job.MaxExecutions,
		ExecutedCount:        executedCount,
		StopReason:           stopReason,
		LogRetentionOverride: strings.TrimSpace(job.LogRetentionRaw),
		Status:               string(effectiveStatus),
		IsBuiltin:            1,
		SeedVersion:          1,
		CreatedBy:            createdBy,
		UpdatedBy:            0,
	}
	return record, existing, nil
}

// groupByCode queries one job group by stable code.
func (s *serviceImpl) groupByCode(ctx context.Context, code string) (*entity.SysJobGroup, error) {
	var group *entity.SysJobGroup
	err := dao.SysJobGroup.Ctx(ctx).
		Where(do.SysJobGroup{Code: strings.TrimSpace(code)}).
		Scan(&group)
	return group, err
}

// builtinJobByHandlerRef queries one code-owned job by handler reference.
func (s *serviceImpl) builtinJobByHandlerRef(ctx context.Context, handlerRef string) (*entity.SysJob, error) {
	trimmedRef := strings.TrimSpace(handlerRef)
	if trimmedRef == "" {
		return nil, nil
	}

	var job *entity.SysJob
	err := dao.SysJob.Ctx(ctx).
		Where(do.SysJob{IsBuiltin: 1, HandlerRef: trimmedRef}).
		Scan(&job)
	return job, err
}
