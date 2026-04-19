// This file implements execution-log listing, detail, cleanup, and cancellation.

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

	total, err := model.Count()
	if err != nil {
		return nil, err
	}

	var logs []*entity.SysJobLog
	err = applySingleOrder(
		model,
		in.OrderBy,
		in.OrderDirection,
		map[string]string{
			"id":          cols.Id,
			"start_at":    cols.StartAt,
			"end_at":      cols.EndAt,
			"duration_ms": cols.DurationMs,
			"status":      cols.Status,
			"created_at":  cols.CreatedAt,
		},
		cols.StartAt,
		"desc",
	).Page(in.PageNum, in.PageSize).Scan(&logs)
	if err != nil {
		return nil, err
	}

	jobMap, err := s.jobNameMapByLogs(ctx, logs)
	if err != nil {
		return nil, err
	}
	items := make([]*LogListItem, 0, len(logs))
	for _, logRow := range logs {
		if logRow == nil {
			continue
		}
		items = append(items, &LogListItem{
			SysJobLog: logRow,
			JobName:   resolveLogJobName(logRow, jobMap),
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
		return nil, gerror.New("执行日志不存在")
	}

	jobNameMap, err := s.jobNameMapByLogs(ctx, []*entity.SysJobLog{logRow})
	if err != nil {
		return nil, err
	}
	return &LogDetailOutput{
		SysJobLog: logRow,
		JobName:   resolveLogJobName(logRow, jobNameMap),
	}, nil
}

// ClearLogs deletes matching execution logs.
func (s *serviceImpl) ClearLogs(ctx context.Context, jobID *uint64) error {
	model := dao.SysJobLog.Ctx(ctx)
	if jobID != nil && *jobID > 0 {
		model = model.Where(do.SysJobLog{JobId: *jobID})
	}
	_, err := model.Delete()
	return err
}

// CancelLog cancels one currently running execution instance.
func (s *serviceImpl) CancelLog(ctx context.Context, id uint64) error {
	if s.scheduler == nil {
		return gerror.New("定时任务调度器未初始化")
	}
	return s.scheduler.CancelLog(ctx, id)
}

// CleanupDueLogs removes logs that exceed the effective retention policies.
func (s *serviceImpl) CleanupDueLogs(ctx context.Context) (int64, error) {
	globalCfg := s.configSvc.GetCronLogRetention(ctx)
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

// jobNameMapByLogs loads job names for the given logs.
func (s *serviceImpl) jobNameMapByLogs(
	ctx context.Context,
	logs []*entity.SysJobLog,
) (map[uint64]string, error) {
	jobIDs := make([]uint64, 0, len(logs))
	for _, logRow := range logs {
		if logRow == nil || logRow.JobId == 0 {
			continue
		}
		jobIDs = append(jobIDs, logRow.JobId)
	}
	if len(jobIDs) == 0 {
		return map[uint64]string{}, nil
	}

	var jobs []*entity.SysJob
	err := dao.SysJob.Ctx(ctx).
		WhereIn(dao.SysJob.Columns().Id, jobIDs).
		Fields(dao.SysJob.Columns().Id, dao.SysJob.Columns().Name).
		Scan(&jobs)
	if err != nil {
		return nil, err
	}

	result := make(map[uint64]string, len(jobs))
	for _, job := range jobs {
		if job == nil {
			continue
		}
		result[job.Id] = job.Name
	}
	return result, nil
}

// resolveLogJobName resolves one log row's job name from live data or the stored job snapshot.
func resolveLogJobName(logRow *entity.SysJobLog, names map[uint64]string) string {
	if logRow == nil {
		return ""
	}
	if name := names[logRow.JobId]; name != "" {
		return name
	}

	var snapshot struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(logRow.JobSnapshot), &snapshot); err != nil {
		return ""
	}
	return snapshot.Name
}
