package auth

import (
	"context"

	v1 "lina-core/api/auth/v1"
)

// Logout handles user logout.
func (c *ControllerV1) Logout(ctx context.Context, req *v1.LogoutReq) (res *v1.LogoutRes, err error) {
	// Record logout log and delete session
	if bizCtx := c.bizCtxSvc.Get(ctx); bizCtx != nil {
		c.authSvc.Logout(ctx, bizCtx.Username, bizCtx.TokenId)
	}
	return &v1.LogoutRes{}, nil
}
