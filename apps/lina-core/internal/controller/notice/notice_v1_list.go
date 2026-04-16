package notice

import (
	"context"

	v1 "lina-core/api/notice/v1"
	noticesvc "lina-core/internal/service/notice"
)

// List queries notice list
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.noticeSvc.List(ctx, noticesvc.ListInput{
		PageNum:   req.PageNum,
		PageSize:  req.PageSize,
		Title:     req.Title,
		Type:      req.Type,
		CreatedBy: req.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.ListItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, &v1.ListItem{
			SysNotice:     item.SysNotice,
			CreatedByName: item.CreatedByName,
		})
	}
	return &v1.ListRes{List: items, Total: out.Total}, nil
}
