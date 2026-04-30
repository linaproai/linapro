// This file maps the role batch-delete API to the role service.

package role

import (
	"context"

	"lina-core/api/role/v1"
)

// RoleBatchDelete deletes multiple roles.
func (c *ControllerV1) RoleBatchDelete(
	ctx context.Context,
	req *v1.RoleBatchDeleteReq,
) (res *v1.RoleBatchDeleteRes, err error) {
	if err = c.roleSvc.BatchDelete(ctx, req.Ids); err != nil {
		return nil, err
	}
	return &v1.RoleBatchDeleteRes{}, nil
}
