package dict

import (
	"context"

	v1 "lina-core/api/dict/v1"
)

// TypeDelete deletes a dictionary type by ID.
func (c *ControllerV1) TypeDelete(ctx context.Context, req *v1.TypeDeleteReq) (res *v1.TypeDeleteRes, err error) {
	err = c.dictSvc.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.TypeDeleteRes{}, nil
}
