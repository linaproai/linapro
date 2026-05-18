package dict

import (
	"context"

	v1 "lina-core/api/dict/v1"
	dictsvc "lina-core/internal/service/dict"
)

// TypeCreate creates a new dictionary type.
func (c *ControllerV1) TypeCreate(ctx context.Context, req *v1.TypeCreateReq) (res *v1.TypeCreateRes, err error) {
	id, err := c.dictSvc.Create(ctx, dictsvc.CreateInput{
		Name:   req.Name,
		Type:   req.Type,
		Status: *req.Status,
		Remark: req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &v1.TypeCreateRes{Id: id}, nil
}
