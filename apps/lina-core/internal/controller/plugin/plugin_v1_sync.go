package plugin

import (
	"context"

	"lina-core/api/plugin/v1"
)

// Sync scans source plugins and synchronizes plugin registry metadata.
func (c *ControllerV1) Sync(ctx context.Context, _ *v1.SyncReq) (res *v1.SyncRes, err error) {
	out, err := c.pluginSvc.SyncAndList(ctx)
	if err != nil {
		return nil, err
	}
	c.roleSvc.NotifyAccessTopologyChanged(ctx)

	return &v1.SyncRes{Total: out.Total}, nil
}
