// This file implements scheduled-job CRUD and validation logic.

package jobmgmt

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/jobhandler"
	"lina-core/internal/service/jobmeta"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/gdbutil"
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
		keywordPattern := "%" + keyword + "%"
		keywordFilter := model.Builder().
			WhereLike(cols.Name, keywordPattern).
			WhereOrLike(cols.Description, keywordPattern)
		handlerRefs, matchErr := s.localizedHandlerRefsMatchingKeyword(ctx, keyword)
		if matchErr != nil {
			return nil, matchErr
		}
		if len(handlerRefs) > 0 {
			keywordFilter = keywordFilter.WhereOrIn(cols.HandlerRef, handlerRefs)
		}
		model = model.Where(keywordFilter)
	}
	var scopeErr error
	model, scopeErr = s.applyJobDataScope(ctx, model)
	if scopeErr != nil {
		return nil, scopeErr
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
		map[orderField]string{
			orderFieldID:        cols.Id,
			orderFieldName:      cols.Name,
			orderFieldGroupID:   cols.GroupId,
			orderFieldStatus:    cols.Status,
			orderFieldTaskType:  cols.TaskType,
			orderFieldCreatedAt: cols.CreatedAt,
			orderFieldUpdatedAt: cols.UpdatedAt,
		},
		cols.UpdatedAt,
		gdbutil.OrderDirectionDESC,
	).Page(in.PageNum, in.PageSize).Scan(&jobs)
	if err != nil {
		return nil, err
	}

	groupMap, err := s.groupMapByJobGroupIDs(ctx, jobs)
	if err != nil {
		return nil, err
	}
	i18nCache := make(handlerSourceTextCache)
	for _, group := range groupMap {
		s.localizeGroupForDisplay(ctx, group)
	}
	items := make([]*JobListItem, 0, len(jobs))
	for _, job := range jobs {
		if job == nil {
			continue
		}
		group := groupMap[job.GroupId]
		s.localizeBuiltinJobForDisplayWithCache(ctx, job, i18nCache)
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
		return nil, bizerr.NewCode(jobmeta.CodeJobNotFound)
	}
	if err = s.ensureJobVisible(ctx, job); err != nil {
		return nil, err
	}

	group, err := s.groupByID(ctx, job.GroupId)
	if err != nil {
		return nil, err
	}

	out := &JobDetailOutput{SysJob: job}
	s.localizeBuiltinJobForDisplay(ctx, job)
	if group != nil {
		s.localizeGroupForDisplay(ctx, group)
		out.GroupCode = group.Code
		out.GroupName = group.Name
	}
	return out, nil
}

// CreateJob persists one new scheduled job and refreshes the scheduler when needed.
func (s *serviceImpl) CreateJob(ctx context.Context, in SaveJobInput) (uint64, error) {
	if in.TaskType != jobmeta.TaskTypeShell {
		return 0, bizerr.NewCode(CodeJobCreateShellOnly)
	}
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
		return bizerr.NewCode(jobmeta.CodeJobNotFound)
	}
	if existing.IsBuiltin == 1 {
		return bizerr.NewCode(CodeJobBuiltinUpdateDenied)
	}
	if err = s.ensureJobVisible(ctx, existing); err != nil {
		return err
	}
	if in.TaskType != jobmeta.TaskTypeShell {
		return bizerr.NewCode(CodeJobUpdateShellOnly)
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
		return bizerr.NewCode(CodeJobDeleteRequired)
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
			return bizerr.NewCode(CodeJobBuiltinDeleteDenied)
		}
		if err = s.ensureJobVisible(ctx, job); err != nil {
			return err
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
		return do.SysJob{}, bizerr.NewCode(CodeJobGroupNotFound)
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		return do.SysJob{}, bizerr.NewCode(CodeJobNameRequired)
	}
	if len(name) > 128 {
		return do.SysJob{}, bizerr.NewCode(CodeJobNameTooLong)
	}
	if in.Timeout%time.Second != 0 {
		return do.SysJob{}, bizerr.NewCode(CodeJobTimeoutSecondAlignedRequired)
	}
	if in.Timeout <= 0 {
		return do.SysJob{}, bizerr.NewCode(CodeJobTimeoutOutOfRange)
	}
	if in.Timeout > 24*time.Hour {
		return do.SysJob{}, bizerr.NewCode(CodeJobTimeoutOutOfRange)
	}
	cronExpr, _, err := normalizeCronExpression(in.CronExpr)
	if err != nil {
		return do.SysJob{}, err
	}
	timezone, _, err := normalizeJobTimezone(in.Timezone)
	if err != nil {
		return do.SysJob{}, err
	}
	if !in.TaskType.IsValid() {
		return do.SysJob{}, bizerr.NewCode(CodeJobTaskTypeInvalid)
	}
	if !in.Scope.IsValid() {
		return do.SysJob{}, bizerr.NewCode(CodeJobScopeInvalid)
	}
	if !in.Concurrency.IsValid() {
		return do.SysJob{}, bizerr.NewCode(CodeJobConcurrencyInvalid)
	}
	if !in.Status.IsValid() || in.Status == jobmeta.JobStatusPausedByPlugin {
		return do.SysJob{}, bizerr.NewCode(CodeJobStatusInvalid)
	}
	if in.MaxExecutions < 0 {
		return do.SysJob{}, bizerr.NewCode(CodeJobMaxExecutionsInvalid)
	}

	maxConcurrency := in.MaxConcurrency
	if in.Concurrency == jobmeta.JobConcurrencySingleton {
		maxConcurrency = 1
	}
	if maxConcurrency <= 0 || maxConcurrency > 100 {
		return do.SysJob{}, bizerr.NewCode(CodeJobMaxConcurrencyInvalid)
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
			return do.SysJob{}, bizerr.NewCode(jobhandler.CodeJobHandlerNotFound)
		}
		paramsData, marshalErr := json.Marshal(in.Params)
		if marshalErr != nil {
			return do.SysJob{}, bizerr.WrapCode(marshalErr, CodeJobParamsMarshalFailed)
		}
		if err = jobhandler.ValidateParams(def.ParamsSchema, paramsData); err != nil {
			return do.SysJob{}, err
		}
		paramsJSON = string(paramsData)
		shellCmd = ""
		workDir = ""

	case jobmeta.TaskTypeShell:
		shellEnabled, shellErr := s.configSvc.IsCronShellEnabled(ctx)
		if shellErr != nil {
			return do.SysJob{}, shellErr
		}
		if !shellEnabled {
			return do.SysJob{}, bizerr.NewCode(jobmeta.CodeJobShellDisabled)
		}
		if shellCmd == "" {
			return do.SysJob{}, bizerr.NewCode(jobmeta.CodeJobShellCommandRequired)
		}
		if err = validateWorkDir(workDir); err != nil {
			return do.SysJob{}, err
		}
		envData, marshalErr := json.Marshal(in.Env)
		if marshalErr != nil {
			return do.SysJob{}, bizerr.WrapCode(marshalErr, CodeJobShellEnvMarshalFailed)
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
		CronExpr:             cronExpr,
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
		return "", bizerr.NewCode(jobmeta.CodeJobRetentionModeUnsupported)
	}
	if option.Mode != jobmeta.RetentionModeNone && option.Value <= 0 {
		return "", bizerr.NewCode(jobmeta.CodeJobRetentionValueInvalid)
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
		return bizerr.NewCode(CodeJobNameExistsInGroup)
	}
	return nil
}
