// notice_v1_create.go implements the controller method that creates a new
// notice record.

package notice

import (
	"context"

	v1 "lina-plugin-content-notice/backend/api/notice/v1"
	noticesvc "lina-plugin-content-notice/backend/internal/service/notice"
)

// Create creates a notice
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	status := 0
	if req.Status != nil {
		status = *req.Status
	}
	id, err := c.noticeSvc.Create(ctx, noticesvc.CreateInput{
		Title:   req.Title,
		Type:    req.Type,
		Content: req.Content,
		FileIds: req.FileIds,
		Status:  status,
		Remark:  req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{Id: id}, nil
}
