// This file implements the scheduled job create endpoint.

package job

import (
	"context"

	"lina-core/api/job/v1"
)

// Create handles scheduled job creation requests.
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	id, err := c.jobMgmtSvc.CreateJob(ctx, buildSaveJobInput(req.JobPayload))
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{Id: id}, nil
}
