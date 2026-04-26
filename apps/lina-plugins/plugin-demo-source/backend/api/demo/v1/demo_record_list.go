// demo_record_list.go defines the request and response DTOs for querying
// plugin-demo-source records.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ListRecordsReq is the request for querying plugin-demo-source records.
type ListRecordsReq struct {
	g.Meta   `path:"/plugins/plugin-demo-source/records" method:"get" tags:"Source Plugin Demo" summary:"List source plugin sample records" dc:"List business records for the plugin-demo-source sample page, with optional fuzzy title filtering. This demonstrates CRUD operations against a table created by the source plugin installation SQL." permission:"plugin-demo-source:example:view"`
	PageNum  int    `json:"pageNum" dc:"Page number, defaults to page 1 when omitted" eg:"1"`
	PageSize int    `json:"pageSize" dc:"Number of items per page; defaults to 10 when omitted" eg:"10"`
	Keyword  string `json:"keyword" dc:"Fuzzy filter by Record title; queries all records when omitted" eg:"Example"`
}

// ListRecordsRes is the response for querying plugin-demo-source records.
type ListRecordsRes struct {
	List  []*RecordItem `json:"list" dc:"Records on the current page" eg:"[]"`
	Total int           `json:"total" dc:"Total number of matching records" eg:"1"`
}

// RecordItem defines one plugin-demo-source record row.
type RecordItem struct {
	Id             int64  `json:"id" dc:"Record ID" eg:"1"`
	Title          string `json:"title" dc:"Record title" eg:"Source plugin SQL sample record"`
	Content        string `json:"content" dc:"Summary of record content" eg:"This record is used to demonstrate how the source plugin page operates the data table created by installing SQL."`
	AttachmentName string `json:"attachmentName" dc:"The original file name of the attachment. If there is no attachment, an empty string is returned." eg:"plugin-demo-source-note.txt"`
	HasAttachment  int    `json:"hasAttachment" dc:"Whether the attachment exists: 1=exists 0=does not exist" eg:"1"`
	CreatedAt      string `json:"createdAt" dc:"Creation time" eg:"2026-04-16 18:00:00"`
	UpdatedAt      string `json:"updatedAt" dc:"Update time" eg:"2026-04-16 18:05:00"`
}
