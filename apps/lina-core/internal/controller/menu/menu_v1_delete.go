package menu

import (
	"context"

	"lina-core/api/menu/v1"
	menusvc "lina-core/internal/service/menu"
)

// Delete deletes the specified menu.
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	err = c.menuSvc.Delete(ctx, menusvc.DeleteInput{
		Id:            req.Id,
		CascadeDelete: req.CascadeDelete,
	})
	if err != nil {
		return nil, err
	}

	return &v1.DeleteRes{}, nil
}
