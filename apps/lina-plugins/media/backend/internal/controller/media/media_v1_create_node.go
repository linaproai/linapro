// This file implements the media node creation controller endpoint.

package media

import (
	"context"
	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// CreateNode creates one media node.
func (c *ControllerV1) CreateNode(ctx context.Context, req *v1.CreateNodeReq) (res *v1.CreateNodeRes, err error) {
	out, err := c.mediaSvc.CreateNode(ctx, mediasvc.NodeMutationInput{
		NodeNum:  req.NodeNum,
		Name:     req.Name,
		QnUrl:    req.QnUrl,
		BasicUrl: req.BasicUrl,
		DnUrl:    req.DnUrl,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateNodeRes{NodeNum: out.NodeNum}, nil
}
