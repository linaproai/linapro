// This file implements the scheduled job log list endpoint.

package joblog

import (
	"context"

	"lina-core/api/joblog/v1"
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
			SysJobLog: item.SysJobLog,
			JobName:   item.JobName,
		})
	}
	return &v1.ListRes{List: items, Total: out.Total}, nil
}
