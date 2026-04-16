// This file implements the v1 file-download HTTP handler.

package file

import (
	"context"
	"io"
	"net/http"

	"github.com/gogf/gf/v2/frame/g"

	v1 "lina-core/api/file/v1"
	"lina-core/pkg/closeutil"
)

// Download downloads a file
func (c *ControllerV1) Download(ctx context.Context, req *v1.DownloadReq) (res *v1.DownloadRes, err error) {
	r := g.RequestFromCtx(ctx)

	fileInfo, err := c.fileSvc.Info(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	reader, err := c.fileSvc.GetStorage().Get(ctx, fileInfo.Path)
	if err != nil {
		return nil, err
	}
	defer closeutil.Close(reader, &err, "关闭下载文件流失败")

	r.Response.Header().Set("Content-Disposition", "attachment; filename=\""+fileInfo.Original+"\"")
	r.Response.Header().Set("Content-Type", "application/octet-stream")
	if isPreviewable(fileInfo.Suffix) {
		contentType := mimeTypeByExt(fileInfo.Suffix)
		r.Response.Header().Set("Content-Type", contentType)
	}

	if _, err = io.Copy(r.Response.RawWriter(), reader); err != nil {
		return nil, err
	}
	r.ExitAll()
	return nil, nil
}

// isPreviewable checks if a file is previewable
func isPreviewable(suffix string) bool {
	switch suffix {
	case "jpg", "jpeg", "png", "gif", "webp", "svg", "pdf":
		return true
	}
	return false
}

// mimeTypeByExt returns MIME type by file extension
func mimeTypeByExt(suffix string) string {
	mimeTypes := map[string]string{
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"png":  "image/png",
		"gif":  "image/gif",
		"webp": "image/webp",
		"svg":  "image/svg+xml",
		"pdf":  "application/pdf",
	}
	if mt, ok := mimeTypes[suffix]; ok {
		return mt
	}
	return http.DetectContentType(nil)
}
