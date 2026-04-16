package dept

import (
	"context"

	v1 "lina-core/api/dept/v1"
	deptsvc "lina-core/internal/service/dept"
)

// Update updates a department.
func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	return nil, c.deptSvc.Update(ctx, deptsvc.UpdateInput{
		Id:       req.Id,
		ParentId: req.ParentId,
		Name:     req.Name,
		Code:     req.Code,
		OrderNum: req.OrderNum,
		Leader:   req.Leader,
		Phone:    req.Phone,
		Email:    req.Email,
		Status:   req.Status,
		Remark:   req.Remark,
	})
}
