// This file implements the scheduled job handler detail endpoint.

package jobhandler

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/api/jobhandler/v1"
)

// Detail handles scheduled job handler detail lookup requests.
func (c *ControllerV1) Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error) {
	item, ok := c.registry.Lookup(req.Ref)
	if !ok {
		return nil, gerror.New("任务处理器不存在")
	}
	return &v1.DetailRes{
		Ref:          item.Ref,
		DisplayName:  c.localizeHandlerName(ctx, item.Ref, item.DisplayName),
		Description:  c.localizeHandlerDescription(ctx, item.Ref, item.Description),
		Source:       string(item.Source),
		PluginId:     item.PluginID,
		ParamsSchema: item.ParamsSchema,
	}, nil
}
