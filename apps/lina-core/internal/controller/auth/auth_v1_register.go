// This file handles public self-registration.

package auth

import (
	"context"

	v1 "lina-core/api/auth/v1"
	authsvc "lina-core/internal/service/auth"
)

// Register handles public account registration.
func (c *ControllerV1) Register(ctx context.Context, req *v1.RegisterReq) (res *v1.RegisterRes, err error) {
	out, err := c.authSvc.Register(ctx, authsvc.RegisterInput{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		Nickname: req.Nickname,
	})
	if err != nil {
		return nil, err
	}
	return &v1.RegisterRes{UserId: out.UserID}, nil
}
