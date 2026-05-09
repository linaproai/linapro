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
	"lina-core/internal/service/startupstats"
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
// without pruning removed built-ins and returns declaration-derived projection
// snapshots keyed with stable sys_job IDs.
func (s *serviceImpl) SyncBuiltinJobs(ctx context.Context, jobs []BuiltinJobDef) ([]*entity.SysJob, error) {
	if len(jobs) == 0 {
		return nil, nil
	}

	syncCtx, err := s.withStartupDataSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	projections := make([]*entity.SysJob, 0, len(jobs))
	for _, job := range jobs {
		record, existing, err := s.buildBuiltinJobRecord(syncCtx, job)
		if err != nil {
			return nil, err
		}
		projection, err := s.upsertBuiltinJobRecord(syncCtx, record, existing)
		if err != nil {
			return nil, err
		}
		projections = append(projections, projection)
	}
	return projections, nil
}

// ReconcileBuiltinJobs refreshes the full code-owned job projection and
// removes stale built-ins before writing the current snapshot.
func (s *serviceImpl) ReconcileBuiltinJobs(ctx context.Context, jobs []BuiltinJobDef) ([]*entity.SysJob, error) {
	if len(jobs) == 0 {
		return nil, nil
	}

	syncCtx, err := s.withStartupDataSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	plan, err := s.buildBuiltinJobPlan(syncCtx, jobs)
	if err != nil {
		return nil, err
	}
	if err = s.pruneRemovedBuiltinJobs(syncCtx, plan); err != nil {
		return nil, err
	}
	projections := make([]*entity.SysJob, 0, len(plan.items))
	for _, item := range plan.items {
		projection, err := s.upsertBuiltinJobRecord(syncCtx, item.record, item.existing)
		if err != nil {
			return nil, err
		}
		projections = append(projections, projection)
	}
	return projections, nil
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

		nameKey := buildBuiltinJobNameKey(gconv.Int64(record.GroupId), record.Name)
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

	jobs := []*entity.SysJob(nil)
	if snapshot := startupDataSnapshotFromContext(ctx); snapshot != nil {
		jobs = snapshot.listBuiltinJobs()
	} else {
		if err := dao.SysJob.Ctx(ctx).
			Where(do.SysJob{IsBuiltin: 1}).
			Scan(&jobs); err != nil {
			return err
		}
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
		if snapshot := startupDataSnapshotFromContext(ctx); snapshot != nil {
			snapshot.deleteBuiltinJob(job.Id)
		}
	}
	return nil
}

// deleteBuiltinJobHard removes one stale code-owned job and its execution logs.
func (s *serviceImpl) deleteBuiltinJobHard(ctx context.Context, jobID int64) error {
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
func buildBuiltinJobNameKey(groupID int64, name any) string {
	return fmt.Sprintf("%d:%s", groupID, strings.TrimSpace(fmt.Sprint(name)))
}

// upsertBuiltinJobRecord writes one prepared code-owned scheduled-job snapshot.
func (s *serviceImpl) upsertBuiltinJobRecord(
	ctx context.Context,
	record do.SysJob,
	existing *entity.SysJob,
) (*entity.SysJob, error) {
	if existing == nil {
		insertID, err := dao.SysJob.Ctx(ctx).Data(record).InsertAndGetId()
		if err != nil {
			return nil, err
		}
		startupstats.Add(ctx, startupstats.CounterBuiltinJobProjections, 1)
		projection := buildBuiltinJobEntity(int64(insertID), record)
		if snapshot := startupDataSnapshotFromContext(ctx); snapshot != nil {
			snapshot.storeBuiltinJob(projection)
		}
		return projection, nil
	}

	if builtinJobRecordMatches(existing, record) {
		startupstats.Add(ctx, startupstats.CounterBuiltinJobProjectionNoop, 1)
		return buildBuiltinJobEntity(existing.Id, record), nil
	}

	_, err := dao.SysJob.Ctx(ctx).
		Where(do.SysJob{Id: existing.Id}).
		Data(record).
		Update()
	if err != nil {
		return nil, err
	}
	startupstats.Add(ctx, startupstats.CounterBuiltinJobProjections, 1)
	projection := buildBuiltinJobEntity(existing.Id, record)
	if snapshot := startupDataSnapshotFromContext(ctx); snapshot != nil {
		snapshot.storeBuiltinJob(projection)
	}
	return projection, nil
}

// builtinJobRecordMatches reports whether an existing built-in job row already
// matches the code-owned projection that would otherwise be written.
func builtinJobRecordMatches(existing *entity.SysJob, record do.SysJob) bool {
	if existing == nil {
		return false
	}
	return existing.GroupId == gconv.Int64(record.GroupId) &&
		existing.Name == gconv.String(record.Name) &&
		existing.Description == gconv.String(record.Description) &&
		existing.TaskType == gconv.String(record.TaskType) &&
		existing.HandlerRef == gconv.String(record.HandlerRef) &&
		existing.Params == gconv.String(record.Params) &&
		existing.TimeoutSeconds == gconv.Int(record.TimeoutSeconds) &&
		existing.ShellCmd == gconv.String(record.ShellCmd) &&
		existing.WorkDir == gconv.String(record.WorkDir) &&
		existing.Env == gconv.String(record.Env) &&
		existing.CronExpr == gconv.String(record.CronExpr) &&
		existing.Timezone == gconv.String(record.Timezone) &&
		existing.Scope == gconv.String(record.Scope) &&
		existing.Concurrency == gconv.String(record.Concurrency) &&
		existing.MaxConcurrency == gconv.Int(record.MaxConcurrency) &&
		existing.MaxExecutions == gconv.Int(record.MaxExecutions) &&
		existing.ExecutedCount == gconv.Int64(record.ExecutedCount) &&
		existing.StopReason == gconv.String(record.StopReason) &&
		existing.LogRetentionOverride == gconv.String(record.LogRetentionOverride) &&
		existing.Status == gconv.String(record.Status) &&
		existing.IsBuiltin == gconv.Int(record.IsBuiltin) &&
		existing.SeedVersion == gconv.Int(record.SeedVersion) &&
		existing.CreatedBy == gconv.Int64(record.CreatedBy) &&
		existing.UpdatedBy == gconv.Int64(record.UpdatedBy)
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
	if snapshot := startupDataSnapshotFromContext(ctx); snapshot != nil {
		return snapshot.groupByCode(code), nil
	}

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
	if snapshot := startupDataSnapshotFromContext(ctx); snapshot != nil {
		return snapshot.builtinJobByHandlerRef(trimmedRef), nil
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
	groupID int64,
	name string,
) (*entity.SysJob, error) {
	trimmedName := strings.TrimSpace(name)
	if groupID == 0 || trimmedName == "" {
		return nil, nil
	}
	if snapshot := startupDataSnapshotFromContext(ctx); snapshot != nil {
		return snapshot.builtinJobByGroupAndName(groupID, trimmedName), nil
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
