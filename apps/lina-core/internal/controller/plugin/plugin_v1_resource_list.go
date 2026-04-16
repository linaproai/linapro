package plugin

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/api/plugin/v1"
	pluginsvc "lina-core/internal/service/plugin"
)

// ResourceList queries plugin-owned backend resources through the generic resource contract.
func (c *ControllerV1) ResourceList(ctx context.Context, req *v1.ResourceListReq) (res *v1.ResourceListRes, err error) {
	filters := make(map[string]string)
	if r := g.RequestFromCtx(ctx); r != nil {
		for key, value := range r.GetQueryMap() {
			if key == "id" || key == "resource" || key == "pageNum" || key == "pageSize" {
				continue
			}
			filters[key] = gconv.String(value)
		}
	}

	out, err := c.pluginSvc.ListResourceRecords(ctx, pluginsvc.ResourceListInput{
		PluginID:   req.Id,
		ResourceID: req.Resource,
		Filters:    filters,
		PageNum:    req.PageNum,
		PageSize:   req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ResourceListRes{List: out.List, Total: out.Total}, nil
}
