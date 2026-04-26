// demo_record_update.go defines the request and response DTOs for updating one
// plugin-demo-source record.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateRecordReq is the request for updating one plugin-demo-source record.
type UpdateRecordReq struct {
	g.Meta           `path:"/plugins/plugin-demo-source/records/{id}" method:"put" mime:"multipart/form-data" tags:"Source Plugin Demo" summary:"Update source plugin example record" dc:"Update a plugin-demo-source sample record and optionally replace or remove its attachment, demonstrating writes to plugin-owned tables and storage files." permission:"plugin-demo-source:example:update"`
	Id               int64  `json:"id" v:"required|min:1" dc:"Record ID" eg:"1"`
	Title            string `json:"title" v:"required|length:1,128" dc:"Record title" eg:"Source plugin SQL sample record"`
	Content          string `json:"content" dc:"Record content" eg:"Updated description content"`
	RemoveAttachment int    `json:"removeAttachment" dc:"Whether to remove the current attachment: 1=remove 0=keep. When a new file is uploaded, it replaces the old attachment." eg:"0"`
}

// UpdateRecordRes is the response for updating one plugin-demo-source record.
type UpdateRecordRes struct {
	Id int64 `json:"id" dc:"Updated record ID" eg:"1"`
}
