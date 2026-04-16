package config

import (
	"context"

	v1 "lina-core/api/config/v1"
)

// ByKey returns the config item for the specified key.
func (c *ControllerV1) ByKey(ctx context.Context, req *v1.ByKeyReq) (res *v1.ByKeyRes, err error) {
	cfg, err := c.svc.GetByKey(ctx, req.Key)
	if err != nil {
		return nil, err
	}
	return &v1.ByKeyRes{SysConfig: cfg}, nil
}
