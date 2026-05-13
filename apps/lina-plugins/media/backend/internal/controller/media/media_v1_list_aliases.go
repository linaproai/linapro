// This file implements the stream alias list controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// ListAliases returns paged stream aliases.
func (c *ControllerV1) ListAliases(ctx context.Context, req *v1.ListAliasesReq) (res *v1.ListAliasesRes, err error) {
	out, err := c.mediaSvc.ListAliases(ctx, mediasvc.ListAliasesInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Keyword:  req.Keyword,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.AliasListItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, &v1.AliasListItem{
			Id:         item.Id,
			Alias:      item.Alias,
			AutoRemove: item.AutoRemove,
			StreamPath: item.StreamPath,
			CreateTime:  item.CreateTime,
		})
	}
	return &v1.ListAliasesRes{List: items, Total: out.Total}, nil
}
