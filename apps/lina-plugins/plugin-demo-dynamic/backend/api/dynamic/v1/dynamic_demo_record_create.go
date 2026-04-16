// This file defines the demo-record create DTOs for the dynamic plugin sample.

package v1

import "github.com/gogf/gf/v2/frame/g"

// CreateDemoRecordReq is the request for creating one dynamic demo record.
type CreateDemoRecordReq struct {
	g.Meta                  `path:"/demo-records" method:"post" tags:"动态插件示例" summary:"创建动态插件示例记录" dc:"创建一条 plugin-demo-dynamic 示例记录，并可同时上传一个插件自有附件文件，用于演示动态插件页面对安装 SQL 创建的数据表与授权存储文件的写入操作" access:"login" permission:"plugin-demo-dynamic:record:create" operLog:"create"`
	Title                   string `json:"title" v:"required|length:1,128" dc:"记录标题" eg:"动态插件 SQL 示例记录"`
	Content                 string `json:"content" v:"max-length:1000" dc:"记录内容" eg:"该记录由动态插件示例页面创建，用于演示 SQL 数据表的新增操作。"`
	AttachmentName          string `json:"attachmentName" dc:"附件原始文件名；未上传附件时传空字符串" eg:"plugin-demo-dynamic-note.txt"`
	AttachmentContentBase64 string `json:"attachmentContentBase64" dc:"附件内容的 Base64 编码；未上传附件时传空字符串" eg:"SGVsbG8sIHBsdWdpbi1kZW1vLWR5bmFtaWMh"`
	AttachmentContentType   string `json:"attachmentContentType" dc:"附件内容类型；未上传附件时传空字符串" eg:"text/plain"`
}

// CreateDemoRecordRes is the response for creating one dynamic demo record.
type CreateDemoRecordRes struct {
	DemoRecordItem
}
