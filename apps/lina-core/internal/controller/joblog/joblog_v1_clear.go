// This file implements the scheduled job log clear endpoint.

package joblog

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/api/joblog/v1"
)

// Clear handles scheduled job log cleanup requests.
func (c *ControllerV1) Clear(ctx context.Context, req *v1.ClearReq) (res *v1.ClearRes, err error) {
	logIDs := req.LogIds
	if logIDs == "" {
		logIDs = g.RequestFromCtx(ctx).Get("logIds").String()
	}
	if err = c.jobMgmtSvc.ClearLogs(ctx, req.JobId, logIDs); err != nil {
		return nil, err
	}
	return &v1.ClearRes{}, nil
}
