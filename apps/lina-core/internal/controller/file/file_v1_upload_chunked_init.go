package file

import (
	"context"

	v1 "lina-core/api/file/v1"
	filesvc "lina-core/internal/service/file"
)

// ChunkedUploadInit starts a host-mediated chunked upload session.
func (c *ControllerV1) ChunkedUploadInit(ctx context.Context, req *v1.ChunkedUploadInitReq) (res *v1.ChunkedUploadInitRes, err error) {
	out, err := c.fileSvc.ChunkedUploadInit(ctx, &filesvc.ChunkedUploadInitInput{
		Scene:       req.Scene,
		FileName:    req.FileName,
		Size:        req.Size,
		ContentType: req.ContentType,
		ContentHash: req.ContentHash,
	})
	if err != nil {
		return nil, err
	}
	res = &v1.ChunkedUploadInitRes{UploadSessionId: out.UploadSessionID}
	if out.Strategy != nil {
		res.Strategy = &v1.UploadStrategy{Channel: out.Strategy.Channel, Encoding: out.Strategy.Encoding}
	}
	if out.Multipart != nil {
		res.Multipart = &v1.UploadMultipartPlan{
			PartSize:       out.Multipart.PartSize,
			MinPartSize:    out.Multipart.MinPartSize,
			MaxParts:       out.Multipart.MaxParts,
			MaxConcurrency: out.Multipart.MaxConcurrency,
		}
	}
	return res, nil
}
