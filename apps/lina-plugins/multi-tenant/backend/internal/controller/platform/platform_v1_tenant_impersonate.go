// This file implements the platform tenant-impersonation start endpoint.

package platform

import (
	"context"

	v1 "lina-plugin-multi-tenant/backend/api/platform/v1"
	"lina-plugin-multi-tenant/backend/internal/service/impersonate"
)

// TenantImpersonate starts platform impersonation.
func (c *ControllerV1) TenantImpersonate(ctx context.Context, req *v1.TenantImpersonateReq) (res *v1.TenantImpersonateRes, err error) {
	out, err := c.impersonateSvc.Start(ctx, impersonate.StartInput{
		TenantID: req.Id,
		Reason:   req.Reason,
	})
	if err != nil && out == nil {
		return nil, err
	}
	if out == nil {
		out = &impersonate.StartOutput{TenantID: req.Id}
	}
	return &v1.TenantImpersonateRes{
		Token:          out.Token,
		TenantId:       out.TenantID,
		ActingUserId:   out.ActingUserID,
		IsImpersonated: out.IsImpersonated,
	}, err
}
