package loginlog

import (
	"context"

	v1 "lina-core/api/loginlog/v1"
	loginlogsvc "lina-core/internal/service/loginlog"
)

// List returns login log list.
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.loginLogSvc.List(ctx, loginlogsvc.ListInput{
		PageNum:        req.PageNum,
		PageSize:       req.PageSize,
		UserName:       req.UserName,
		Ip:             req.Ip,
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
