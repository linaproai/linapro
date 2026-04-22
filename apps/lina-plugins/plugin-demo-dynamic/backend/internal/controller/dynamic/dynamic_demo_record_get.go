// Demo-record detail route controller.

package dynamic

import (
	"context"

	"lina-plugin-demo-dynamic/backend/api/dynamic/v1"
)

// DemoRecord returns one plugin-owned demo record detail.
func (c *Controller) DemoRecord(
	_ context.Context,
	req *v1.DemoRecordReq,
) (res *v1.DemoRecordRes, err error) {
	payload, err := c.dynamicSvc.GetDemoRecordPayload(req.Id)
	if err != nil {
		return nil, wrapDynamicError(err)
	}
	return &v1.DemoRecordRes{
		DemoRecordItem: v1.DemoRecordItem{
			Id:             payload.Id,
			Title:          payload.Title,
			Content:        payload.Content,
			AttachmentName: payload.AttachmentName,
			HasAttachment:  payload.HasAttachment,
			CreatedAt:      payload.CreatedAt,
			UpdatedAt:      payload.UpdatedAt,
		},
	}, nil
}
