// This file implements the scheduled job reset endpoint.

package job

import (
	"context"

	"lina-core/api/job/v1"
)

// Reset handles requests that reset scheduled job execution counters.
func (c *ControllerV1) Reset(ctx context.Context, req *v1.ResetReq) (res *v1.ResetRes, err error) {
	if err = c.jobMgmtSvc.ResetJob(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.ResetRes{}, nil
}
