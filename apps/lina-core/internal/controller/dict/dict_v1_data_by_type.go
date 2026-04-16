package dict

import (
	"context"

	v1 "lina-core/api/dict/v1"
)

// DataByType returns dictionary data by dictionary type.
func (c *ControllerV1) DataByType(ctx context.Context, req *v1.DataByTypeReq) (res *v1.DataByTypeRes, err error) {
	list, err := c.dictSvc.DataByType(ctx, req.DictType)
	if err != nil {
		return nil, err
	}
	return &v1.DataByTypeRes{List: list}, nil
}
