// notice_v1_update.go implements the controller method that updates an
// existing notice record.

package notice

import (
	"context"

	v1 "lina-plugin-content-notice/backend/api/notice/v1"
	noticesvc "lina-plugin-content-notice/backend/internal/service/notice"
)

// Update updates a notice
func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	err = c.noticeSvc.Update(ctx, noticesvc.UpdateInput{
		Id:      req.Id,
		Title:   req.Title,
		Type:    req.Type,
		Content: req.Content,
		FileIds: req.FileIds,
		Status:  req.Status,
		Remark:  req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRes{}, nil
}
