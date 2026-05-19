// This file maps current-user message list rows into public DTOs and applies
// shared read-state flag contracts.

package usermsg

import (
	"context"

	v1 "lina-core/api/usermsg/v1"
	usermsgsvc "lina-core/internal/service/usermsg"
	"lina-core/pkg/apitime"
	"lina-core/pkg/statusflag"
)

// List queries user message list
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.usermsgSvc.List(ctx, usermsgsvc.ListInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	items := make([]*v1.MessageItem, 0, len(out.List))
	for _, item := range out.List {
		if item == nil {
			continue
		}
		items = append(items, &v1.MessageItem{
			Id:           item.Id,
			UserId:       item.UserId,
			Title:        item.Title,
			CategoryCode: item.CategoryCode,
			TypeLabel:    item.TypeLabel,
			TypeColor:    item.TypeColor,
			SourceType:   v1.SourceType(item.SourceType),
			SourceId:     item.SourceId,
			IsRead:       statusflag.ReadState(item.IsRead),
			ReadAt:       apitime.Milli(item.ReadAt),
			CreatedAt:    apitime.Milli(item.CreatedAt),
		})
	}
	return &v1.ListRes{List: items, Total: out.Total}, nil
}
