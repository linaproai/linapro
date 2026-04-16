package dict

import (
	"context"

	v1 "lina-core/api/dict/v1"
	dictsvc "lina-core/internal/service/dict"
)

// DataList returns dictionary data list.
func (c *ControllerV1) DataList(ctx context.Context, req *v1.DataListReq) (res *v1.DataListRes, err error) {
	out, err := c.dictSvc.DataList(ctx, dictsvc.DataListInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		DictType: req.DictType,
		Label:    req.Label,
	})
	if err != nil {
		return nil, err
	}
	return &v1.DataListRes{List: out.List, Total: out.Total}, nil
}
