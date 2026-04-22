// Demo-record update route controller.

package dynamic

import (
	"context"

	"lina-plugin-demo-dynamic/backend/api/dynamic/v1"
	dynamicservice "lina-plugin-demo-dynamic/backend/internal/service/dynamic"
)

// UpdateDemoRecord updates one plugin-owned demo record.
func (c *Controller) UpdateDemoRecord(
	_ context.Context,
	req *v1.UpdateDemoRecordReq,
) (res *v1.UpdateDemoRecordRes, err error) {
	payload, err := c.dynamicSvc.UpdateDemoRecordPayload(req.Id, &dynamicservice.DemoRecordMutationInput{
		Title:                   req.Title,
		Content:                 req.Content,
		AttachmentName:          req.AttachmentName,
		AttachmentContentBase64: req.AttachmentContentBase64,
		AttachmentContentType:   req.AttachmentContentType,
		RemoveAttachment:        req.RemoveAttachment,
	})
	if err != nil {
		return nil, wrapDynamicError(err)
	}
	return &v1.UpdateDemoRecordRes{
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
