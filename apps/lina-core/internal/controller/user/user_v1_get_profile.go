package user

import (
	"context"

	v1 "lina-core/api/user/v1"
)

// GetProfile returns user profile
func (c *ControllerV1) GetProfile(ctx context.Context, req *v1.GetProfileReq) (res *v1.GetProfileRes, err error) {
	user, err := c.userSvc.GetProfile(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetProfileRes{SysUser: user}, nil
}
