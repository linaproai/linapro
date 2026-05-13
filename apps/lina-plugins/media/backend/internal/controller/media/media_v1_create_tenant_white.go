// This file implements the tenant whitelist creation controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// CreateTenantWhite creates one tenant whitelist entry.
func (c *ControllerV1) CreateTenantWhite(ctx context.Context, req *v1.CreateTenantWhiteReq) (res *v1.CreateTenantWhiteRes, err error) {
	out, err := c.mediaSvc.CreateTenantWhite(ctx, mediasvc.TenantWhiteMutationInput{
		TenantId:    req.TenantId,
		Ip:          req.Ip,
		Description: req.Description,
		Enable:      req.Enable,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateTenantWhiteRes{TenantId: out.TenantId, Ip: out.Ip}, nil
}
