// This file defines the demo-record delete DTOs for the dynamic plugin sample.

package v1

import "github.com/gogf/gf/v2/util/gmeta"

// DeleteDemoRecordReq is the request for deleting one dynamic demo record.
type DeleteDemoRecordReq struct {
	gmeta.Meta `path:"/demo-records/{id}" method:"delete" tags:"Dynamic Plugin Demo" summary:"Delete dynamic plugin example record" dc:"Delete a plugin-demo-dynamic sample record and clean up its plugin-owned attachment file." access:"login" permission:"plugin-demo-dynamic:record:delete" operLog:"delete"`
	Id         string `json:"id" v:"required|length:1,64" dc:"Record unique identifier" eg:"demo-record-1"`
}

// DeleteDemoRecordRes is the response for deleting one dynamic demo record.
type DeleteDemoRecordRes struct {
	Id      string `json:"id" dc:"Deleted Record unique identifier" eg:"demo-record-1"`
	Deleted bool   `json:"deleted" dc:"Deletion result: true=success false=failure" eg:"true"`
}
