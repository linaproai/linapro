package dict

import (
	"context"

	v1 "lina-core/api/dict/v1"
)

// TypeOptions returns dictionary type option list.
func (c *ControllerV1) TypeOptions(ctx context.Context, req *v1.TypeOptionsReq) (res *v1.TypeOptionsRes, err error) {
	options, err := c.dictSvc.Options(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.DictTypeOptionItem, 0, len(options))
	for _, opt := range options {
		list = append(list, &v1.DictTypeOptionItem{
			Id:   opt.Id,
			Name: opt.Name,
			Type: opt.Type,
		})
	}
	return &v1.TypeOptionsRes{List: list}, nil
}
