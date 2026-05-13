// This file implements the tenant-device strategy binding deletion controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
)

// DeleteTenantDeviceBinding deletes one tenant-device strategy binding.
func (c *ControllerV1) DeleteTenantDeviceBinding(ctx context.Context, req *v1.DeleteTenantDeviceBindingReq) (res *v1.DeleteTenantDeviceBindingRes, err error) {
	out, err := c.mediaSvc.DeleteTenantDeviceBinding(ctx, req.TenantId, req.DeviceId)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteTenantDeviceBindingRes{
		TenantId: out.TenantId,
		DeviceId: out.DeviceId,
	}, nil
}
