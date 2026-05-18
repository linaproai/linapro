package user

import (
	"context"

	v1 "lina-core/api/user/v1"
)

// UpdateAvatar updates user avatar
func (c *ControllerV1) UpdateAvatar(ctx context.Context, req *v1.UpdateAvatarReq) (res *v1.UpdateAvatarRes, err error) {
	err = c.userSvc.UpdateAvatar(ctx, req.Avatar)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateAvatarRes{}, nil
}
