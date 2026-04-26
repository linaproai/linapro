// This file defines the demo-record create DTOs for the dynamic plugin sample.

package v1

import "github.com/gogf/gf/v2/util/gmeta"

// CreateDemoRecordReq is the request for creating one dynamic demo record.
type CreateDemoRecordReq struct {
	gmeta.Meta              `path:"/demo-records" method:"post" tags:"Dynamic Plugin Demo" summary:"Create dynamic plugin sample record" dc:"Create a plugin-demo-dynamic sample record with an optional plugin-owned attachment, demonstrating writes to the data table and authorized storage file created by the installation SQL." access:"login" permission:"plugin-demo-dynamic:record:create" operLog:"create"`
	Title                   string `json:"title" v:"required|length:1,128" dc:"Record title" eg:"Dynamic plugin SQL sample records"`
	Content                 string `json:"content" v:"max-length:1000" dc:"Record content" eg:"This record is created by the dynamic plugin sample page to demonstrate the new operations of the SQL data table."`
	AttachmentName          string `json:"attachmentName" dc:"The original file name of the attachment; pass an empty string when no attachment is uploaded." eg:"plugin-demo-dynamic-note.txt"`
	AttachmentContentBase64 string `json:"attachmentContentBase64" dc:"Base64 encoding of the attachment content; pass an empty string when no attachment is uploaded" eg:"SGVsbG8sIHBsdWdpbi1kZW1vLWR5bmFtaWMh"`
	AttachmentContentType   string `json:"attachmentContentType" dc:"Attachment content type; pass an empty string when no attachment is uploaded" eg:"text/plain"`
}

// CreateDemoRecordRes is the response for creating one dynamic demo record.
type CreateDemoRecordRes struct {
	DemoRecordItem
}
