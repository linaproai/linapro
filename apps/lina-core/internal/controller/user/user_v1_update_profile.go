package user

import (
	"context"

	v1 "lina-core/api/user/v1"
	usersvc "lina-core/internal/service/user"
)

// UpdateProfile updates user profile
func (c *ControllerV1) UpdateProfile(ctx context.Context, req *v1.UpdateProfileReq) (res *v1.UpdateProfileRes, err error) {
	return nil, c.userSvc.UpdateProfile(ctx, usersvc.UpdateProfileInput{
		Nickname: req.Nickname,
		Email:    req.Email,
		Phone:    req.Phone,
		Sex:      req.Sex,
		Password: req.Password,
	})
}
