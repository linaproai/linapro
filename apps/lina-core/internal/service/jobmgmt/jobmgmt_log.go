// This file implements execution-log listing, detail, cleanup, and cancellation.

package jobmgmt

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/jobmeta"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/gdbutil"
)

// ListLogs returns scheduled-job execution logs with pagination and job metadata.
func (s *serviceImpl) ListLogs(ctx context.Context, in ListLogsInput) (*ListLogsOutput, error) {
	model := dao.SysJobLog.Ctx(ctx)
	cols := dao.SysJobLog.Columns()

	if in.JobID != nil && *in.JobID > 0 {
		model = model.Where(cols.JobId, *in.JobID)
	}
	if in.Status.IsValid() {
		model = model.Where(cols.Status, string(in.Status))
	}
	if in.Trigger.IsValid() {
		model = model.Where(cols.Trigger, string(in.Trigger))
	}
	if nodeID := strings.TrimSpace(in.NodeID); nodeID != "" {
		model = model.Where(cols.NodeId, nodeID)
	}
	if beginTime := strings.TrimSpace(in.BeginTime); beginTime != "" {
		model = model.WhereGTE(cols.StartAt, beginTime)
	}
	if endTime := strings.TrimSpace(in.EndTime); endTime != "" {
		model = model.WhereLTE(cols.StartAt, endTime)
	}
	var scopeErr error
	model, scopeErr = s.applyJobLogDataScope(ctx, model)
	if scopeErr != nil {
		return nil, scopeErr
	}

	total, err := model.Count()
	if err != nil {
		return nil, err
	}

	var logs []*entity.SysJobLog
	err = applySingleOrder(
		model,
		in.OrderBy,
		in.OrderDirection,
		map[orderField]string{
			orderFieldID:         cols.Id,
			orderFieldStartAt:    cols.StartAt,
			orderFieldEndAt:      cols.EndAt,
			orderFieldDurationMs: cols.DurationMs,
			orderFieldStatus:     cols.Status,
			orderFieldCreatedAt:  cols.CreatedAt,
		},
		cols.StartAt,
		gdbutil.OrderDirectionDESC,
	).Page(in.PageNum, in.PageSize).Scan(&logs)
	if err != nil {
		return nil, err
	}

	jobMap, err := s.jobDisplayMapByLogs(ctx, logs)
	if err != nil {
		return nil, err
	}
	i18nCache := make(handlerSourceTextCache)
	items := make([]*LogListItem, 0, len(logs))
	for _, logRow := range logs {
		if logRow == nil {
			continue
		}
		items = append(items, &LogListItem{
			SysJobLog: logRow,
			JobName:   s.resolveLogJobName(ctx, logRow, jobMap, i18nCache),
		})
	}
	return &ListLogsOutput{List: items, Total: total}, nil
}

// GetLog returns one execution-log detail snapshot.
func (s *serviceImpl) GetLog(ctx context.Context, id int64) (*LogDetailOutput, error) {
	var logRow *entity.SysJobLog
	err := dao.SysJobLog.Ctx(ctx).
		Where(do.SysJobLog{Id: id}).
		Scan(&logRow)
	if err != nil {
		return nil, err
	}
	if logRow == nil {
		return nil, bizerr.NewCode(CodeJobLogNotFound)
	}
	if err = s.ensureLogVisible(ctx, logRow); err != nil {
		return nil, err
	}

	jobMap, err := s.jobDisplayMapByLogs(ctx, []*entity.SysJobLog{logRow})
	if err != nil {
		return nil, err
	}
	return &LogDetailOutput{
		SysJobLog: logRow,
		JobName:   s.resolveLogJobName(ctx, logRow, jobMap, make(handlerSourceTextCache)),
	}, nil
}

// ClearLogs deletes matching execution logs and returns the deleted row count.
func (s *serviceImpl) ClearLogs(ctx context.Context, in ClearLogsInput) (int64, error) {
	model := dao.SysJobLog.Ctx(ctx)
	logIDs := parseInt64IDs(in.IDs)
	cols := dao.SysJobLog.Columns()
	beginTime := strings.TrimSpace(in.BeginTime)
	endTime := strings.TrimSpace(in.EndTime)

	switch {
	case len(logIDs) > 0:
		model = model.WhereIn(cols.Id, logIDs)
		if err := s.ensureLogsVisible(ctx, logIDs); err != nil {
			return 0, err
		}
	case in.JobID != nil && *in.JobID > 0:
		if err := s.ensureJobsVisibleByID(ctx, []int64{*in.JobID}); err != nil {
			return 0, err
		}
		model = model.Where(do.SysJobLog{JobId: *in.JobID})
		var err error
		model, err = s.applyJobLogDataScope(ctx, model)
		if err != nil {
			return 0, err
		}
	default:
		var err error
		model, err = s.applyJobLogDataScope(ctx, model)
		if err != nil {
			return 0, err
		}
	}
	if len(logIDs) == 0 && beginTime != "" {
		model = model.WhereGTE(cols.StartAt, beginTime)
	}
	if len(logIDs) == 0 && endTime != "" {
		model = model.WhereLTE(cols.StartAt, normalizeLogCleanupEndTime(endTime))
	}
	if len(logIDs) == 0 && (in.JobID == nil || *in.JobID <= 0) &&
		beginTime == "" && endTime == "" {
		// GoFrame blocks DELETE without WHERE by default, so explicit full-scope
		// cleanup must still provide a tautology condition.
		model = model.Where("1 = 1")
	}

	result, err := model.Delete()
	if err != nil {
		return 0, err
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return deleted, nil
}

// normalizeLogCleanupEndTime expands date-only end values to the end of day.
func normalizeLogCleanupEndTime(value string) string {
	if len(value) == 10 {
		return value + " 23:59:59"
	}
	return value
}

// CancelLog cancels one currently running execution instance.
func (s *serviceImpl) CancelLog(ctx context.Context, id int64) error {
	if s.scheduler == nil {
		return bizerr.NewCode(CodeJobSchedulerUninitialized)
	}
	var logRow *entity.SysJobLog
	if err := dao.SysJobLog.Ctx(ctx).
		Where(do.SysJobLog{Id: id}).
		Scan(&logRow); err != nil {
		return err
	}
	if logRow == nil {
		return bizerr.NewCode(CodeJobLogNotFound)
	}
	if err := s.ensureLogVisible(ctx, logRow); err != nil {
		return err
	}
	return s.scheduler.CancelLog(ctx, id)
}

// CleanupDueLogs removes logs that exceed the effective retention policies.
func (s *serviceImpl) CleanupDueLogs(ctx context.Context) (int64, error) {
	logRetentionDays, err := s.configSvc.GetLogRetentionDays(ctx)
	if err != nil {
		return 0, err
	}
	deletedTotal, err := s.cleanupJobLogsByGlobalRetention(ctx, logRetentionDays)
	if err != nil {
		return deletedTotal, err
	}

	globalCfg, err := s.configSvc.GetCronLogRetention(ctx)
	if err != nil {
		return deletedTotal, err
	}
	globalOption := &jobmeta.RetentionOption{
		Mode:  jobmeta.NormalizeRetentionMode(string(globalCfg.Mode)),
		Value: globalCfg.Value,
	}

	var jobs []*entity.SysJob
	model := dao.SysJob.Ctx(ctx)
	model = datascope.ApplyTenantScope(ctx, model, datascope.TenantColumn)
	if err := model.Scan(&jobs); err != nil {
		return deletedTotal, err
	}

	for _, job := range jobs {
		if job == nil {
			continue
		}
		policy := globalOption
		if override, err := jobmeta.ParseRetentionOption(job.LogRetentionOverride); err != nil {
			return deletedTotal, err
		} else if override != nil {
			policy = override
		}

		deleted, err := s.cleanupJobLogsByPolicy(ctx, job.Id, policy)
		if err != nil {
			return deletedTotal, err
		}
		deletedTotal += deleted
	}
	return deletedTotal, nil
}

// cleanupJobLogsByGlobalRetention enforces the system-wide maximum retention
// period before per-job retention policies apply stricter cleanup.
func (s *serviceImpl) cleanupJobLogsByGlobalRetention(ctx context.Context, days int64) (int64, error) {
	if days <= 0 {
		return 0, nil
	}
	cols := dao.SysJobLog.Columns()
	model := dao.SysJobLog.Ctx(ctx).
		WhereLT(cols.StartAt, time.Now().AddDate(0, 0, -int(days)).Format("2006-01-02 15:04:05"))
	model = datascope.ApplyTenantScope(ctx, model, datascope.TenantColumn)
	result, err := model.Delete()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// cleanupJobLogsByPolicy applies one retention policy to one job's logs.
func (s *serviceImpl) cleanupJobLogsByPolicy(
	ctx context.Context,
	jobID int64,
	policy *jobmeta.RetentionOption,
) (int64, error) {
	if policy == nil || policy.Mode == jobmeta.RetentionModeNone {
		return 0, nil
	}

	cols := dao.SysJobLog.Columns()
	switch policy.Mode {
	case jobmeta.RetentionModeDays:
		model := dao.SysJobLog.Ctx(ctx).
			Where(do.SysJobLog{JobId: jobID}).
			WhereLT(cols.StartAt, time.Now().AddDate(0, 0, -int(policy.Value)).Format("2006-01-02 15:04:05"))
		model = datascope.ApplyTenantScope(ctx, model, datascope.TenantColumn)
		result, err := model.Delete()
		if err != nil {
			return 0, err
		}
		return result.RowsAffected()

	case jobmeta.RetentionModeCount:
		var rows []*entity.SysJobLog
		model := dao.SysJobLog.Ctx(ctx).
			Where(do.SysJobLog{JobId: jobID}).
			Fields(cols.Id, cols.StartAt).
			OrderDesc(cols.StartAt).
			OrderDesc(cols.Id)
		model = datascope.ApplyTenantScope(ctx, model, datascope.TenantColumn)
		if err := model.Scan(&rows); err != nil {
			return 0, err
		}
		if int64(len(rows)) <= policy.Value {
			return 0, nil
		}

		deleteIDs := make([]int64, 0, len(rows)-int(policy.Value))
		for _, row := range rows[policy.Value:] {
			if row == nil {
				continue
			}
			deleteIDs = append(deleteIDs, row.Id)
		}
		if len(deleteIDs) == 0 {
			return 0, nil
		}
		deleteModel := dao.SysJobLog.Ctx(ctx).WhereIn(cols.Id, deleteIDs)
		deleteModel = datascope.ApplyTenantScope(ctx, deleteModel, datascope.TenantColumn)
		result, err := deleteModel.Delete()
		if err != nil {
			return 0, err
		}
		return result.RowsAffected()
	}
	return 0, nil
}

// logJobDisplay stores the live display anchors needed by execution-log rows.
type logJobDisplay struct {
	Name       string // Name stores the current persisted job name.
	HandlerRef string // HandlerRef stores the stable handler anchor.
	IsBuiltin  int    // IsBuiltin identifies code-owned jobs.
}

// jobDisplayMapByLogs loads job display anchors for the given logs.
func (s *serviceImpl) jobDisplayMapByLogs(
	ctx context.Context,
	logs []*entity.SysJobLog,
) (map[int64]logJobDisplay, error) {
	jobIDs := make([]int64, 0, len(logs))
	for _, logRow := range logs {
		if logRow == nil || logRow.JobId == 0 {
			continue
		}
		jobIDs = append(jobIDs, logRow.JobId)
	}
	if len(jobIDs) == 0 {
		return map[int64]logJobDisplay{}, nil
	}

	var jobs []*entity.SysJob
	err := dao.SysJob.Ctx(ctx).
		WhereIn(dao.SysJob.Columns().Id, jobIDs).
		Fields(
			dao.SysJob.Columns().Id,
			dao.SysJob.Columns().Name,
			dao.SysJob.Columns().HandlerRef,
			dao.SysJob.Columns().IsBuiltin,
		).
		Scan(&jobs)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]logJobDisplay, len(jobs))
	for _, job := range jobs {
		if job == nil {
			continue
		}
		result[job.Id] = logJobDisplay{
			Name:       job.Name,
			HandlerRef: job.HandlerRef,
			IsBuiltin:  job.IsBuiltin,
		}
	}
	return result, nil
}

// resolveLogJobName resolves one log row's job name from live data or the stored job snapshot.
func (s *serviceImpl) resolveLogJobName(
	ctx context.Context,
	logRow *entity.SysJobLog,
	jobs map[int64]logJobDisplay,
	cache handlerSourceTextCache,
) string {
	if logRow == nil {
		return ""
	}
	if job, ok := jobs[logRow.JobId]; ok && job.Name != "" {
		return s.localizeBuiltinJobNameWithCache(ctx, job.HandlerRef, job.Name, job.IsBuiltin, cache)
	}

	var snapshot struct {
		Name       string `json:"name"`
		HandlerRef string `json:"handlerRef"`
		IsBuiltin  int    `json:"isBuiltin"`
	}
	if err := json.Unmarshal([]byte(logRow.JobSnapshot), &snapshot); err != nil {
		return ""
	}
	return s.localizeBuiltinJobNameWithCache(ctx, snapshot.HandlerRef, snapshot.Name, snapshot.IsBuiltin, cache)
}
