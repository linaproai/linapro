package dict

import (
	"context"

	v1 "lina-core/api/dict/v1"
	dictsvc "lina-core/internal/service/dict"
)

// TypeList returns dictionary type list.
func (c *ControllerV1) TypeList(ctx context.Context, req *v1.TypeListReq) (res *v1.TypeListRes, err error) {
	out, err := c.dictSvc.List(ctx, dictsvc.ListInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Name:     req.Name,
		Type:     req.Type,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.DictTypeItem, 0, len(out.List))
	for _, row := range out.List {
		item := dictTypeItem(row)
		list = append(list, &item)
	}
	return &v1.TypeListRes{List: list, Total: out.Total}, nil
}
