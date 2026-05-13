// This file implements the tenant-switch validation endpoint.

package auth

import (
	"context"

	pluginauth "lina-core/pkg/pluginservice/contract"
	v1 "lina-plugin-multi-tenant/backend/api/auth/v1"
)

// SwitchTenant validates target tenant visibility and requests a host-signed token.
func (c *ControllerV1) SwitchTenant(ctx context.Context, req *v1.SwitchTenantReq) (res *v1.SwitchTenantRes, err error) {
	tokenString, _ := pluginauth.BearerTokenFromContext(ctx)
	out, err := c.authSvc.SwitchTenant(ctx, pluginauth.SwitchTenantInput{
		BearerToken: tokenString,
		TenantID:    int(req.TenantId),
	})
	if err != nil {
		return nil, err
	}
	return &v1.SwitchTenantRes{AccessToken: out.AccessToken, RefreshToken: out.RefreshToken}, nil
}
