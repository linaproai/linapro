// This file implements the tenant whitelist deletion controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
)

// DeleteTenantWhite deletes one tenant whitelist entry.
func (c *ControllerV1) DeleteTenantWhite(ctx context.Context, req *v1.DeleteTenantWhiteReq) (res *v1.DeleteTenantWhiteRes, err error) {
	out, err := c.mediaSvc.DeleteTenantWhite(ctx, req.TenantId, req.Ip)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteTenantWhiteRes{TenantId: out.TenantId, Ip: out.Ip}, nil
}
