package file

import (
	"context"

	v1 "lina-core/api/file/v1"
)

// InfoByIds returns file information by ID list.
func (c *ControllerV1) InfoByIds(ctx context.Context, req *v1.InfoByIdsReq) (res *v1.InfoByIdsRes, err error) {
	files, err := c.fileSvc.InfoByIds(ctx, req.Ids)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.FileItem, 0, len(files))
	for _, file := range files {
		item := fileItem(file)
		list = append(list, &item)
	}
	return &v1.InfoByIdsRes{List: list}, nil
}
