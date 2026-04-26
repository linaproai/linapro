// This file defines the demo-record update DTOs for the dynamic plugin sample.

package v1

import "github.com/gogf/gf/v2/util/gmeta"

// UpdateDemoRecordReq is the request for updating one dynamic demo record.
type UpdateDemoRecordReq struct {
	gmeta.Meta              `path:"/demo-records/{id}" method:"put" tags:"Dynamic Plugin Demo" summary:"Update dynamic plugin sample record" dc:"Update a plugin-demo-dynamic sample record and optionally replace or remove its attachment, demonstrating writes to plugin-owned tables and authorized storage files." access:"login" permission:"plugin-demo-dynamic:record:update" operLog:"update"`
	Id                      string `json:"id" v:"required|length:1,64" dc:"Record unique identifier" eg:"demo-record-1"`
	Title                   string `json:"title" v:"required|length:1,128" dc:"Record title" eg:"Dynamic plugin SQL sample records"`
	Content                 string `json:"content" v:"max-length:1000" dc:"Record content" eg:"Updated dynamic plugin example logging content."`
	AttachmentName          string `json:"attachmentName" dc:"The original file name of the newly uploaded attachment; pass an empty string when no new attachment is uploaded." eg:"plugin-demo-dynamic-note.txt"`
	AttachmentContentBase64 string `json:"attachmentContentBase64" dc:"Base64 encoding of the newly uploaded attachment content; pass an empty string when no new attachment is uploaded" eg:"SGVsbG8sIHVwZGF0ZWQgZHluYW1pYyBwbHVnaW4h"`
	AttachmentContentType   string `json:"attachmentContentType" dc:"The content type of the newly uploaded attachment; pass an empty string when no new attachment is uploaded" eg:"text/plain"`
	RemoveAttachment        bool   `json:"removeAttachment" dc:"Whether to remove the current attachment: true=remove false=keep. When a new attachment is uploaded, it replaces the old attachment." eg:"false"`
}

// UpdateDemoRecordRes is the response for updating one dynamic demo record.
type UpdateDemoRecordRes struct {
	DemoRecordItem
}
