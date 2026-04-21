// notice_v1_get.go implements the controller method that serves the notice
// detail endpoint.

package notice

import (
	"context"

	v1 "lina-plugin-content-notice/backend/api/notice/v1"
)

// Get returns notice details
func (c *ControllerV1) Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error) {
	item, err := c.noticeSvc.GetById(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetRes{
		NoticeEntity:  toAPINoticeEntity(item.NoticeEntity),
		CreatedByName: item.CreatedByName,
	}, nil
}
