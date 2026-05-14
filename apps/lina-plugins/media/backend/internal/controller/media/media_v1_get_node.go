// This file implements the media node detail controller endpoint.

package media

import (
	"context"
	"lina-plugin-media/backend/api/media/v1"
)

// GetNode returns one media node by node number.
func (c *ControllerV1) GetNode(ctx context.Context, req *v1.GetNodeReq) (res *v1.GetNodeRes, err error) {
	out, err := c.mediaSvc.GetNode(ctx, req.NodeNum)
	if err != nil {
		return nil, err
	}
	return nodeOutputToItem(out), nil
}
