// This file implements the device strategy binding deletion controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
)

// DeleteDeviceBinding deletes one device strategy binding.
func (c *ControllerV1) DeleteDeviceBinding(ctx context.Context, req *v1.DeleteDeviceBindingReq) (res *v1.DeleteDeviceBindingRes, err error) {
	out, err := c.mediaSvc.DeleteDeviceBinding(ctx, req.DeviceId)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteDeviceBindingRes{DeviceId: out.DeviceId}, nil
}
