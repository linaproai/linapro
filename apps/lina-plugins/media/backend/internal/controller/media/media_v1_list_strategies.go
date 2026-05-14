// This file implements the media strategy list controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// ListStrategies returns paged media strategies.
func (c *ControllerV1) ListStrategies(ctx context.Context, req *v1.ListStrategiesReq) (res *v1.ListStrategiesRes, err error) {
	out, err := c.mediaSvc.ListStrategies(ctx, mediasvc.ListStrategiesInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Keyword:  req.Keyword,
		Enable:   req.Enable,
		Global:   req.Global,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.StrategyListItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, &v1.StrategyListItem{
			Id:         item.Id,
			Name:       item.Name,
			Strategy:   item.Strategy,
			Global:     item.Global,
			Enable:     item.Enable,
			CreatorId:  item.CreatorId,
			UpdaterId:  item.UpdaterId,
			CreateTime: item.CreateTime,
			UpdateTime: item.UpdateTime,
		})
	}
	return &v1.ListStrategiesRes{List: items, Total: out.Total}, nil
}
