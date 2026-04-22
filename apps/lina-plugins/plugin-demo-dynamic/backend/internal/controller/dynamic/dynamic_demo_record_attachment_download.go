// Demo-record attachment download route controller.

package dynamic

import (
	"context"
	"fmt"

	"lina-core/pkg/pluginbridge"
	"lina-plugin-demo-dynamic/backend/api/dynamic/v1"
)

// DownloadDemoRecordAttachment streams one plugin-owned attachment file.
func (c *Controller) DownloadDemoRecordAttachment(
	ctx context.Context,
	req *v1.DownloadDemoRecordAttachmentReq,
) (res *v1.DownloadDemoRecordAttachmentRes, err error) {
	payload, err := c.dynamicSvc.BuildDemoRecordAttachmentDownload(req.Id)
	if err != nil {
		return nil, wrapDynamicError(err)
	}
	if err = pluginbridge.SetResponseHeader(
		ctx,
		"Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s"`, payload.OriginalName),
	); err != nil {
		return nil, err
	}
	if err = pluginbridge.WriteResponse(ctx, 200, payload.ContentType, payload.Body); err != nil {
		return nil, err
	}
	return nil, nil
}
