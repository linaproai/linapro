// notice_v1_list.go implements the controller method that serves the
// paged notice list endpoint.

package notice

import (
	"context"

	v1 "lina-plugin-content-notice/backend/api/notice/v1"
	noticesvc "lina-plugin-content-notice/backend/internal/service/notice"
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
			NoticeEntity:  toAPINoticeEntity(item.NoticeEntity),
			CreatedByName: item.CreatedByName,
		})
	}
	return &v1.ListRes{List: items, Total: out.Total}, nil
}
