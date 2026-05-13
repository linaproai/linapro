// This file implements the tenant whitelist list controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// ListTenantWhites returns paged tenant whitelist entries.
func (c *ControllerV1) ListTenantWhites(ctx context.Context, req *v1.ListTenantWhitesReq) (res *v1.ListTenantWhitesRes, err error) {
	out, err := c.mediaSvc.ListTenantWhites(ctx, mediasvc.ListTenantWhitesInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Keyword:  req.Keyword,
		Enable:   req.Enable,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.TenantWhiteListItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, tenantWhiteOutputToItem(item))
	}
	return &v1.ListTenantWhitesRes{List: items, Total: out.Total}, nil
}

// tenantWhiteOutputToItem converts service tenant whitelist output to API item.
func tenantWhiteOutputToItem(out *mediasvc.TenantWhiteOutput) *v1.TenantWhiteListItem {
	if out == nil {
		return &v1.TenantWhiteListItem{}
	}
	return &v1.TenantWhiteListItem{
		TenantId:    out.TenantId,
		Ip:          out.Ip,
		Description: out.Description,
		Enable:      out.Enable,
		CreatorId:   out.CreatorId,
		CreateTime:  out.CreateTime,
		UpdaterId:   out.UpdaterId,
		UpdateTime:  out.UpdateTime,
	}
}
