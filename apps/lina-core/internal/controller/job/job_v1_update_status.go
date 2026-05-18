// This file implements the scheduled job status update endpoint.

package job

import (
	"context"

	"lina-core/api/job/v1"
	"lina-core/internal/service/jobmeta"
)

// UpdateStatus handles scheduled job status change requests.
func (c *ControllerV1) UpdateStatus(ctx context.Context, req *v1.UpdateStatusReq) (res *v1.UpdateStatusRes, err error) {
	if err = c.jobMgmtSvc.UpdateJobStatus(ctx, req.Id, jobmeta.NormalizeJobStatus(req.Status)); err != nil {
		return nil, err
	}
	return &v1.UpdateStatusRes{}, nil
}
