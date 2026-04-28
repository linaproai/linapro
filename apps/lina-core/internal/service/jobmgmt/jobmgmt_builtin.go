// This file synchronizes code-owned scheduled-job definitions into sys_job so
// host and plugin built-ins share the same management view as user-created jobs.

package jobmgmt

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/jobmeta"
	"lina-core/pkg/bizerr"
)

const (
	defaultBuiltinGroupCode = "default"
)

// builtinJobPlan stores one reconciled code-owned job record together with the
// desired identity set used to prune removed built-ins.
type builtinJobPlan struct {
	items       []builtinJobPlanItem
	handlerRefs map[string]struct{}
	nameKeys    map[string]struct{}
}

// builtinJobPlanItem stores one upsert-ready built-in job snapshot.
type builtinJobPlanItem struct {
	record   do.SysJob
	existing *entity.SysJob
}

// SyncBuiltinJobs upserts code-owned scheduled-job definitions into sys_job
// without pruning removed built-ins.
func (s *serviceImpl) SyncBuiltinJobs(ctx context.Context, jobs []BuiltinJobDef) error {
	if len(jobs) == 0 {
		return nil
	}

	for _, job := range jobs {
		record, existing, err := s.buildBuiltinJobRecord(ctx, job)
		if err != nil {
			return err
		}
		if err = s.upsertBuiltinJobRecord(ctx, record, existing); err != nil {
			return err
		}
	}
	return nil
}

// ReconcileBuiltinJobs refreshes the full code-owned job projection and
// removes stale built-ins before writing the current snapshot.
func (s *serviceImpl) ReconcileBuiltinJobs(ctx context.Context, jobs []BuiltinJobDef) error {
	if len(jobs) == 0 {
		return nil
	}

	plan, err := s.buildBuiltinJobPlan(ctx, jobs)
	if err != nil {
		return err
	}
	if err = s.pruneRemovedBuiltinJobs(ctx, plan); err != nil {
		return err
	}
	for _, item := range plan.items {
		if err = s.upsertBuiltinJobRecord(ctx, item.record, item.existing); err != nil {
			return err
		}
	}
	return nil
}

// buildBuiltinJobPlan materializes one full synchronization plan so removed
// built-ins can be pruned before the remaining records are upserted.
func (s *serviceImpl) buildBuiltinJobPlan(
	ctx context.Context,
	jobs []BuiltinJobDef,
) (*builtinJobPlan, error) {
	plan := &builtinJobPlan{
		items:       make([]builtinJobPlanItem, 0, len(jobs)),
		handlerRefs: make(map[string]struct{}),
		nameKeys:    make(map[string]struct{}),
	}

	for _, job := range jobs {
		record, existing, err := s.buildBuiltinJobRecord(ctx, job)
		if err != nil {
			return nil, err
		}
		plan.items = append(plan.items, builtinJobPlanItem{
			record:   record,
			existing: existing,
		})

		nameKey := buildBuiltinJobNameKey(gconv.Uint64(record.GroupId), record.Name)
		plan.nameKeys[nameKey] = struct{}{}

		handlerRef := strings.TrimSpace(gconv.String(record.HandlerRef))
		if handlerRef != "" {
			plan.handlerRefs[handlerRef] = struct{}{}
		}
	}
	return plan, nil
}

// pruneRemovedBuiltinJobs hard-deletes code-owned jobs that no longer exist in
// the current host or plugin definitions so stale handler refs never leak into
// scheduler startup or name uniqueness checks.
func (s *serviceImpl) pruneRemovedBuiltinJobs(
	ctx context.Context,
	plan *builtinJobPlan,
) error {
	if plan == nil {
		return nil
	}

	var jobs []*entity.SysJob
	if err := dao.SysJob.Ctx(ctx).
		Where(do.SysJob{IsBuiltin: 1}).
		Scan(&jobs); err != nil {
		return err
	}

	for _, job := range jobs {
		if job == nil || job.Id == 0 {
			continue
		}
		if _, ok := plan.handlerRefs[strings.TrimSpace(job.HandlerRef)]; ok {
			continue
		}
		if _, ok := plan.nameKeys[buildBuiltinJobNameKey(job.GroupId, job.Name)]; ok {
			continue
		}
		if err := s.deleteBuiltinJobHard(ctx, job.Id); err != nil {
			return err
		}
	}
	return nil
}

// deleteBuiltinJobHard removes one stale code-owned job and its execution logs.
func (s *serviceImpl) deleteBuiltinJobHard(ctx context.Context, jobID uint64) error {
	if jobID == 0 {
		return nil
	}

	if s != nil && s.scheduler != nil {
		s.scheduler.Remove(jobID)
	}

	return dao.SysJob.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if _, err := dao.SysJobLog.Ctx(ctx).
			Where(do.SysJobLog{JobId: jobID}).
			Delete(); err != nil {
			return err
		}
		_, err := dao.SysJob.Ctx(ctx).
			Unscoped().
			Where(do.SysJob{Id: jobID}).
			Delete()
		return err
	})
}

// buildBuiltinJobNameKey returns the stable uniqueness key for one built-in
// job inside the persistent sys_job table.
func buildBuiltinJobNameKey(groupID uint64, name any) string {
	return fmt.Sprintf("%d:%s", groupID, strings.TrimSpace(fmt.Sprint(name)))
}

// upsertBuiltinJobRecord writes one prepared code-owned scheduled-job snapshot.
func (s *serviceImpl) upsertBuiltinJobRecord(
	ctx context.Context,
	record do.SysJob,
	existing *entity.SysJob,
) error {
	if existing == nil {
		_, err := dao.SysJob.Ctx(ctx).Data(record).Insert()
		return err
	}

	_, err := dao.SysJob.Ctx(ctx).
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
		return do.SysJob{}, nil, bizerr.NewCode(
			CodeJobBuiltinGroupNotFound,
			bizerr.P("groupCode", groupCode),
		)
	}

	taskType := job.TaskType
	if !taskType.IsValid() {
		taskType = jobmeta.TaskTypeHandler
	}
	if taskType != jobmeta.TaskTypeHandler && taskType != jobmeta.TaskTypeShell {
		return do.SysJob{}, nil, bizerr.NewCode(CodeJobBuiltinTypeUnsupported)
	}

	name := strings.TrimSpace(job.Name)
	if name == "" {
		return do.SysJob{}, nil, bizerr.NewCode(CodeJobBuiltinNameRequired)
	}

	timeout := job.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	if timeout%time.Second != 0 {
		return do.SysJob{}, nil, bizerr.NewCode(CodeJobBuiltinTimeoutSecondAlignedRequired)
	}

	pattern := strings.TrimSpace(job.Pattern)
	if pattern == "" {
		return do.SysJob{}, nil, bizerr.NewCode(CodeJobBuiltinCronExpressionRequired)
	}
	if len(pattern) > 128 {
		return do.SysJob{}, nil, bizerr.NewCode(CodeJobBuiltinCronExpressionTooLong)
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
		return do.SysJob{}, nil, bizerr.NewCode(CodeJobBuiltinMaxExecutionsInvalid)
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
			return do.SysJob{}, nil, bizerr.NewCode(CodeJobBuiltinHandlerRefRequired)
		}
		paramsData, marshalErr := json.Marshal(job.Params)
		if marshalErr != nil {
			return do.SysJob{}, nil, bizerr.WrapCode(marshalErr, CodeJobBuiltinParamsMarshalFailed)
		}
		paramsJSON = string(paramsData)
	case jobmeta.TaskTypeShell:
		shellCmd = strings.TrimSpace(shellCmd)
	}

	existing, err := s.builtinJobByHandlerRef(ctx, handlerRef)
	if err != nil {
		return do.SysJob{}, nil, err
	}
	if existing == nil {
		existing, err = s.builtinJobByGroupAndName(ctx, group.Id, name)
		if err != nil {
			return do.SysJob{}, nil, err
		}
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

// builtinJobByGroupAndName queries one code-owned job by group and display name.
func (s *serviceImpl) builtinJobByGroupAndName(
	ctx context.Context,
	groupID uint64,
	name string,
) (*entity.SysJob, error) {
	trimmedName := strings.TrimSpace(name)
	if groupID == 0 || trimmedName == "" {
		return nil, nil
	}

	var job *entity.SysJob
	err := dao.SysJob.Ctx(ctx).
		Where(do.SysJob{
			IsBuiltin: 1,
			GroupId:   groupID,
			Name:      trimmedName,
		}).
		Scan(&job)
	return job, err
}
