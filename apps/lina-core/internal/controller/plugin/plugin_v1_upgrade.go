// This file implements the plugin runtime-upgrade execution management API.

package plugin

import (
	"context"

	"lina-core/api/plugin/v1"
	pluginsvc "lina-core/internal/service/plugin"
)

// Upgrade executes a confirmed runtime upgrade for one pending plugin.
func (c *ControllerV1) Upgrade(
	ctx context.Context,
	req *v1.UpgradeReq,
) (res *v1.UpgradeRes, err error) {
	result, err := c.pluginSvc.ExecuteRuntimeUpgrade(ctx, req.Id, pluginsvc.RuntimeUpgradeOptions{
		Confirmed:     req.Confirmed,
		Authorization: buildAuthorizationInput(req.Authorization),
	})
	if err != nil {
		return nil, err
	}
	c.roleSvc.NotifyAccessTopologyChanged(ctx)
	return buildUpgradeResponse(result), nil
}

// buildUpgradeResponse maps service upgrade result output into the public DTO.
func buildUpgradeResponse(result *pluginsvc.RuntimeUpgradeResult) *v1.UpgradeRes {
	if result == nil {
		return nil
	}
	return &v1.UpgradeRes{
		PluginId:          result.PluginID,
		RuntimeState:      result.RuntimeState.String(),
		EffectiveVersion:  result.EffectiveVersion,
		DiscoveredVersion: result.DiscoveredVersion,
		FromVersion:       result.FromVersion,
		ToVersion:         result.ToVersion,
		Executed:          result.Executed,
	}
}
