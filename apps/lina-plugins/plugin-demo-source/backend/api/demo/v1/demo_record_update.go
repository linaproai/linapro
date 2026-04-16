// demo_record_update.go defines the request and response DTOs for updating one
// plugin-demo-source record.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateRecordReq is the request for updating one plugin-demo-source record.
type UpdateRecordReq struct {
	g.Meta           `path:"/plugins/plugin-demo-source/records/{id}" method:"put" mime:"multipart/form-data" tags:"源码插件示例" summary:"更新源码插件示例记录" dc:"更新一条 plugin-demo-source 示例记录，并可替换或移除当前附件，用于演示源码插件页面对自有数据表与存储文件的修改操作" permission:"plugin-demo-source:example:update"`
	Id               int64  `json:"id" v:"required|min:1" dc:"记录ID" eg:"1"`
	Title            string `json:"title" v:"required|length:1,128" dc:"记录标题" eg:"源码插件 SQL 示例记录"`
	Content          string `json:"content" dc:"记录内容" eg:"更新后的说明内容"`
	RemoveAttachment int    `json:"removeAttachment" dc:"是否移除当前附件：1=移除 0=保留；当同时上传新文件时会自动替换旧附件" eg:"0"`
}

// UpdateRecordRes is the response for updating one plugin-demo-source record.
type UpdateRecordRes struct {
	Id int64 `json:"id" dc:"更新后的记录ID" eg:"1"`
}
