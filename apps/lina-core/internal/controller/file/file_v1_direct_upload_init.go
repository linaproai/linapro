package file

import (
	"context"

	v1 "lina-core/api/file/v1"
	filesvc "lina-core/internal/service/file"
	"lina-core/pkg/plugin/capability/storagecap"
)

// DirectUploadInit starts a client direct upload session or returns proxy/instant reuse.
func (c *ControllerV1) DirectUploadInit(ctx context.Context, req *v1.DirectUploadInitReq) (res *v1.DirectUploadInitRes, err error) {
	out, err := c.fileSvc.DirectUploadInit(ctx, &filesvc.DirectUploadInitInput{
		Scene:       req.Scene,
		FileName:    req.FileName,
		Size:        req.Size,
		ContentType: req.ContentType,
		ContentHash: req.ContentHash,
	})
	if err != nil {
		return nil, err
	}
	res = &v1.DirectUploadInitRes{
		InstantReuse:    out.InstantReuse,
		UploadSessionId: out.UploadSessionID,
		Access:          mapDirectAccess(out.Access),
	}
	if out.File != nil {
		res.File = &v1.UploadRes{
			Id:       out.File.Id,
			Name:     out.File.Name,
			Original: out.File.Original,
			Url:      out.File.Url,
			Suffix:   out.File.Suffix,
			Size:     out.File.Size,
		}
	}
	return res, nil
}

func mapDirectAccess(access *storagecap.DirectAccess) *v1.DirectUploadAccess {
	if access == nil {
		return nil
	}
	out := &v1.DirectUploadAccess{
		Mode:            access.Mode,
		Operation:       string(access.Operation),
		Method:          access.Method,
		URL:             access.URL,
		Headers:         access.Headers,
		FormFields:      access.FormFields,
		AccessKeyID:     access.AccessKeyID,
		SecretAccessKey: access.SecretAccessKey,
		SessionToken:    access.SessionToken,
	}
	if !access.ExpiresAt.IsZero() {
		out.ExpiresAt = access.ExpiresAt.UnixMilli()
	}
	return out
}
