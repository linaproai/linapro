// This file centralizes binary file response streaming for file controllers.

package file

import (
	"context"
	"io"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	filesvc "lina-core/internal/service/file"
	"lina-core/pkg/closeutil"
)

// writeFileStream writes one storage-backed file stream to the HTTP response.
func writeFileStream(
	ctx context.Context,
	r *ghttp.Request,
	fileStream *filesvc.OpenOutput,
	attachment bool,
) (err error) {
	if r == nil || fileStream == nil || fileStream.Reader == nil {
		return nil
	}
	defer closeutil.Close(ctx, fileStream.Reader, &err, "close file stream failed")

	contentType := strings.TrimSpace(fileStream.ContentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	r.Response.Header().Set("Content-Type", contentType)
	if fileStream.Size > 0 {
		r.Response.Header().Set("Content-Length", strconv.FormatInt(fileStream.Size, 10))
	}
	if attachment {
		r.Response.Header().Set(
			"Content-Disposition",
			"attachment; filename=\""+sanitizeContentDispositionFilename(fileStream.Original)+"\"",
		)
	}

	if _, err = io.Copy(r.Response.RawWriter(), fileStream.Reader); err != nil {
		return err
	}
	r.ExitAll()
	return nil
}

// sanitizeContentDispositionFilename keeps response headers valid for common
// filenames without changing the stored metadata.
func sanitizeContentDispositionFilename(filename string) string {
	replacer := strings.NewReplacer("\\", "_", "\"", "_", "\r", "_", "\n", "_")
	return replacer.Replace(filename)
}
