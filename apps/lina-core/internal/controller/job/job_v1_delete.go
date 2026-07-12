// This file maps the scheduled job delete API to the job management service.

package job

import (
	"context"

	v1 "lina-core/api/job/v1"
)

// Delete deletes scheduled jobs by ID list.
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	if err = c.jobMgmtSvc.DeleteJobs(ctx, req.Ids); err != nil {
		return nil, err
	}
	return &v1.DeleteRes{}, nil
}
