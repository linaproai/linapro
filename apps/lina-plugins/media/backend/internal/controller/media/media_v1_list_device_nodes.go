// This file implements the device-node list controller endpoint.

package media

import (
	"context"
	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// ListDeviceNodes returns paged device-node mappings.
func (c *ControllerV1) ListDeviceNodes(ctx context.Context, req *v1.ListDeviceNodesReq) (res *v1.ListDeviceNodesRes, err error) {
	out, err := c.mediaSvc.ListDeviceNodes(ctx, mediasvc.ListDeviceNodesInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Keyword:  req.Keyword,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.DeviceNodeListItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, deviceNodeOutputToItem(item))
	}
	return &v1.ListDeviceNodesRes{List: items, Total: out.Total}, nil
}
