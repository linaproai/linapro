// This file implements the media node deletion controller endpoint.

package media

import (
	"context"
	"lina-plugin-media/backend/api/media/v1"
)

// DeleteNode deletes one unreferenced media node.
func (c *ControllerV1) DeleteNode(ctx context.Context, req *v1.DeleteNodeReq) (res *v1.DeleteNodeRes, err error) {
	out, err := c.mediaSvc.DeleteNode(ctx, req.NodeNum)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteNodeRes{NodeNum: out.NodeNum}, nil
}
