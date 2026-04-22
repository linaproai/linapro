// This file defines the demo-record detail DTOs for the dynamic plugin sample.

package v1

import "github.com/gogf/gf/v2/util/gmeta"

// DemoRecordReq is the request for querying one dynamic demo record detail.
type DemoRecordReq struct {
	gmeta.Meta `path:"/demo-records/{id}" method:"get" tags:"动态插件示例" summary:"查询动态插件示例记录详情" dc:"查询一条 plugin-demo-dynamic 示例记录详情，用于动态插件页面编辑表单回填和附件下载前检查" access:"login" permission:"plugin-demo-dynamic:record:view" operLog:"other"`
	Id         string `json:"id" v:"required|length:1,64" dc:"记录唯一标识" eg:"demo-record-1"`
}

// DemoRecordRes is the response for querying one dynamic demo record detail.
type DemoRecordRes struct {
	DemoRecordItem
}

// DemoRecordItem defines one dynamic plugin demo-record row.
type DemoRecordItem struct {
	Id             string `json:"id" dc:"记录唯一标识" eg:"demo-record-1"`
	Title          string `json:"title" dc:"记录标题" eg:"动态插件 SQL 示例记录"`
	Content        string `json:"content" dc:"记录内容" eg:"该记录用于演示动态插件示例页面对安装 SQL 创建的数据表执行增删查改操作。"`
	AttachmentName string `json:"attachmentName" dc:"附件原始文件名，没有附件时返回空字符串" eg:"plugin-demo-dynamic-note.txt"`
	HasAttachment  bool   `json:"hasAttachment" dc:"是否存在当前附件：true=存在 false=不存在" eg:"true"`
	CreatedAt      string `json:"createdAt" dc:"记录创建时间，由示例数据表的默认时间戳字段自动维护" eg:"2026-04-16 10:00:00"`
	UpdatedAt      string `json:"updatedAt" dc:"记录最近更新时间，由示例数据表的默认时间戳字段自动维护" eg:"2026-04-16 10:05:00"`
}
