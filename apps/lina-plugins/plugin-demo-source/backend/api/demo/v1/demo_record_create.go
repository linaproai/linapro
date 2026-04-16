// demo_record_create.go defines the request and response DTOs for creating one
// plugin-demo-source record.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CreateRecordReq is the request for creating one plugin-demo-source record.
type CreateRecordReq struct {
	g.Meta  `path:"/plugins/plugin-demo-source/records" method:"post" mime:"multipart/form-data" tags:"源码插件示例" summary:"创建源码插件示例记录" dc:"创建一条 plugin-demo-source 示例记录，并可同时上传一个插件自有附件文件，用于演示源码插件页面对安装 SQL 创建的数据表与存储文件的写入操作" permission:"plugin-demo-source:example:create"`
	Title   string `json:"title" v:"required|length:1,128" dc:"记录标题" eg:"源码插件 SQL 示例记录"`
	Content string `json:"content" dc:"记录内容" eg:"该记录用于演示源码插件页面如何操作安装 SQL 创建的数据表。"`
}

// CreateRecordRes is the response for creating one plugin-demo-source record.
type CreateRecordRes struct {
	Id int64 `json:"id" dc:"新建后的记录ID" eg:"1"`
}
