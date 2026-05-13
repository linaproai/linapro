// This file implements the tenant strategy binding deletion controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
)

// DeleteTenantBinding deletes one tenant strategy binding.
func (c *ControllerV1) DeleteTenantBinding(ctx context.Context, req *v1.DeleteTenantBindingReq) (res *v1.DeleteTenantBindingRes, err error) {
	out, err := c.mediaSvc.DeleteTenantBinding(ctx, req.TenantId)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteTenantBindingRes{TenantId: out.TenantId}, nil
}
