// usermsg_v1_get.go implements the current-user message-detail endpoint used by
// the inbox preview dialog.

package usermsg

import (
	"context"

	"lina-core/api/usermsg/v1"
)

// Get returns one current-user message detail for inbox preview.
func (c *ControllerV1) Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error) {
	detail, err := c.usermsgSvc.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1.GetRes{
		Id:            detail.Id,
		Title:         detail.Title,
		Type:          detail.Type,
		SourceType:    detail.SourceType,
		SourceId:      detail.SourceId,
		Content:       detail.Content,
		CreatedByName: detail.CreatedByName,
		CreatedAt:     detail.CreatedAt,
	}, nil
}
