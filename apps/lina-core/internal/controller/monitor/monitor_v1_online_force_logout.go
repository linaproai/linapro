package monitor

import (
	"context"

	v1 "lina-core/api/monitor/v1"
)

// OnlineForceLogout forces online user to logout
func (c *ControllerV1) OnlineForceLogout(ctx context.Context, req *v1.OnlineForceLogoutReq) (res *v1.OnlineForceLogoutRes, err error) {
	if err = c.authSvc.RevokeSession(ctx, req.TokenId); err != nil {
		return nil, err
	}
	return &v1.OnlineForceLogoutRes{}, nil
}
