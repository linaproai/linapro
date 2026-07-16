package file

import (
	"context"

	v1 "lina-core/api/file/v1"
	filesvc "lina-core/internal/service/file"
)

// DirectUploadComplete validates the client-uploaded object and creates file metadata.
func (c *ControllerV1) DirectUploadComplete(ctx context.Context, req *v1.DirectUploadCompleteReq) (res *v1.DirectUploadCompleteRes, err error) {
	parts := make([]filesvc.MultipartPartRef, 0, len(req.Parts))
	for _, part := range req.Parts {
		parts = append(parts, filesvc.MultipartPartRef{PartNumber: part.PartNumber, ETag: part.ETag})
	}
	out, err := c.fileSvc.DirectUploadComplete(ctx, &filesvc.DirectUploadCompleteInput{
		UploadSessionID: req.UploadSessionId,
		ETag:            req.ETag,
		Parts:           parts,
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
