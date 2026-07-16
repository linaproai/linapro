package file

import (
	"context"

	v1 "lina-core/api/file/v1"
	filesvc "lina-core/internal/service/file"
)

// DirectUploadPartURL issues short-lived client access for one direct multipart part.
func (c *ControllerV1) DirectUploadPartURL(ctx context.Context, req *v1.DirectUploadPartURLReq) (res *v1.DirectUploadPartURLRes, err error) {
	out, err := c.fileSvc.DirectUploadPartURL(ctx, &filesvc.DirectUploadPartURLInput{
		UploadSessionID: req.UploadSessionId,
		PartNumber:      req.PartNumber,
		Size:            req.Size,
	})
	if err != nil {
		return nil, err
	}
	return &v1.DirectUploadPartURLRes{Access: mapDirectAccess(out.Access)}, nil
}
