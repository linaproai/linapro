package config

import (
	"context"

	v1 "lina-core/api/config/v1"
)

// Delete deletes the specified config item.
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	err = c.svc.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteRes{}, nil
}
