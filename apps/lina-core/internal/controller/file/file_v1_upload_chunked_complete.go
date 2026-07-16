package file

import (
	"context"

	v1 "lina-core/api/file/v1"
	filesvc "lina-core/internal/service/file"
)

// ChunkedUploadComplete finalizes a host-mediated chunked upload.
func (c *ControllerV1) ChunkedUploadComplete(ctx context.Context, req *v1.ChunkedUploadCompleteReq) (res *v1.ChunkedUploadCompleteRes, err error) {
	parts := make([]filesvc.MultipartPartRef, 0, len(req.Parts))
	for _, part := range req.Parts {
		parts = append(parts, filesvc.MultipartPartRef{PartNumber: part.PartNumber, ETag: part.ETag})
	}
	out, err := c.fileSvc.ChunkedUploadComplete(ctx, &filesvc.ChunkedUploadCompleteInput{
		UploadSessionID: req.UploadSessionId,
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
