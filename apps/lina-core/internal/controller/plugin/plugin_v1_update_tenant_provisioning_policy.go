package plugin

import (
	"context"

	"lina-core/api/plugin/v1"
)

// UpdateTenantProvisioningPolicy updates a platform-owned new-tenant plugin policy.
func (c *ControllerV1) UpdateTenantProvisioningPolicy(ctx context.Context, req *v1.UpdateTenantProvisioningPolicyReq) (res *v1.UpdateTenantProvisioningPolicyRes, err error) {
	if err = c.pluginSvc.UpdateTenantProvisioningPolicy(ctx, req.Id, req.AutoEnableForNewTenants); err != nil {
		return nil, err
	}
	return &v1.UpdateTenantProvisioningPolicyRes{
		Id:                      req.Id,
		AutoEnableForNewTenants: req.AutoEnableForNewTenants,
	}, nil
}
