// This file implements the tenant strategy binding list controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// ListTenantBindings returns paged tenant strategy bindings.
func (c *ControllerV1) ListTenantBindings(ctx context.Context, req *v1.ListTenantBindingsReq) (res *v1.ListTenantBindingsRes, err error) {
	out, err := c.mediaSvc.ListTenantBindings(ctx, mediasvc.ListBindingsInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Keyword:  req.Keyword,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.TenantBindingItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, &v1.TenantBindingItem{
			RowKey:       item.RowKey,
			TenantId:     item.TenantId,
			StrategyId:   item.StrategyId,
			StrategyName: item.StrategyName,
		})
	}
	return &v1.ListTenantBindingsRes{List: items, Total: out.Total}, nil
}
