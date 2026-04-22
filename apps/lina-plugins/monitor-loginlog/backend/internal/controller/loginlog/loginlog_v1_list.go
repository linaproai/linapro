package loginlog

import (
	"context"

	v1 "lina-plugin-monitor-loginlog/backend/api/loginlog/v1"
	loginlogsvc "lina-plugin-monitor-loginlog/backend/internal/service/loginlog"
)

// List returns the paginated login-log list.
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

	items := make([]*v1.LoginLogEntity, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, toAPILoginLogEntity(item))
	}
	return &v1.ListRes{Items: items, Total: out.Total}, nil
}
