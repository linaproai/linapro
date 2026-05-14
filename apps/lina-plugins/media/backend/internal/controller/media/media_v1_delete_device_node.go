// This file implements the device-node deletion controller endpoint.

package media

import (
	"context"
	"lina-plugin-media/backend/api/media/v1"
)

// DeleteDeviceNode deletes one device-node mapping.
func (c *ControllerV1) DeleteDeviceNode(ctx context.Context, req *v1.DeleteDeviceNodeReq) (res *v1.DeleteDeviceNodeRes, err error) {
	out, err := c.mediaSvc.DeleteDeviceNode(ctx, req.DeviceId)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteDeviceNodeRes{DeviceId: out.DeviceId}, nil
}
