package config

import (
	"context"

	v1 "lina-core/api/config/v1"
)

// Get returns the detail of the specified config item.
func (c *ControllerV1) Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error) {
	cfg, err := c.svc.GetById(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetRes{SysConfig: cfg}, nil
}
