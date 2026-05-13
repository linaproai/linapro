// This file implements the tenant whitelist detail controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
)

// GetTenantWhite returns one tenant whitelist entry.
func (c *ControllerV1) GetTenantWhite(ctx context.Context, req *v1.GetTenantWhiteReq) (res *v1.GetTenantWhiteRes, err error) {
	out, err := c.mediaSvc.GetTenantWhite(ctx, req.TenantId, req.Ip)
	if err != nil {
		return nil, err
	}
	return tenantWhiteOutputToItem(out), nil
}
