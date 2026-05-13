// This file implements the tenant-device strategy binding list controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// ListTenantDeviceBindings returns paged tenant-device strategy bindings.
func (c *ControllerV1) ListTenantDeviceBindings(ctx context.Context, req *v1.ListTenantDeviceBindingsReq) (res *v1.ListTenantDeviceBindingsRes, err error) {
	out, err := c.mediaSvc.ListTenantDeviceBindings(ctx, mediasvc.ListBindingsInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Keyword:  req.Keyword,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.TenantDeviceBindingItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, &v1.TenantDeviceBindingItem{
			RowKey:       item.RowKey,
			TenantId:     item.TenantId,
			DeviceId:     item.DeviceId,
			StrategyId:   item.StrategyId,
			StrategyName: item.StrategyName,
		})
	}
	return &v1.ListTenantDeviceBindingsRes{List: items, Total: out.Total}, nil
}
