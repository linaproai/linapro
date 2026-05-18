// This file implements the scheduled job group update endpoint.

package jobgroup

import (
	"context"

	"lina-core/api/jobgroup/v1"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
)

// Update handles scheduled job group update requests.
func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	err = c.jobMgmtSvc.UpdateGroup(ctx, jobmgmtsvc.UpdateGroupInput{
		ID: req.Id,
		SaveGroupInput: jobmgmtsvc.SaveGroupInput{
			Code:      req.Code,
			Name:      req.Name,
			Remark:    req.Remark,
			SortOrder: req.SortOrder,
		},
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRes{}, nil
}
