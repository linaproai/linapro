package auth

import (
	"context"

	v1 "lina-core/api/auth/v1"
	authsvc "lina-core/internal/service/auth"
)

// Logout handles user logout.
func (c *ControllerV1) Logout(ctx context.Context, req *v1.LogoutReq) (res *v1.LogoutRes, err error) {
	// Record logout log and delete session
	if bizCtx := c.bizCtxSvc.Get(ctx); bizCtx != nil {
		if err = c.authSvc.Logout(ctx, authsvc.LogoutInput{
			Username:   bizCtx.Username,
			TenantID:   bizCtx.TenantId,
			TokenID:    bizCtx.TokenId,
			ClientType: authsvc.ClientType(bizCtx.ClientType),
		}); err != nil {
			return nil, err
		}
	}
	return &v1.LogoutRes{}, nil
}
