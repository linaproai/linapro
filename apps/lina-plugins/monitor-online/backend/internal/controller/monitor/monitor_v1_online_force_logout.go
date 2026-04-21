package monitor

import (
	"context"

	v1 "lina-plugin-monitor-online/backend/api/monitor/v1"
)

// OnlineForceLogout invalidates one active online-user session.
func (c *ControllerV1) OnlineForceLogout(ctx context.Context, req *v1.OnlineForceLogoutReq) (res *v1.OnlineForceLogoutRes, err error) {
	if err = c.monitorSvc.ForceLogout(ctx, req.TokenId); err != nil {
		return nil, err
	}
	return &v1.OnlineForceLogoutRes{}, nil
}
