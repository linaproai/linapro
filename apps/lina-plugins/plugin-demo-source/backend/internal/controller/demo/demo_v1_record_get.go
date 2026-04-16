// demo_v1_record_get.go implements the plugin-demo-source record detail HTTP handler.

package demo

import (
	"context"

	"lina-plugin-demo-source/backend/api/demo/v1"
)

// GetRecord returns one demo record detail for edit forms.
func (c *ControllerV1) GetRecord(
	ctx context.Context,
	req *v1.GetRecordReq,
) (res *v1.GetRecordRes, err error) {
	out, err := c.demoSvc.GetRecord(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetRecordRes{
		Id:             out.Id,
		Title:          out.Title,
		Content:        out.Content,
		AttachmentName: out.AttachmentName,
		HasAttachment:  boolToInt(out.HasAttachment),
	}, nil
}
