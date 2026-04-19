// Demo-record attachment download route controller.

package dynamic

import (
	"fmt"

	"lina-core/pkg/pluginbridge"
)

// DownloadDemoRecordAttachment streams one plugin-owned attachment file.
func (c *Controller) DownloadDemoRecordAttachment(request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	payload, err := c.dynamicSvc.BuildDemoRecordAttachmentDownload(readDemoRecordIDFromAttachmentDownloadRoute(request))
	if err != nil {
		return buildDynamicErrorResponse(err), nil
	}
	response := pluginbridge.NewSuccessResponse(200, payload.ContentType, payload.Body)
	response.Headers = map[string][]string{
		"Content-Disposition": {fmt.Sprintf(`attachment; filename="%s"`, payload.OriginalName)},
	}
	return response, nil
}

// readDemoRecordIDFromAttachmentDownloadRoute reads the record identifier from
// the attachment download path parameters.
func readDemoRecordIDFromAttachmentDownloadRoute(request *pluginbridge.BridgeRequestEnvelopeV1) string {
	return readDynamicPathParam(request, "id")
}
