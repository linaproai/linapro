package dict

import (
	"context"

	v1 "lina-core/api/dict/v1"
	dictsvc "lina-core/internal/service/dict"
)

// DataCreate creates a new dictionary data entry.
func (c *ControllerV1) DataCreate(ctx context.Context, req *v1.DataCreateReq) (res *v1.DataCreateRes, err error) {
	id, err := c.dictSvc.DataCreate(ctx, dictsvc.DataCreateInput{
		DictType: req.DictType,
		Label:    req.Label,
		Value:    req.Value,
		Sort:     *req.Sort,
		TagStyle: req.TagStyle,
		CssClass: req.CssClass,
		Status:   *req.Status,
		Remark:   req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.DataCreateRes{Id: id}, nil
}
