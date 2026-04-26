// This file defines the demo-record list DTOs for the dynamic plugin sample.

package v1

import "github.com/gogf/gf/v2/util/gmeta"

// DemoRecordListReq is the request for querying dynamic demo records.
type DemoRecordListReq struct {
	gmeta.Meta `path:"/demo-records" method:"get" tags:"Dynamic Plugin Demo" summary:"List dynamic plugin sample records" dc:"List business records for the plugin-demo-dynamic sample page, with optional fuzzy title filtering. This demonstrates CRUD operations against a table created by the dynamic plugin installation SQL." access:"login" permission:"plugin-demo-dynamic:record:view" operLog:"other"`
	PageNum    int    `json:"pageNum" dc:"Page number; defaults to 1 when omitted" eg:"1"`
	PageSize   int    `json:"pageSize" dc:"Number of items per page; defaults to 20 and is capped at 100 when omitted" eg:"20"`
	Keyword    string `json:"keyword" dc:"Fuzzy filter by Record title keyword; queries all records when omitted" eg:"SQL"`
}

// DemoRecordListRes is the response for querying dynamic demo records.
type DemoRecordListRes struct {
	List  []*DemoRecordItem `json:"list" dc:"Records on the current page" eg:"[{\"id\":\"demo-record-1\",\"title\":\"Dynamic plugin SQL sample records\"}]"`
	Total int               `json:"total" dc:"Total number of records matching the filter criteria" eg:"1"`
}
