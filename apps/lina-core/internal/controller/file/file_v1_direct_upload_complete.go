package file

import (
	"context"

	v1 "lina-core/api/file/v1"
	filesvc "lina-core/internal/service/file"
)

// DirectUploadComplete validates the client-uploaded object and creates file metadata.
func (c *ControllerV1) DirectUploadComplete(ctx context.Context, req *v1.DirectUploadCompleteReq) (res *v1.DirectUploadCompleteRes, err error) {
	out, err := c.fileSvc.DirectUploadComplete(ctx, &filesvc.DirectUploadCompleteInput{
		UploadSessionID: req.UploadSessionId,
		ETag:            req.ETag,
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
