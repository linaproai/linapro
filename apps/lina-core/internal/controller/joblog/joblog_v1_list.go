// This file implements the scheduled job log list endpoint.

package joblog

import (
	"context"

	"lina-core/api/joblog/v1"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/jobmeta"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
)

// List handles scheduled job log list requests.
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.jobMgmtSvc.ListLogs(ctx, jobmgmtsvc.ListLogsInput{
		PageNum:        req.PageNum,
		PageSize:       req.PageSize,
		JobID:          req.JobId,
		Status:         jobmeta.NormalizeLogStatus(req.Status),
		Trigger:        jobmeta.NormalizeTriggerType(req.Trigger),
		NodeID:         req.NodeId,
		BeginTime:      req.BeginTime,
		EndTime:        req.EndTime,
		OrderBy:        req.OrderBy,
		OrderDirection: req.OrderDirection,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.ListItem, 0, len(out.List))
	for _, item := range out.List {
		if item == nil {
			continue
		}
		items = append(items, &v1.ListItem{
			JobLogItem: jobLogItem(item.SysJobLog),
			JobName:    item.JobName,
		})
	}
	return &v1.ListRes{List: items, Total: out.Total}, nil
}

// jobLogItem maps a scheduled-job log entity to the API-safe response DTO.
func jobLogItem(log *entity.SysJobLog) v1.JobLogItem {
	if log == nil {
		return v1.JobLogItem{}
	}
	return v1.JobLogItem{
		Id:             log.Id,
		JobId:          log.JobId,
		JobSnapshot:    log.JobSnapshot,
		NodeId:         log.NodeId,
		Trigger:        log.Trigger,
		ParamsSnapshot: log.ParamsSnapshot,
		StartAt:        log.StartAt,
		EndAt:          log.EndAt,
		DurationMs:     log.DurationMs,
		Status:         log.Status,
		ErrMsg:         log.ErrMsg,
		ResultJson:     log.ResultJson,
		CreatedAt:      log.CreatedAt,
	}
}
