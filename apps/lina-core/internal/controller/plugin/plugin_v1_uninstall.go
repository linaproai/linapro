package plugin

import (
	"context"

	"lina-core/api/plugin/v1"
)

// Uninstall executes plugin uninstall lifecycle.
func (c *ControllerV1) Uninstall(ctx context.Context, req *v1.UninstallReq) (res *v1.UninstallRes, err error) {
	if err = c.pluginSvc.Uninstall(ctx, req.Id); err != nil {
		return nil, err
	}
	c.roleSvc.NotifyAccessTopologyChanged(ctx)
	return &v1.UninstallRes{Id: req.Id, Installed: 0, Enabled: 0}, nil
}
