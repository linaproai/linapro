// This file implements the scheduled job log detail endpoint.

package joblog

import (
	"context"

	"lina-core/api/joblog/v1"
)

// Detail handles scheduled job log detail lookup requests.
func (c *ControllerV1) Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error) {
	out, err := c.jobMgmtSvc.GetLog(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.DetailRes{
		JobLogItem: jobLogItem(out.SysJobLog),
		JobName:    out.JobName,
	}, nil
}
