package notice

import (
	"context"

	v1 "lina-core/api/notice/v1"
)

// Get returns notice details
func (c *ControllerV1) Get(ctx context.Context, req *v1.GetReq) (res *v1.GetRes, err error) {
	item, err := c.noticeSvc.GetById(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetRes{
		SysNotice:     item.SysNotice,
		CreatedByName: item.CreatedByName,
	}, nil
}
