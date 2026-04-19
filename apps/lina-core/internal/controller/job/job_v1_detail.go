// This file implements the scheduled job detail endpoint.

package job

import (
	"context"

	"lina-core/api/job/v1"
)

// Detail handles scheduled job detail lookup requests.
func (c *ControllerV1) Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error) {
	out, err := c.jobMgmtSvc.GetJob(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.DetailRes{
		SysJob:    out.SysJob,
		GroupCode: out.GroupCode,
		GroupName: out.GroupName,
	}, nil
}
