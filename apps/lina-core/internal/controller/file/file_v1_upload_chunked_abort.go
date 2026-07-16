package file

import (
	"context"

	v1 "lina-core/api/file/v1"
	filesvc "lina-core/internal/service/file"
)

// ChunkedUploadAbort discards an in-flight chunked upload session.
func (c *ControllerV1) ChunkedUploadAbort(ctx context.Context, req *v1.ChunkedUploadAbortReq) (res *v1.ChunkedUploadAbortRes, err error) {
	if err = c.fileSvc.ChunkedUploadAbort(ctx, &filesvc.ChunkedUploadAbortInput{
		UploadSessionID: req.UploadSessionId,
	}); err != nil {
		return nil, err
	}
	return &v1.ChunkedUploadAbortRes{}, nil
}
