// demo_v1_record_attachment_download.go implements the plugin-demo-source attachment download HTTP handler.

package demo

import (
	"context"
	"io"
	"os"

	"github.com/gogf/gf/v2/frame/g"

	"lina-plugin-demo-source/backend/api/demo/v1"
)

// DownloadAttachment streams one plugin-owned attachment file to the client.
func (c *ControllerV1) DownloadAttachment(
	ctx context.Context,
	req *v1.DownloadAttachmentReq,
) (res *v1.DownloadAttachmentRes, err error) {
	out, err := c.demoSvc.BuildAttachmentDownload(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	fileReader, err := os.Open(out.FullPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := fileReader.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	r := g.RequestFromCtx(ctx)
	r.Response.Header().Set("Content-Disposition", "attachment; filename=\""+out.OriginalName+"\"")
	r.Response.Header().Set("Content-Type", out.ContentType)
	if _, err = io.Copy(r.Response.RawWriter(), fileReader); err != nil {
		return nil, err
	}
	r.ExitAll()
	return nil, nil
}
