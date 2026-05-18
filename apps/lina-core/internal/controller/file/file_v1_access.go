// This file implements storage-backed access to uploaded file URLs.

package file

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/file/v1"
)

// Access streams an uploaded file by storage path through the file service.
func (c *ControllerV1) Access(ctx context.Context, req *v1.AccessReq) (res *v1.AccessRes, err error) {
	r := g.RequestFromCtx(ctx)

	fileStream, err := c.fileSvc.OpenByPath(ctx, req.Path)
	if err != nil {
		return nil, err
	}
	return nil, writeFileStream(ctx, r, fileStream, false)
}
