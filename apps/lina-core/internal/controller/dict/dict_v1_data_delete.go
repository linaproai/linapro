package dict

import (
	"context"

	v1 "lina-core/api/dict/v1"
)

// DataDelete deletes a dictionary data entry by ID.
func (c *ControllerV1) DataDelete(ctx context.Context, req *v1.DataDeleteReq) (res *v1.DataDeleteRes, err error) {
	err = c.dictSvc.DataDelete(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.DataDeleteRes{}, nil
}
