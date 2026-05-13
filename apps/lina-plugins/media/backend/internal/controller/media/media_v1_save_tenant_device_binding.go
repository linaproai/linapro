// This file implements the tenant-device strategy binding save controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// SaveTenantDeviceBinding creates or updates one tenant-device strategy binding.
func (c *ControllerV1) SaveTenantDeviceBinding(ctx context.Context, req *v1.SaveTenantDeviceBindingReq) (res *v1.SaveTenantDeviceBindingRes, err error) {
	out, err := c.mediaSvc.SaveTenantDeviceBinding(ctx, mediasvc.TenantDeviceBindingMutationInput{
		TenantId:   req.TenantId,
		DeviceId:   req.DeviceId,
		StrategyId: req.StrategyId,
	})
	if err != nil {
		return nil, err
	}
	return &v1.SaveTenantDeviceBindingRes{
		TenantId:   out.TenantId,
		DeviceId:   out.DeviceId,
		StrategyId: out.StrategyId,
	}, nil
}
