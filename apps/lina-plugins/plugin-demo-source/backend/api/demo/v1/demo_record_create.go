// demo_record_create.go defines the request and response DTOs for creating one
// plugin-demo-source record.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CreateRecordReq is the request for creating one plugin-demo-source record.
type CreateRecordReq struct {
	g.Meta  `path:"/plugins/plugin-demo-source/records" method:"post" mime:"multipart/form-data" tags:"Source Plugin Demo" summary:"Create source plugin sample record" dc:"Create a plugin-demo-source sample record with an optional plugin-owned attachment, demonstrating writes to the data table and storage files created by the installation SQL." permission:"plugin-demo-source:example:create"`
	Title   string `json:"title" v:"required|length:1,128" dc:"Record title" eg:"Source plugin SQL sample record"`
	Content string `json:"content" dc:"Record content" eg:"This record is used to demonstrate how the source plugin page operates the data table created by installing SQL."`
}

// CreateRecordRes is the response for creating one plugin-demo-source record.
type CreateRecordRes struct {
	Id int64 `json:"id" dc:"The newly created record ID" eg:"1"`
}
