// This file implements the scheduled job group delete endpoint.

package jobgroup

import (
	"context"

	"lina-core/api/jobgroup/v1"
)

// Delete handles scheduled job group deletion requests.
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	if err = c.jobMgmtSvc.DeleteGroups(ctx, req.Ids); err != nil {
		return nil, err
	}
	return &v1.DeleteRes{}, nil
}
