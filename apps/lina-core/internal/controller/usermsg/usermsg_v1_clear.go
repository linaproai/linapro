package usermsg

import (
	"context"

	v1 "lina-core/api/usermsg/v1"
)

// Clear clears user messages
func (c *ControllerV1) Clear(ctx context.Context, req *v1.ClearReq) (res *v1.ClearRes, err error) {
	err = c.usermsgSvc.Clear(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ClearRes{}, nil
}
