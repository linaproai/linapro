package usermsg

import (
	"context"

	v1 "lina-core/api/usermsg/v1"
)

// ReadAll marks all messages as read
func (c *ControllerV1) ReadAll(ctx context.Context, req *v1.ReadAllReq) (res *v1.ReadAllRes, err error) {
	err = c.usermsgSvc.MarkReadAll(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ReadAllRes{}, nil
}
