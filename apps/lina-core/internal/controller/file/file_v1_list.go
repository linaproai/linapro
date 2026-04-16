package file

import (
	"context"

	v1 "lina-core/api/file/v1"
	filesvc "lina-core/internal/service/file"
)

// List queries file list
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	out, err := c.fileSvc.List(ctx, &filesvc.ListInput{
		PageNum:        req.PageNum,
		PageSize:       req.PageSize,
		Name:           req.Name,
		Original:       req.Original,
		Suffix:         req.Suffix,
		Scene:          req.Scene,
		BeginTime:      req.BeginTime,
		EndTime:        req.EndTime,
		OrderBy:        req.OrderBy,
		OrderDirection: req.OrderDirection,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*v1.ListItem, len(out.List))
	for i, item := range out.List {
		items[i] = &v1.ListItem{
			SysFile:       item.SysFile,
			CreatedByName: item.CreatedByName,
		}
	}
	return &v1.ListRes{
		List:  items,
		Total: out.Total,
	}, nil
}
