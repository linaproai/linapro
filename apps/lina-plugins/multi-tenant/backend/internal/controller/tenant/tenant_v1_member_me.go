// This file implements the tenant current-member endpoint.

package tenant

import (
	"context"

	v1 "lina-plugin-multi-tenant/backend/api/tenant/v1"
)

// MemberMe returns the current user's membership profile.
func (c *ControllerV1) MemberMe(ctx context.Context, _ *v1.MemberMeReq) (res *v1.MemberMeRes, err error) {
	item, err := c.membershipSvc.Current(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.MemberMeRes{MemberEntity: toAPIMember(item)}, nil
}
