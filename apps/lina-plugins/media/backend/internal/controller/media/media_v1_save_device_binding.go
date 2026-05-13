// This file implements the device strategy binding save controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// SaveDeviceBinding creates or updates one device strategy binding.
func (c *ControllerV1) SaveDeviceBinding(ctx context.Context, req *v1.SaveDeviceBindingReq) (res *v1.SaveDeviceBindingRes, err error) {
	out, err := c.mediaSvc.SaveDeviceBinding(ctx, mediasvc.DeviceBindingMutationInput{
		DeviceId:   req.DeviceId,
		StrategyId: req.StrategyId,
	})
	if err != nil {
		return nil, err
	}
	return &v1.SaveDeviceBindingRes{
		DeviceId:   out.DeviceId,
		StrategyId: out.StrategyId,
	}, nil
}
