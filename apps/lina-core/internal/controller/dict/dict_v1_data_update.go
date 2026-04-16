package dict

import (
	"context"

	v1 "lina-core/api/dict/v1"
	dictsvc "lina-core/internal/service/dict"
)

// DataUpdate updates a dictionary data entry.
func (c *ControllerV1) DataUpdate(ctx context.Context, req *v1.DataUpdateReq) (res *v1.DataUpdateRes, err error) {
	err = c.dictSvc.DataUpdate(ctx, dictsvc.DataUpdateInput{
		Id:       req.Id,
		DictType: req.DictType,
		Label:    req.Label,
		Value:    req.Value,
		Sort:     req.Sort,
		TagStyle: req.TagStyle,
		CssClass: req.CssClass,
		Status:   req.Status,
		Remark:   req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.DataUpdateRes{}, nil
}
