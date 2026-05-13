// This file implements the tenant whitelist update controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// UpdateTenantWhite updates one tenant whitelist entry.
func (c *ControllerV1) UpdateTenantWhite(ctx context.Context, req *v1.UpdateTenantWhiteReq) (res *v1.UpdateTenantWhiteRes, err error) {
	out, err := c.mediaSvc.UpdateTenantWhite(ctx, req.OldTenantId, req.OldIp, mediasvc.TenantWhiteMutationInput{
		TenantId:    req.TenantId,
		Ip:          req.Ip,
		Description: req.Description,
		Enable:      req.Enable,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateTenantWhiteRes{TenantId: out.TenantId, Ip: out.Ip}, nil
}
