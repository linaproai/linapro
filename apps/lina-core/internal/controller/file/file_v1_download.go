// This file implements the v1 file-download HTTP handler.

package file

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/file/v1"
)

// Download downloads a file
func (c *ControllerV1) Download(ctx context.Context, req *v1.DownloadReq) (res *v1.DownloadRes, err error) {
	r := g.RequestFromCtx(ctx)

	fileStream, err := c.fileSvc.OpenByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return nil, writeFileStream(ctx, r, fileStream, true)
}
