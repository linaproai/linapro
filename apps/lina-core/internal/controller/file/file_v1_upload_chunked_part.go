package file

import (
	"context"
	"io"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/file/v1"
	filesvc "lina-core/internal/service/file"
)

// ChunkedUploadPart uploads one part of a host-mediated chunked session.
func (c *ControllerV1) ChunkedUploadPart(ctx context.Context, req *v1.ChunkedUploadPartReq) (res *v1.ChunkedUploadPartRes, err error) {
	r := g.RequestFromCtx(ctx)
	uploadFile := r.GetUploadFile("file")
	if uploadFile == nil {
		uploadFile = req.File
	}
	var (
		body io.Reader
		size int64 = -1
	)
	if uploadFile != nil {
		opened, openErr := uploadFile.Open()
		if openErr != nil {
			return nil, openErr
		}
		defer func() { _ = opened.Close() }()
		body = opened
		size = uploadFile.Size
	}
	out, err := c.fileSvc.ChunkedUploadPart(ctx, &filesvc.ChunkedUploadPartInput{
		UploadSessionID: req.UploadSessionId,
		PartNumber:      req.PartNumber,
		Body:            body,
		Size:            size,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ChunkedUploadPartRes{
		PartNumber:    out.PartNumber,
		ETag:          out.ETag,
		ReceivedBytes: out.ReceivedBytes,
	}, nil
}
