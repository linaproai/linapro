package config

import (
	"context"

	v1 "lina-core/api/config/v1"
	"lina-core/internal/service/sysconfig"
)

// Update updates the specified config item.
func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	err = c.svc.Update(ctx, sysconfig.UpdateInput{
		Id:     req.Id,
		Name:   req.Name,
		Key:    req.Key,
		Value:  req.Value,
		Remark: req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRes{}, nil
}
