// This file implements the device-node creation controller endpoint.

package media

import (
	"context"
	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// CreateDeviceNode creates one device-node mapping.
func (c *ControllerV1) CreateDeviceNode(ctx context.Context, req *v1.CreateDeviceNodeReq) (res *v1.CreateDeviceNodeRes, err error) {
	out, err := c.mediaSvc.CreateDeviceNode(ctx, mediasvc.DeviceNodeMutationInput{
		DeviceId: req.DeviceId,
		NodeNum:  req.NodeNum,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateDeviceNodeRes{DeviceId: out.DeviceId}, nil
}
