package dict

import (
	"context"

	v1 "lina-core/api/dict/v1"
)

// DataGet returns dictionary data details by ID.
func (c *ControllerV1) DataGet(ctx context.Context, req *v1.DataGetReq) (res *v1.DataGetRes, err error) {
	dictData, err := c.dictSvc.DataGetById(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.DataGetRes{SysDictData: dictData}, nil
}
