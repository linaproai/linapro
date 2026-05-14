// This file implements the tenant stream config list controller endpoint.

package media

import (
	"context"
	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// ListTenantStreamConfigs returns paged tenant stream configs.
func (c *ControllerV1) ListTenantStreamConfigs(ctx context.Context, req *v1.ListTenantStreamConfigsReq) (res *v1.ListTenantStreamConfigsRes, err error) {
	out, err := c.mediaSvc.ListTenantStreamConfigs(ctx, mediasvc.ListTenantStreamConfigsInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Keyword:  req.Keyword,
		Enable:   req.Enable,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.TenantStreamConfigListItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, tenantStreamConfigOutputToItem(item))
	}
	return &v1.ListTenantStreamConfigsRes{List: items, Total: out.Total}, nil
}
