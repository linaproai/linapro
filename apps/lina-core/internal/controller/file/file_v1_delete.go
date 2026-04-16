package file

import (
	"context"

	v1 "lina-core/api/file/v1"
)

// Delete deletes a file
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	err = c.fileSvc.Delete(ctx, req.Ids)
	return nil, err
}
