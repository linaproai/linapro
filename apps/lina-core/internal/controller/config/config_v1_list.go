package config

import (
	"context"

	v1 "lina-core/api/config/v1"
	"lina-core/internal/service/sysconfig"
)

// List queries config items with pagination and filters.
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.svc.List(ctx, sysconfig.ListInput{
		PageNum:   req.PageNum,
		PageSize:  req.PageSize,
		Name:      req.Name,
		Key:       req.Key,
		BeginTime: req.BeginTime,
		EndTime:   req.EndTime,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ListRes{List: out.List, Total: out.Total}, nil
}
