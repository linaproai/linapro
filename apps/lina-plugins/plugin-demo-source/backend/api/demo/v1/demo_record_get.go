// demo_record_get.go defines the request and response DTOs for querying one
// plugin-demo-source record detail.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetRecordReq is the request for querying one plugin-demo-source record detail.
type GetRecordReq struct {
	g.Meta `path:"/plugins/plugin-demo-source/records/{id}" method:"get" tags:"源码插件示例" summary:"查询源码插件示例记录详情" dc:"查询一条 plugin-demo-source 示例记录详情，用于源码插件页面编辑弹窗回填" permission:"plugin-demo-source:example:view"`
	Id     int64 `json:"id" v:"required|min:1" dc:"记录ID" eg:"1"`
}

// GetRecordRes is the response for querying one plugin-demo-source record detail.
type GetRecordRes struct {
	Id             int64  `json:"id" dc:"记录ID" eg:"1"`
	Title          string `json:"title" dc:"记录标题" eg:"源码插件 SQL 示例记录"`
	Content        string `json:"content" dc:"记录内容" eg:"该记录用于演示源码插件页面如何操作安装 SQL 创建的数据表。"`
	AttachmentName string `json:"attachmentName" dc:"附件原始文件名，没有附件时返回空字符串" eg:"plugin-demo-source-note.txt"`
	HasAttachment  int    `json:"hasAttachment" dc:"是否存在附件：1=存在 0=不存在" eg:"1"`
}
