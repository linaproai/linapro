// This file handles public password-reset confirmation.

package auth

import (
	"context"

	v1 "lina-core/api/auth/v1"
	authsvc "lina-core/internal/service/auth"
)

// ResetPassword handles password-reset confirmation with a one-time token.
func (c *ControllerV1) ResetPassword(ctx context.Context, req *v1.ResetPasswordReq) (res *v1.ResetPasswordRes, err error) {
	if err = c.authSvc.ConfirmPasswordReset(ctx, authsvc.PasswordResetConfirmInput{
		Token:    req.Token,
		Password: req.Password,
	}); err != nil {
		return nil, err
	}
	return &v1.ResetPasswordRes{Reset: true}, nil
}
