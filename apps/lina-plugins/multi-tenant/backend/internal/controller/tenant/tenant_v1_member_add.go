// This file implements the tenant member-add endpoint.

package tenant

import (
	"context"

	v1 "lina-plugin-multi-tenant/backend/api/tenant/v1"
	"lina-plugin-multi-tenant/backend/internal/service/membership"
)

// MemberAdd adds a member to the current tenant.
func (c *ControllerV1) MemberAdd(ctx context.Context, req *v1.MemberAddReq) (res *v1.MemberAddRes, err error) {
	id, err := c.membershipSvc.Add(ctx, membership.AddInput{
		TenantID: req.TenantId,
		UserID:   req.UserId,
	})
	if err != nil {
		return nil, err
	}
	return &v1.MemberAddRes{Id: id}, nil
}
