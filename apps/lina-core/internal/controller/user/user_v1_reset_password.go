package user

import (
	"context"

	v1 "lina-core/api/user/v1"
)

// ResetPassword resets user password
func (c *ControllerV1) ResetPassword(ctx context.Context, req *v1.ResetPasswordReq) (res *v1.ResetPasswordRes, err error) {
	return nil, c.userSvc.ResetPassword(ctx, req.Id, req.Password)
}
