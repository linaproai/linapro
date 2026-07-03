// This file implements plugin disablement and projects the resulting disabled
// state through the public enabled flag contract.

package plugin

import (
	"context"

	v1 "lina-core/api/plugin/v1"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/statusflag"
)

// Disable updates plugin status to disabled.
func (c *ControllerV1) Disable(ctx context.Context, req *v1.DisableReq) (res *v1.DisableRes, err error) {
	if err = c.pluginSvc.UpdateStatus(ctx, req.Id, pluginsvc.UpdateStatusOptions{
		Status: statusflag.Disabled.Int(),
	}); err != nil {
		return nil, err
	}
	c.roleSvc.NotifyAccessTopologyChanged(ctx)

	return &v1.DisableRes{Id: req.Id, Enabled: statusflag.Disabled}, nil
}
