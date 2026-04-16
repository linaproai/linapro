package user

import (
	"context"

	v1 "lina-core/api/user/v1"
)

// Delete deletes a user
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	return nil, c.userSvc.Delete(ctx, req.Id)
}
