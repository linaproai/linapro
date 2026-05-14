// This file implements the tenant stream config update controller endpoint.

package media

import (
	"context"
	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// UpdateTenantStreamConfig updates one tenant stream config.
func (c *ControllerV1) UpdateTenantStreamConfig(ctx context.Context, req *v1.UpdateTenantStreamConfigReq) (res *v1.UpdateTenantStreamConfigRes, err error) {
	out, err := c.mediaSvc.UpdateTenantStreamConfig(ctx, req.OldTenantId, mediasvc.TenantStreamConfigMutationInput{
		TenantId:      req.TenantId,
		MaxConcurrent: req.MaxConcurrent,
		NodeNum:       req.NodeNum,
		Enable:        req.Enable,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateTenantStreamConfigRes{TenantId: out.TenantId}, nil
}
