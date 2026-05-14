// This file implements the plugin dependency-check management endpoint.
package plugin

import (
	"context"

	"lina-core/api/plugin/v1"
)

// DependencyCheck returns the server-side dependency check result for one plugin.
func (c *ControllerV1) DependencyCheck(ctx context.Context, req *v1.DependencyCheckReq) (res *v1.DependencyCheckRes, err error) {
	out, err := c.pluginSvc.CheckPluginDependencies(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return buildPluginDependencyCheckResult(out), nil
}
