package operlog

import (
	"context"

	v1 "lina-core/api/operlog/v1"
	operlogsvc "lina-core/internal/service/operlog"
)

// List queries operation log list
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.operLogSvc.List(ctx, operlogsvc.ListInput{
		PageNum:        req.PageNum,
		PageSize:       req.PageSize,
		Title:          req.Title,
		OperName:       req.OperName,
		OperType:       req.OperType,
		Status:         req.Status,
		BeginTime:      req.BeginTime,
		EndTime:        req.EndTime,
		OrderBy:        req.OrderBy,
		OrderDirection: req.OrderDirection,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ListRes{
		Items: out.List,
		Total: out.Total,
	}, nil
}
