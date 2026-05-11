// This file implements the tenant member-update endpoint.

package tenant

import (
	"context"

	v1 "lina-plugin-multi-tenant/backend/api/tenant/v1"
	"lina-plugin-multi-tenant/backend/internal/service/membership"
)

// MemberUpdate updates current tenant membership flags.
func (c *ControllerV1) MemberUpdate(ctx context.Context, req *v1.MemberUpdateReq) (res *v1.MemberUpdateRes, err error) {
	err = c.membershipSvc.Update(ctx, membership.UpdateInput{
		Id:     req.Id,
		Status: req.Status,
	})
	if err != nil {
		return nil, err
	}
	return &v1.MemberUpdateRes{}, nil
}
