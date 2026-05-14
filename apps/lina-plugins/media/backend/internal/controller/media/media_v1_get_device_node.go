// This file implements the device-node detail controller endpoint.

package media

import (
	"context"
	"lina-plugin-media/backend/api/media/v1"
)

// GetDeviceNode returns one device-node mapping by device ID.
func (c *ControllerV1) GetDeviceNode(ctx context.Context, req *v1.GetDeviceNodeReq) (res *v1.GetDeviceNodeRes, err error) {
	out, err := c.mediaSvc.GetDeviceNode(ctx, req.DeviceId)
	if err != nil {
		return nil, err
	}
	return deviceNodeOutputToItem(out), nil
}
