// Demo-record list route controller.

package dynamic

import (
	"context"

	"lina-plugin-demo-dynamic/backend/api/dynamic/v1"
	dynamicservice "lina-plugin-demo-dynamic/backend/internal/service/dynamic"
)

// DemoRecordList returns one paged list of plugin-owned demo records.
func (c *Controller) DemoRecordList(
	_ context.Context,
	req *v1.DemoRecordListReq,
) (res *v1.DemoRecordListRes, err error) {
	payload, err := c.dynamicSvc.ListDemoRecordsPayload(&dynamicservice.DemoRecordListInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Keyword:  req.Keyword,
	})
	if err != nil {
		return nil, wrapDynamicError(err)
	}
	items := make([]*v1.DemoRecordItem, 0, len(payload.List))
	for _, item := range payload.List {
		items = append(items, &v1.DemoRecordItem{
			Id:             item.Id,
			Title:          item.Title,
			Content:        item.Content,
			AttachmentName: item.AttachmentName,
			HasAttachment:  item.HasAttachment,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
		})
	}
	return &v1.DemoRecordListRes{
		List:  items,
		Total: payload.Total,
	}, nil
}
