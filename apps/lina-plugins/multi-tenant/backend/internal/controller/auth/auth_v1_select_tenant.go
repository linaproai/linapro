// This file implements the tenant-selection validation endpoint.

package auth

import (
	"context"

	pluginauth "lina-core/pkg/pluginservice/contract"
	v1 "lina-plugin-multi-tenant/backend/api/auth/v1"
)

// SelectTenant validates tenant selection and requests a host-signed token.
func (c *ControllerV1) SelectTenant(ctx context.Context, req *v1.SelectTenantReq) (res *v1.SelectTenantRes, err error) {
	out, err := c.authSvc.SelectTenant(ctx, pluginauth.SelectTenantInput{
		PreToken: req.PreToken,
		TenantID: int(req.TenantId),
	})
	if err != nil {
		return nil, err
	}
	return &v1.SelectTenantRes{AccessToken: out.AccessToken, RefreshToken: out.RefreshToken}, nil
}
