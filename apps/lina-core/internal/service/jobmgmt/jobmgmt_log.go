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
func (s *serviceImpl) GetLog(ctx context.Context, id uint64) (*LogDetailOutput, error) {
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

// ClearLogs deletes matching execution logs.
func (s *serviceImpl) ClearLogs(ctx context.Context, jobID *uint64, ids string) error {
	model := dao.SysJobLog.Ctx(ctx)
	logIDs := parseUint64IDs(ids)
	cols := dao.SysJobLog.Columns()

	switch {
	case len(logIDs) > 0:
		model = model.WhereIn(cols.Id, logIDs)
		if err := s.ensureLogsVisible(ctx, logIDs); err != nil {
			return err
		}
	case jobID != nil && *jobID > 0:
		if err := s.ensureJobsVisibleByID(ctx, []uint64{*jobID}); err != nil {
			return err
		}
		model = model.Where(do.SysJobLog{JobId: *jobID})
	default:
		var err error
		model, err = s.applyJobLogDataScope(ctx, model)
		if err != nil {
			return err
		}
		// GoFrame blocks DELETE without WHERE by default, so explicit full-table
		// cleanup must still provide a tautology condition.
		model = model.Where("1 = 1")
	}

	_, err := model.Delete()
	return err
}

// CancelLog cancels one currently running execution instance.
func (s *serviceImpl) CancelLog(ctx context.Context, id uint64) error {
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
	globalCfg, err := s.configSvc.GetCronLogRetention(ctx)
	if err != nil {
		return 0, err
	}
	globalOption := &jobmeta.RetentionOption{
		Mode:  jobmeta.NormalizeRetentionMode(string(globalCfg.Mode)),
		Value: globalCfg.Value,
	}

	var jobs []*entity.SysJob
	if err := dao.SysJob.Ctx(ctx).Scan(&jobs); err != nil {
		return 0, err
	}

	var deletedTotal int64
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

// cleanupJobLogsByPolicy applies one retention policy to one job's logs.
func (s *serviceImpl) cleanupJobLogsByPolicy(
	ctx context.Context,
	jobID uint64,
	policy *jobmeta.RetentionOption,
) (int64, error) {
	if policy == nil || policy.Mode == jobmeta.RetentionModeNone {
		return 0, nil
	}

	cols := dao.SysJobLog.Columns()
	switch policy.Mode {
	case jobmeta.RetentionModeDays:
		result, err := dao.SysJobLog.Ctx(ctx).
			Where(do.SysJobLog{JobId: jobID}).
			WhereLT(cols.StartAt, time.Now().AddDate(0, 0, -int(policy.Value)).Format("2006-01-02 15:04:05")).
			Delete()
		if err != nil {
			return 0, err
		}
		return result.RowsAffected()

	case jobmeta.RetentionModeCount:
		var rows []*entity.SysJobLog
		if err := dao.SysJobLog.Ctx(ctx).
			Where(do.SysJobLog{JobId: jobID}).
			Fields(cols.Id, cols.StartAt).
			OrderDesc(cols.StartAt).
			OrderDesc(cols.Id).
			Scan(&rows); err != nil {
			return 0, err
		}
		if int64(len(rows)) <= policy.Value {
			return 0, nil
		}

		deleteIDs := make([]uint64, 0, len(rows)-int(policy.Value))
		for _, row := range rows[policy.Value:] {
			if row == nil {
				continue
			}
			deleteIDs = append(deleteIDs, row.Id)
		}
		if len(deleteIDs) == 0 {
			return 0, nil
		}
		result, err := dao.SysJobLog.Ctx(ctx).
			WhereIn(cols.Id, deleteIDs).
			Delete()
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
) (map[uint64]logJobDisplay, error) {
	jobIDs := make([]uint64, 0, len(logs))
	for _, logRow := range logs {
		if logRow == nil || logRow.JobId == 0 {
			continue
		}
		jobIDs = append(jobIDs, logRow.JobId)
	}
	if len(jobIDs) == 0 {
		return map[uint64]logJobDisplay{}, nil
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

	result := make(map[uint64]logJobDisplay, len(jobs))
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
	jobs map[uint64]logJobDisplay,
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
