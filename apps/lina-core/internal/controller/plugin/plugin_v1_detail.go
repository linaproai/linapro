// This file implements plugin detail projection for the management API.

package plugin

import (
	"context"
	"strings"

	v1 "lina-core/api/plugin/v1"
	pluginsvc "lina-core/internal/service/plugin"
)

// Detail returns one plugin management detail projection.
func (c *ControllerV1) Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error) {
	item, err := c.pluginSvc.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	var (
		tableComments = c.pluginSvc.ResolveDataTableComments(
			ctx,
			collectPluginDataAuthorizationTables([]*pluginsvc.PluginItem{item}),
		)
		autoEnableManagedSet = buildAutoEnableManagedSet(c.configSvc.GetPluginAutoEnable(ctx))
		response             = c.buildPluginItemResponse(
			item,
			tableComments,
			autoEnableManagedSet[strings.TrimSpace(item.Id)],
		)
	)
	return &v1.DetailRes{PluginItem: *response}, nil
}
