// This file implements the device-node update controller endpoint.

package media

import (
	"context"
	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// UpdateDeviceNode updates one device-node mapping.
func (c *ControllerV1) UpdateDeviceNode(ctx context.Context, req *v1.UpdateDeviceNodeReq) (res *v1.UpdateDeviceNodeRes, err error) {
	out, err := c.mediaSvc.UpdateDeviceNode(ctx, req.OldDeviceId, mediasvc.DeviceNodeMutationInput{
		DeviceId: req.DeviceId,
		NodeNum:  req.NodeNum,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateDeviceNodeRes{DeviceId: out.DeviceId}, nil
}
