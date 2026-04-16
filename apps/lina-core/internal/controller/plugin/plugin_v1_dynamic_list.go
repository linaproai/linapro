// This file exposes the public dynamic plugin runtime-state endpoint on the
// shared plugin controller.

package plugin

import (
	"context"

	"lina-core/api/plugin/v1"
)

// DynamicList returns public dynamic-plugin states for shell slot rendering.
func (c *ControllerV1) DynamicList(ctx context.Context, req *v1.DynamicListReq) (res *v1.DynamicListRes, err error) {
	out, err := c.pluginSvc.ListRuntimeStates(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]*v1.PluginDynamicItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, &v1.PluginDynamicItem{
			Id:         item.Id,
			Installed:  item.Installed,
			Enabled:    item.Enabled,
			Version:    item.Version,
			Generation: item.Generation,
			StatusKey:  item.StatusKey,
		})
	}
	return &v1.DynamicListRes{List: items}, nil
}
