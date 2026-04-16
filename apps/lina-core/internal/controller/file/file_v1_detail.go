package file

import (
	"context"

	v1 "lina-core/api/file/v1"
)

// Detail returns file details
func (c *ControllerV1) Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error) {
	out, err := c.fileSvc.Detail(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.DetailRes{
		SysFile:       out.SysFile,
		CreatedByName: out.CreatedByName,
		SceneLabel:    out.SceneLabel,
	}, nil
}
