// demo_v1_record_list.go implements the plugin-demo-source record list HTTP handler.

package demo

import (
	"context"

	"lina-plugin-demo-source/backend/api/demo/v1"
	demosvc "lina-plugin-demo-source/backend/internal/service/demo"
)

// ListRecords returns one paged record list for the source-plugin CRUD page.
func (c *ControllerV1) ListRecords(
	ctx context.Context,
	req *v1.ListRecordsReq,
) (res *v1.ListRecordsRes, err error) {
	out, err := c.demoSvc.ListRecords(ctx, &demosvc.ListRecordsInput{
		Keyword:  req.Keyword,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	items := make([]*v1.RecordItem, 0, len(out.List))
	for _, item := range out.List {
		items = append(items, &v1.RecordItem{
			Id:             item.Id,
			Title:          item.Title,
			Content:        item.Content,
			AttachmentName: item.AttachmentName,
			HasAttachment:  boolToInt(item.HasAttachment),
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
		})
	}
	return &v1.ListRecordsRes{List: items, Total: out.Total}, nil
}
