// This file implements platform tenant membership endpoints.

package platform

import (
	"context"

	v1 "lina-plugin-multi-tenant/backend/api/platform/v1"
	"lina-plugin-multi-tenant/backend/internal/service/membership"
)

// MemberList queries memberships across tenants.
func (c *ControllerV1) MemberList(ctx context.Context, req *v1.MemberListReq) (res *v1.MemberListRes, err error) {
	out, err := c.membershipSvc.List(ctx, membership.ListInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		TenantID: req.TenantId,
		UserID:   req.UserId,
		Status:   req.Status,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.MemberEntity, 0, len(out.List))
	for _, item := range out.List {
		list = append(list, toAPIMember(item))
	}
	return &v1.MemberListRes{List: list, Total: out.Total}, nil
}
