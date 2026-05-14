// This file implements the tenant stream config detail controller endpoint.

package media

import (
	"context"
	"lina-plugin-media/backend/api/media/v1"
)

// GetTenantStreamConfig returns one tenant stream config by tenant ID.
func (c *ControllerV1) GetTenantStreamConfig(ctx context.Context, req *v1.GetTenantStreamConfigReq) (res *v1.GetTenantStreamConfigRes, err error) {
	out, err := c.mediaSvc.GetTenantStreamConfig(ctx, req.TenantId)
	if err != nil {
		return nil, err
	}
	return tenantStreamConfigOutputToItem(out), nil
}
