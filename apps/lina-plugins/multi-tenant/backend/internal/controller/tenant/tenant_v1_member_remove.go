// This file implements the tenant member-remove endpoint.

package tenant

import (
	"context"

	v1 "lina-plugin-multi-tenant/backend/api/tenant/v1"
)

// MemberRemove removes a current tenant member.
func (c *ControllerV1) MemberRemove(ctx context.Context, req *v1.MemberRemoveReq) (res *v1.MemberRemoveRes, err error) {
	if err = c.membershipSvc.Remove(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.MemberRemoveRes{}, nil
}
