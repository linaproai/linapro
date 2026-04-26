// demo_record_delete.go defines the request and response DTOs for deleting one
// plugin-demo-source record.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteRecordReq is the request for deleting one plugin-demo-source record.
type DeleteRecordReq struct {
	g.Meta `path:"/plugins/plugin-demo-source/records/{id}" method:"delete" tags:"Source Plugin Demo" summary:"Delete source plugin sample record" dc:"Delete a plugin-demo-source sample record and clean up its plugin-owned attachment file." permission:"plugin-demo-source:example:delete"`
	Id     int64 `json:"id" v:"required|min:1" dc:"Record ID" eg:"1"`
}

// DeleteRecordRes is the response for deleting one plugin-demo-source record.
type DeleteRecordRes struct {
	Id int64 `json:"id" dc:"Deleted record ID" eg:"1"`
}
