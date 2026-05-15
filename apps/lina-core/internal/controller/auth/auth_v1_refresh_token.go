// This file implements the token refresh endpoint for the auth controller.

package auth

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/api/auth/v1"
)

// RefreshToken validates the current JWT and issues a new token with a refreshed expiry.
func (c *ControllerV1) RefreshToken(ctx context.Context, req *v1.RefreshTokenReq) (res *v1.RefreshTokenRes, err error) {
	r := g.RequestFromCtx(ctx)
	tokenHeader := r.GetHeader("Authorization")
	tokenString := strings.TrimPrefix(tokenHeader, "Bearer ")
	newToken, err := c.authSvc.RefreshToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}
	return &v1.RefreshTokenRes{AccessToken: newToken}, nil
}
