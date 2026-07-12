// This file implements the scheduled job log clear endpoint.

package joblog

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/joblog/v1"
	jobmgmt "lina-core/internal/service/jobmgmt"
)

// Clear handles scheduled job log cleanup requests.
func (c *ControllerV1) Clear(ctx context.Context, req *v1.ClearReq) (res *v1.ClearRes, err error) {
	beginTime := req.BeginTime
	if beginTime == "" {
		if r := g.RequestFromCtx(ctx); r != nil {
			beginTime = r.Get("beginTime").String()
		}
	}
	endTime := req.EndTime
	if endTime == "" {
		if r := g.RequestFromCtx(ctx); r != nil {
			endTime = r.Get("endTime").String()
		}
	}
	deleted, err := c.jobMgmtSvc.ClearLogs(ctx, jobmgmt.ClearLogsInput{
		JobID:     req.JobId,
		IDs:       req.LogIds,
		BeginTime: beginTime,
		EndTime:   endTime,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ClearRes{Deleted: deleted}, nil
}
