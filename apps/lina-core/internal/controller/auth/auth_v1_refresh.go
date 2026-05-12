// This file implements the refresh-token endpoint.

package auth

import (
	"context"

	v1 "lina-core/api/auth/v1"
	authsvc "lina-core/internal/service/auth"
)

// Refresh exchanges a valid refresh token for a fresh access token.
func (c *ControllerV1) Refresh(ctx context.Context, req *v1.RefreshReq) (res *v1.RefreshRes, err error) {
	out, err := c.authSvc.Refresh(ctx, authsvc.RefreshInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return nil, err
	}
	return &v1.RefreshRes{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	}, nil
}
