// This file implements the media node list controller endpoint.

package media

import (
	"context"
	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// ListNodes returns paged media nodes.
func (c *ControllerV1) ListNodes(ctx context.Context, req *v1.ListNodesReq) (res *v1.ListNodesRes, err error) {
	out, err := c.mediaSvc.ListNodes(ctx, mediasvc.ListNodesInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Keyword:  req.Keyword,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.NodeListItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, nodeOutputToItem(item))
	}
	return &v1.ListNodesRes{List: items, Total: out.Total}, nil
}
