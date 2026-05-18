// This file implements the scheduled job update endpoint.

package job

import (
	"context"

	"lina-core/api/job/v1"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
)

// Update handles scheduled job update requests.
func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	err = c.jobMgmtSvc.UpdateJob(ctx, jobmgmtsvc.UpdateJobInput{
		ID:           req.Id,
		SaveJobInput: buildSaveJobInput(req.JobPayload),
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRes{}, nil
}
