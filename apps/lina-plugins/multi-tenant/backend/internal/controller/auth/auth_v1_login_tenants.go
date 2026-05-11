// This file implements the login-tenant candidate endpoint.

package auth

import (
	"context"

	v1 "lina-plugin-multi-tenant/backend/api/auth/v1"
)

// LoginTenants returns tenant candidates for one user.
func (c *ControllerV1) LoginTenants(ctx context.Context, req *v1.LoginTenantsReq) (res *v1.LoginTenantsRes, err error) {
	tenants, err := c.membershipSvc.ListUserTenants(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.LoginTenantEntity, 0, len(tenants))
	for _, item := range tenants {
		list = append(list, toAPILoginTenant(item))
	}
	return &v1.LoginTenantsRes{List: list}, nil
}
