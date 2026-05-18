// This file implements the scheduled job delete endpoint.

package job

import (
	"context"

	"lina-core/api/job/v1"
)

// Delete handles scheduled job deletion requests.
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	if err = c.jobMgmtSvc.DeleteJobs(ctx, req.Ids); err != nil {
		return nil, err
	}
	return &v1.DeleteRes{}, nil
}
