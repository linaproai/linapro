// This file implements plugin installation and maps lifecycle results into
// typed public API status flags.

package plugin

import (
	"context"

	"lina-core/api/plugin/v1"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/statusflag"
)

// Install executes plugin install lifecycle.
func (c *ControllerV1) Install(ctx context.Context, req *v1.InstallReq) (res *v1.InstallRes, err error) {
	options := pluginsvc.InstallOptions{
		Authorization:   buildAuthorizationInput(req.Authorization),
		InstallMode:     string(req.InstallMode),
		InstallMockData: req.InstallMockData,
	}
	dependencyCheck, err := c.pluginSvc.Install(ctx, req.Id, options)
	if err != nil {
		return nil, err
	}
	c.roleSvc.NotifyAccessTopologyChanged(ctx)
	return &v1.InstallRes{
		Id:              req.Id,
		Installed:       statusflag.Installed,
		Enabled:         statusflag.Disabled,
		DependencyCheck: buildPluginDependencyCheckResult(dependencyCheck),
	}, nil
}
