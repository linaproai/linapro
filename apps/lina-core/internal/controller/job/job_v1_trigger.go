// This file implements the scheduled job manual trigger endpoint.

package job

import (
	"context"

	"lina-core/api/job/v1"
)

// Trigger handles requests that trigger one scheduled job immediately.
func (c *ControllerV1) Trigger(ctx context.Context, req *v1.TriggerReq) (res *v1.TriggerRes, err error) {
	logID, err := c.jobMgmtSvc.TriggerJob(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.TriggerRes{LogId: logID}, nil
}
