// This file implements the media node update controller endpoint.

package media

import (
	"context"
	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// UpdateNode updates one media node.
func (c *ControllerV1) UpdateNode(ctx context.Context, req *v1.UpdateNodeReq) (res *v1.UpdateNodeRes, err error) {
	out, err := c.mediaSvc.UpdateNode(ctx, req.OldNodeNum, mediasvc.NodeMutationInput{
		NodeNum:  req.NodeNum,
		Name:     req.Name,
		QnUrl:    req.QnUrl,
		BasicUrl: req.BasicUrl,
		DnUrl:    req.DnUrl,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNodeRes{NodeNum: out.NodeNum}, nil
}
