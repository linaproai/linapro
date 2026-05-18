package plugin

import (
	"context"

	"lina-core/api/plugin/v1"
	pluginsvc "lina-core/internal/service/plugin"
)

// Uninstall executes plugin uninstall lifecycle.
func (c *ControllerV1) Uninstall(ctx context.Context, req *v1.UninstallReq) (res *v1.UninstallRes, err error) {
	options := pluginsvc.UninstallOptions{
		PurgeStorageData: resolvePurgeStorageData(req.PurgeStorageData),
		Force:            req.Force,
	}
	if err = c.pluginSvc.Uninstall(ctx, req.Id, options); err != nil {
		return nil, err
	}
	c.roleSvc.NotifyAccessTopologyChanged(ctx)
	return &v1.UninstallRes{Id: req.Id, Installed: 0, Enabled: 0}, nil
}

// resolvePurgeStorageData converts the optional integer request flag into the
// effective uninstall storage-purge behavior.
func resolvePurgeStorageData(value *int) bool {
	if value == nil {
		return true
	}
	return *value == 1
}
