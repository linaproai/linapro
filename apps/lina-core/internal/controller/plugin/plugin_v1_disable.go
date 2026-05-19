// This file implements plugin disablement and projects the resulting disabled
// state through the public enabled flag contract.

package plugin

import (
	"context"

	"lina-core/api/plugin/v1"
	"lina-core/pkg/statusflag"
)

// Disable updates plugin status to disabled.
func (c *ControllerV1) Disable(ctx context.Context, req *v1.DisableReq) (res *v1.DisableRes, err error) {
	if err = c.pluginSvc.Disable(ctx, req.Id); err != nil {
		return nil, err
	}
	c.roleSvc.NotifyAccessTopologyChanged(ctx)

	return &v1.DisableRes{Id: req.Id, Enabled: statusflag.Disabled}, nil
}
