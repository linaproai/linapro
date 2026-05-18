// This file implements the scheduled job handler detail endpoint.

package jobhandler

import (
	"context"

	"lina-core/api/jobhandler/v1"
	jobhandlersvc "lina-core/internal/service/jobhandler"
	"lina-core/pkg/bizerr"
)

// Detail handles scheduled job handler detail lookup requests.
func (c *ControllerV1) Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error) {
	item, ok := c.registry.Lookup(req.Ref)
	if !ok {
		return nil, bizerr.NewCode(jobhandlersvc.CodeJobHandlerNotFound)
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
