// This file implements the tenant stream config deletion controller endpoint.

package media

import (
	"context"
	"lina-plugin-media/backend/api/media/v1"
)

// DeleteTenantStreamConfig deletes one tenant stream config.
func (c *ControllerV1) DeleteTenantStreamConfig(ctx context.Context, req *v1.DeleteTenantStreamConfigReq) (res *v1.DeleteTenantStreamConfigRes, err error) {
	out, err := c.mediaSvc.DeleteTenantStreamConfig(ctx, req.TenantId)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteTenantStreamConfigRes{TenantId: out.TenantId}, nil
}
