package dept

import (
	"context"

	v1 "lina-core/api/dept/v1"
	deptsvc "lina-core/internal/service/dept"
)

// Create creates a new department.
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	orderNum := 0
	if req.OrderNum != nil {
		orderNum = *req.OrderNum
	}
	leader := 0
	if req.Leader != nil {
		leader = *req.Leader
	}
	status := 1
	if req.Status != nil {
		status = *req.Status
	}
	id, err := c.deptSvc.Create(ctx, deptsvc.CreateInput{
		ParentId: req.ParentId,
		Name:     req.Name,
		Code:     req.Code,
		OrderNum: orderNum,
		Leader:   leader,
		Phone:    req.Phone,
		Email:    req.Email,
		Status:   status,
		Remark:   req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{Id: id}, nil
}
