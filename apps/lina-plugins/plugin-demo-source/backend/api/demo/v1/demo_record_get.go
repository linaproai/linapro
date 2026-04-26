// demo_record_get.go defines the request and response DTOs for querying one
// plugin-demo-source record detail.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetRecordReq is the request for querying one plugin-demo-source record detail.
type GetRecordReq struct {
	g.Meta `path:"/plugins/plugin-demo-source/records/{id}" method:"get" tags:"Source Plugin Demo" summary:"Get source plugin sample record details" dc:"Get plugin-demo-source sample record details for edit dialog backfill." permission:"plugin-demo-source:example:view"`
	Id     int64 `json:"id" v:"required|min:1" dc:"Record ID" eg:"1"`
}

// GetRecordRes is the response for querying one plugin-demo-source record detail.
type GetRecordRes struct {
	Id             int64  `json:"id" dc:"Record ID" eg:"1"`
	Title          string `json:"title" dc:"Record title" eg:"Source plugin SQL sample record"`
	Content        string `json:"content" dc:"Record content" eg:"This record is used to demonstrate how the source plugin page operates the data table created by installing SQL."`
	AttachmentName string `json:"attachmentName" dc:"The original file name of the attachment. If there is no attachment, an empty string is returned." eg:"plugin-demo-source-note.txt"`
	HasAttachment  int    `json:"hasAttachment" dc:"Whether the attachment exists: 1=exists 0=does not exist" eg:"1"`
}
