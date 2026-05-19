// This file implements plugin enablement and keeps the service integer status
// update aligned with the public enabled flag contract.

package plugin

import (
	"context"

	"lina-core/api/plugin/v1"
	"lina-core/pkg/statusflag"
)

// Enable updates plugin status to enabled.
func (c *ControllerV1) Enable(ctx context.Context, req *v1.EnableReq) (res *v1.EnableRes, err error) {
	if err = c.pluginSvc.UpdateStatus(ctx, req.Id, statusflag.EnabledValue.Int(), buildAuthorizationInput(req.Authorization)); err != nil {
		return nil, err
	}
	c.roleSvc.NotifyAccessTopologyChanged(ctx)

	return &v1.EnableRes{Id: req.Id, Enabled: statusflag.EnabledValue}, nil
}
