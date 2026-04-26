// This file defines the demo-record detail DTOs for the dynamic plugin sample.

package v1

import "github.com/gogf/gf/v2/util/gmeta"

// DemoRecordReq is the request for querying one dynamic demo record detail.
type DemoRecordReq struct {
	gmeta.Meta `path:"/demo-records/{id}" method:"get" tags:"Dynamic Plugin Demo" summary:"Query dynamic plugin example record details" dc:"Get plugin-demo-dynamic sample record details for edit-form backfill and attachment checks before download." access:"login" permission:"plugin-demo-dynamic:record:view" operLog:"other"`
	Id         string `json:"id" v:"required|length:1,64" dc:"Record unique identifier" eg:"demo-record-1"`
}

// DemoRecordRes is the response for querying one dynamic demo record detail.
type DemoRecordRes struct {
	DemoRecordItem
}

// DemoRecordItem defines one dynamic plugin demo-record row.
type DemoRecordItem struct {
	Id             string `json:"id" dc:"Record unique identifier" eg:"demo-record-1"`
	Title          string `json:"title" dc:"Record title" eg:"Dynamic plugin SQL sample records"`
	Content        string `json:"content" dc:"Record content" eg:"This record demonstrates CRUD operations against the data table created by the dynamic plugin installation SQL."`
	AttachmentName string `json:"attachmentName" dc:"The original file name of the attachment. If there is no attachment, an empty string is returned." eg:"plugin-demo-dynamic-note.txt"`
	HasAttachment  bool   `json:"hasAttachment" dc:"Whether the current attachment exists: true=exists false=does not exist" eg:"true"`
	CreatedAt      string `json:"createdAt" dc:"Record Creation time, automatically maintained by the default timestamp field of the sample data table" eg:"2026-04-16 10:00:00"`
	UpdatedAt      string `json:"updatedAt" dc:"Record the last update time, automatically maintained by the default timestamp field of the sample data table" eg:"2026-04-16 10:05:00"`
}
