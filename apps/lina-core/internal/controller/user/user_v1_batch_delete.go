// This file maps the user batch-delete API to the user service.

package user

import (
	"context"

	v1 "lina-core/api/user/v1"
)

// BatchDelete deletes multiple users.
func (c *ControllerV1) BatchDelete(ctx context.Context, req *v1.BatchDeleteReq) (res *v1.BatchDeleteRes, err error) {
	if err = c.userSvc.BatchDelete(ctx, req.Ids); err != nil {
		return nil, err
	}
	return &v1.BatchDeleteRes{}, nil
}
