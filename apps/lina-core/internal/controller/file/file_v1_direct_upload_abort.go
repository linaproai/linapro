package file

import (
	"context"

	v1 "lina-core/api/file/v1"
	filesvc "lina-core/internal/service/file"
)

// DirectUploadAbort discards an in-flight direct upload session.
func (c *ControllerV1) DirectUploadAbort(ctx context.Context, req *v1.DirectUploadAbortReq) (res *v1.DirectUploadAbortRes, err error) {
	if err = c.fileSvc.DirectUploadAbort(ctx, &filesvc.DirectUploadAbortInput{
		UploadSessionID: req.UploadSessionId,
	}); err != nil {
		return nil, err
	}
	return &v1.DirectUploadAbortRes{}, nil
}
