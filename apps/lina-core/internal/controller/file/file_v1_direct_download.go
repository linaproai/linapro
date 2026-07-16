package file

import (
	"context"

	v1 "lina-core/api/file/v1"
	filesvc "lina-core/internal/service/file"
)

// DirectDownload issues short-lived client get access or proxy mode for one file.
func (c *ControllerV1) DirectDownload(ctx context.Context, req *v1.DirectDownloadReq) (res *v1.DirectDownloadRes, err error) {
	out, err := c.fileSvc.DirectDownload(ctx, &filesvc.DirectDownloadInput{ID: req.Id})
	if err != nil {
		return nil, err
	}
	return &v1.DirectDownloadRes{
		Access:   mapDirectAccess(out.Access),
		ProxyUrl: out.ProxyURL,
	}, nil
}
