// This file implements scheduled-job CRUD and validation logic.

package jobmgmt

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/jobhandler"
	"lina-core/internal/service/jobmeta"
)

// ListJobs returns scheduled jobs with pagination and group metadata.
func (s *serviceImpl) ListJobs(ctx context.Context, in ListJobsInput) (*ListJobsOutput, error) {
	model := dao.SysJob.Ctx(ctx)
	cols := dao.SysJob.Columns()

	if in.GroupID != nil && *in.GroupID > 0 {
		model = model.Where(cols.GroupId, *in.GroupID)
	}
	if in.Status.IsValid() {
		model = model.Where(cols.Status, string(in.Status))
	}
	if in.TaskType.IsValid() {
		model = model.Where(cols.TaskType, string(in.TaskType))
	}
	if in.Scope.IsValid() {
		model = model.Where(cols.Scope, string(in.Scope))
	}
	if in.Concurrency.IsValid() {
		model = model.Where(cols.Concurrency, string(in.Concurrency))
	}
	if keyword := strings.TrimSpace(in.Keyword); keyword != "" {
		model = model.WhereLike(cols.Name, "%"+keyword+"%").WhereOrLike(cols.Description, "%"+keyword+"%")
	}

	total, err := model.Count()
	if err != nil {
		return nil, err
	}

	var jobs []*entity.SysJob
	err = applySingleOrder(
		model,
		in.OrderBy,
		in.OrderDirection,
		map[string]string{
			"id":         cols.Id,
			"name":       cols.Name,
			"group_id":   cols.GroupId,
			"status":     cols.Status,
			"task_type":  cols.TaskType,
			"created_at": cols.CreatedAt,
			"updated_at": cols.UpdatedAt,
		},
		cols.UpdatedAt,
		"desc",
	).Page(in.PageNum, in.PageSize).Scan(&jobs)
	if err != nil {
		return nil, err
	}

	groupMap, err := s.groupMapByJobGroupIDs(ctx, jobs)
	if err != nil {
		return nil, err
	}
	items := make([]*JobListItem, 0, len(jobs))
	for _, job := range jobs {
		if job == nil {
			continue
		}
		group := groupMap[job.GroupId]
		item := &JobListItem{SysJob: job}
		if group != nil {
			item.GroupCode = group.Code
			item.GroupName = group.Name
		}
		items = append(items, item)
	}
	return &ListJobsOutput{List: items, Total: total}, nil
}

// GetJob returns one scheduled-job detail snapshot.
func (s *serviceImpl) GetJob(ctx context.Context, id uint64) (*JobDetailOutput, error) {
	job, err := s.jobByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, gerror.New("定时任务不存在")
	}

	group, err := s.groupByID(ctx, job.GroupId)
	if err != nil {
		return nil, err
	}

	out := &JobDetailOutput{SysJob: job}
	if group != nil {
		out.GroupCode = group.Code
		out.GroupName = group.Name
	}
	return out, nil
}

// CreateJob persists one new scheduled job and refreshes the scheduler when needed.
func (s *serviceImpl) CreateJob(ctx context.Context, in SaveJobInput) (uint64, error) {
	jobRecord, err := s.normalizeJobRecord(ctx, nil, in)
	if err != nil {
		return 0, err
	}

	insertID, err := dao.SysJob.Ctx(ctx).Data(jobRecord).InsertAndGetId()
	if err != nil {
		return 0, err
	}
	jobID := gconv.Uint64(insertID)
	if jobmeta.NormalizeJobStatus(gconv.String(jobRecord.Status)) == jobmeta.JobStatusEnabled && s.scheduler != nil {
		if err = s.scheduler.Refresh(ctx, jobID); err != nil {
			return 0, err
		}
	}
	return jobID, nil
}

// UpdateJob updates one scheduled job and refreshes the scheduler when needed.
func (s *serviceImpl) UpdateJob(ctx context.Context, in UpdateJobInput) error {
	existing, err := s.jobByID(ctx, in.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return gerror.New("定时任务不存在")
	}

	if existing.IsBuiltin == 1 {
		if err = validateBuiltinMutableFields(existing, in.SaveJobInput); err != nil {
			return err
		}
	}

	jobRecord, err := s.normalizeJobRecord(ctx, existing, in.SaveJobInput)
	if err != nil {
		return err
	}

	_, err = dao.SysJob.Ctx(ctx).
		Where(do.SysJob{Id: in.ID}).
		Data(jobRecord).
		Update()
	if err != nil {
		return err
	}

	if jobmeta.NormalizeJobStatus(gconv.String(jobRecord.Status)) == jobmeta.JobStatusEnabled {
		if s.scheduler != nil {
			return s.scheduler.Refresh(ctx, in.ID)
		}
		return nil
	}
	if s.scheduler != nil {
		s.scheduler.Remove(in.ID)
	}
	return nil
}

// DeleteJobs removes one or more non-built-in scheduled jobs.
func (s *serviceImpl) DeleteJobs(ctx context.Context, ids string) error {
	jobIDs := parseUint64IDs(ids)
	if len(jobIDs) == 0 {
		return gerror.New("请选择要删除的定时任务")
	}

	for _, jobID := range jobIDs {
		job, err := s.jobByID(ctx, jobID)
		if err != nil {
			return err
		}
		if job == nil {
			continue
		}
		if job.IsBuiltin == 1 {
			return gerror.New("系统内置任务不允许删除")
		}
	}

	_, err := dao.SysJob.Ctx(ctx).
		WhereIn(dao.SysJob.Columns().Id, jobIDs).
		Delete()
	if err != nil {
		return err
	}
	if s.scheduler != nil {
		for _, jobID := range jobIDs {
			s.scheduler.Remove(jobID)
		}
	}
	return nil
}

// jobByID returns one scheduled job by ID.
func (s *serviceImpl) jobByID(ctx context.Context, id uint64) (*entity.SysJob, error) {
	var job *entity.SysJob
	err := dao.SysJob.Ctx(ctx).
		Where(do.SysJob{Id: id}).
		Scan(&job)
	return job, err
}

// groupMapByJobGroupIDs loads all groups referenced by the given jobs.
func (s *serviceImpl) groupMapByJobGroupIDs(
	ctx context.Context,
	jobs []*entity.SysJob,
) (map[uint64]*entity.SysJobGroup, error) {
	groupIDs := make([]uint64, 0, len(jobs))
	for _, job := range jobs {
		if job == nil || job.GroupId == 0 {
			continue
		}
		groupIDs = append(groupIDs, job.GroupId)
	}
	if len(groupIDs) == 0 {
		return map[uint64]*entity.SysJobGroup{}, nil
	}

	var groups []*entity.SysJobGroup
	err := dao.SysJobGroup.Ctx(ctx).
		WhereIn(dao.SysJobGroup.Columns().Id, groupIDs).
		Scan(&groups)
	if err != nil {
		return nil, err
	}

	groupMap := make(map[uint64]*entity.SysJobGroup, len(groups))
	for _, group := range groups {
		if group == nil {
			continue
		}
		groupMap[group.Id] = group
	}
	return groupMap, nil
}

// normalizeJobRecord validates one mutable job payload and converts it to DO fields.
func (s *serviceImpl) normalizeJobRecord(
	ctx context.Context,
	existing *entity.SysJob,
	in SaveJobInput,
) (do.SysJob, error) {
	group, err := s.groupByID(ctx, in.GroupID)
	if err != nil {
		return do.SysJob{}, err
	}
	if group == nil {
		return do.SysJob{}, gerror.New("任务分组不存在")
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		return do.SysJob{}, gerror.New("任务名称不能为空")
	}
	if in.Timeout <= 0 {
		return do.SysJob{}, gerror.New("任务超时时间必须大于0")
	}
	if in.Timeout > 24*time.Hour {
		return do.SysJob{}, gerror.New("任务超时时间不能超过24小时")
	}
	if strings.TrimSpace(in.CronExpr) == "" {
		return do.SysJob{}, gerror.New("Cron 表达式不能为空")
	}
	timezone := strings.TrimSpace(in.Timezone)
	if timezone == "" {
		timezone = "Asia/Shanghai"
	}
	if _, err = time.LoadLocation(timezone); err != nil {
		return do.SysJob{}, gerror.Wrap(err, "任务时区不合法")
	}
	if _, err = s.PreviewCron(ctx, in.CronExpr, timezone); err != nil {
		return do.SysJob{}, err
	}
	if !in.TaskType.IsValid() {
		return do.SysJob{}, gerror.New("任务类型不受支持")
	}
	if !in.Scope.IsValid() {
		return do.SysJob{}, gerror.New("任务调度范围不受支持")
	}
	if !in.Concurrency.IsValid() {
		return do.SysJob{}, gerror.New("任务并发策略不受支持")
	}
	if !in.Status.IsValid() || in.Status == jobmeta.JobStatusPausedByPlugin {
		return do.SysJob{}, gerror.New("任务状态不受支持")
	}
	if in.MaxExecutions < 0 {
		return do.SysJob{}, gerror.New("最大执行次数不能小于0")
	}

	maxConcurrency := in.MaxConcurrency
	if in.Concurrency == jobmeta.JobConcurrencySingleton {
		maxConcurrency = 1
	}
	if maxConcurrency <= 0 {
		return do.SysJob{}, gerror.New("最大并发数必须大于0")
	}

	paramsJSON := ""
	envJSON := ""
	handlerRef := strings.TrimSpace(in.HandlerRef)
	shellCmd := strings.TrimSpace(in.ShellCmd)
	workDir := strings.TrimSpace(in.WorkDir)

	switch in.TaskType {
	case jobmeta.TaskTypeHandler:
		def, ok := s.registry.Lookup(handlerRef)
		if !ok {
			return do.SysJob{}, gerror.New("任务处理器不存在")
		}
		paramsData, marshalErr := json.Marshal(in.Params)
		if marshalErr != nil {
			return do.SysJob{}, gerror.Wrap(marshalErr, "序列化任务参数失败")
		}
		if err = jobhandler.ValidateParams(def.ParamsSchema, paramsData); err != nil {
			return do.SysJob{}, err
		}
		paramsJSON = string(paramsData)
		shellCmd = ""
		workDir = ""

	case jobmeta.TaskTypeShell:
		if !s.configSvc.IsCronShellEnabled(ctx) {
			return do.SysJob{}, gerror.New("当前环境未启用 Shell 任务")
		}
		if shellCmd == "" {
			return do.SysJob{}, gerror.New("Shell 脚本内容不能为空")
		}
		if err = validateWorkDir(workDir); err != nil {
			return do.SysJob{}, err
		}
		envData, marshalErr := json.Marshal(in.Env)
		if marshalErr != nil {
			return do.SysJob{}, gerror.Wrap(marshalErr, "序列化 Shell 环境变量失败")
		}
		envJSON = string(envData)
		handlerRef = ""
		paramsJSON = ""
	}

	overrideJSON, err := normalizeRetentionOptionJSON(in.LogRetentionOverride)
	if err != nil {
		return do.SysJob{}, err
	}
	if err = s.ensureJobNameUnique(ctx, existing, in.GroupID, name); err != nil {
		return do.SysJob{}, err
	}

	record := do.SysJob{
		GroupId:              in.GroupID,
		Name:                 name,
		Description:          strings.TrimSpace(in.Description),
		TaskType:             string(in.TaskType),
		HandlerRef:           handlerRef,
		Params:               paramsJSON,
		TimeoutSeconds:       int(in.Timeout.Seconds()),
		ShellCmd:             shellCmd,
		WorkDir:              workDir,
		Env:                  envJSON,
		CronExpr:             strings.TrimSpace(in.CronExpr),
		Timezone:             timezone,
		Scope:                string(in.Scope),
		Concurrency:          string(in.Concurrency),
		MaxConcurrency:       maxConcurrency,
		MaxExecutions:        in.MaxExecutions,
		LogRetentionOverride: overrideJSON,
		Status:               string(in.Status),
		UpdatedBy:            s.currentUserID(ctx),
	}
	if existing == nil {
		record.ExecutedCount = 0
		record.StopReason = ""
		record.IsBuiltin = 0
		record.SeedVersion = 0
		record.CreatedBy = s.currentUserID(ctx)
		return record, nil
	}

	record.ExecutedCount = existing.ExecutedCount
	record.StopReason = existing.StopReason
	record.IsBuiltin = existing.IsBuiltin
	record.SeedVersion = existing.SeedVersion
	record.CreatedBy = existing.CreatedBy
	return record, nil
}

// normalizeRetentionOptionJSON validates one optional retention override and serializes it for persistence.
func normalizeRetentionOptionJSON(option *jobmeta.RetentionOption) (string, error) {
	if option == nil {
		return "", nil
	}
	if !option.Mode.IsValid() {
		return "", gerror.New("任务日志保留策略模式不受支持")
	}
	if option.Mode != jobmeta.RetentionModeNone && option.Value <= 0 {
		return "", gerror.New("任务日志保留策略阈值必须大于0")
	}
	return jobmeta.MustMarshalRetentionOption(option)
}

// ensureJobNameUnique verifies the job name stays unique within its group.
func (s *serviceImpl) ensureJobNameUnique(
	ctx context.Context,
	existing *entity.SysJob,
	groupID uint64,
	name string,
) error {
	model := dao.SysJob.Ctx(ctx).Where(do.SysJob{GroupId: groupID, Name: name})
	if existing != nil {
		model = model.WhereNot(dao.SysJob.Columns().Id, existing.Id)
	}
	count, err := model.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return gerror.New("任务名称在当前分组下已存在")
	}
	return nil
}

// validateBuiltinMutableFields enforces built-in job field locks.
func validateBuiltinMutableFields(existing *entity.SysJob, in SaveJobInput) error {
	if existing == nil {
		return nil
	}
	if existing.GroupId != in.GroupID {
		return gerror.New("系统内置任务不允许修改所属分组")
	}
	if strings.TrimSpace(existing.Name) != strings.TrimSpace(in.Name) {
		return gerror.New("系统内置任务不允许修改任务名称")
	}
	if jobmeta.NormalizeTaskType(existing.TaskType) != in.TaskType {
		return gerror.New("系统内置任务不允许修改任务类型")
	}
	if strings.TrimSpace(existing.HandlerRef) != strings.TrimSpace(in.HandlerRef) {
		return gerror.New("系统内置任务不允许修改处理器引用")
	}
	if strings.TrimSpace(existing.Params) != marshalInputMap(in.Params) {
		return gerror.New("系统内置任务不允许修改处理器参数")
	}
	if jobmeta.NormalizeJobScope(existing.Scope) != in.Scope {
		return gerror.New("系统内置任务不允许修改调度范围")
	}
	if jobmeta.NormalizeJobConcurrency(existing.Concurrency) != in.Concurrency {
		return gerror.New("系统内置任务不允许修改并发策略")
	}
	return nil
}

// marshalInputMap serializes one map for built-in field comparison.
func marshalInputMap(values map[string]any) string {
	data, err := json.Marshal(values)
	if err != nil {
		return ""
	}
	return string(data)
}
