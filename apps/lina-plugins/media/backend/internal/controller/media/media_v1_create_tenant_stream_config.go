// This file implements the tenant stream config creation controller endpoint.

package media

import (
	"context"
	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// CreateTenantStreamConfig creates one tenant stream config.
func (c *ControllerV1) CreateTenantStreamConfig(ctx context.Context, req *v1.CreateTenantStreamConfigReq) (res *v1.CreateTenantStreamConfigRes, err error) {
	out, err := c.mediaSvc.CreateTenantStreamConfig(ctx, mediasvc.TenantStreamConfigMutationInput{
		TenantId:      req.TenantId,
		MaxConcurrent: req.MaxConcurrent,
		NodeNum:       req.NodeNum,
		Enable:        req.Enable,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateTenantStreamConfigRes{TenantId: out.TenantId}, nil
}
