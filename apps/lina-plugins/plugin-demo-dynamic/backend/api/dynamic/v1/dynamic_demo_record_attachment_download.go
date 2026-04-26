// This file defines the demo-record attachment download DTOs for the dynamic
// plugin sample.

package v1

import "github.com/gogf/gf/v2/util/gmeta"

// DownloadDemoRecordAttachmentReq is the request for downloading one dynamic demo-record attachment.
type DownloadDemoRecordAttachmentReq struct {
	gmeta.Meta `path:"/demo-records/{id}/attachment" method:"get" tags:"Dynamic Plugin Demo" summary:"Download the dynamic plugin sample attachment" dc:"Download the attachment currently associated with a plugin-demo-dynamic sample record, demonstrating reads from plugin-owned storage files." access:"login" permission:"plugin-demo-dynamic:record:view" operLog:"other"`
	Id         string `json:"id" v:"required|length:1,64" dc:"Record unique identifier" eg:"demo-record-1"`
}

// DownloadDemoRecordAttachmentRes is the response placeholder for streamed attachment downloads.
type DownloadDemoRecordAttachmentRes struct{}
