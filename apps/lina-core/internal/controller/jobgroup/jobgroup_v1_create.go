// This file implements the scheduled job group create endpoint.

package jobgroup

import (
	"context"

	"lina-core/api/jobgroup/v1"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
)

// Create handles scheduled job group creation requests.
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	id, err := c.jobMgmtSvc.CreateGroup(ctx, jobmgmtsvc.SaveGroupInput{
		Code:      req.Code,
		Name:      req.Name,
		Remark:    req.Remark,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{Id: id}, nil
}
