package file

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/file/v1"
	filesvc "lina-core/internal/service/file"
)

// Upload uploads a file
func (c *ControllerV1) Upload(ctx context.Context, req *v1.UploadReq) (res *v1.UploadRes, err error) {
	r := g.RequestFromCtx(ctx)
	uploadFile := r.GetUploadFile("file")
	out, err := c.fileSvc.Upload(ctx, &filesvc.UploadInput{
		File:  uploadFile,
		Scene: req.Scene,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UploadRes{
		Id:       out.Id,
		Name:     out.Name,
		Original: out.Original,
		Url:      out.Url,
		Suffix:   out.Suffix,
		Size:     out.Size,
	}, nil
}
