// Demo-record create route controller.

package dynamic

import (
	"context"

	"lina-plugin-demo-dynamic/backend/api/dynamic/v1"
	dynamicservice "lina-plugin-demo-dynamic/backend/internal/service/dynamic"
)

// CreateDemoRecord creates one plugin-owned demo record.
func (c *Controller) CreateDemoRecord(
	_ context.Context,
	req *v1.CreateDemoRecordReq,
) (res *v1.CreateDemoRecordRes, err error) {
	payload, err := c.dynamicSvc.CreateDemoRecordPayload(&dynamicservice.DemoRecordMutationInput{
		Title:                   req.Title,
		Content:                 req.Content,
		AttachmentName:          req.AttachmentName,
		AttachmentContentBase64: req.AttachmentContentBase64,
		AttachmentContentType:   req.AttachmentContentType,
	})
	if err != nil {
		return nil, wrapDynamicError(err)
	}
	return &v1.CreateDemoRecordRes{
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
