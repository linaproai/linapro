// This file implements the device strategy binding list controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// ListDeviceBindings returns paged device strategy bindings.
func (c *ControllerV1) ListDeviceBindings(ctx context.Context, req *v1.ListDeviceBindingsReq) (res *v1.ListDeviceBindingsRes, err error) {
	out, err := c.mediaSvc.ListDeviceBindings(ctx, mediasvc.ListBindingsInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Keyword:  req.Keyword,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.DeviceBindingItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, &v1.DeviceBindingItem{
			RowKey:       item.RowKey,
			DeviceId:     item.DeviceId,
			StrategyId:   item.StrategyId,
			StrategyName: item.StrategyName,
		})
	}
	return &v1.ListDeviceBindingsRes{List: items, Total: out.Total}, nil
}
