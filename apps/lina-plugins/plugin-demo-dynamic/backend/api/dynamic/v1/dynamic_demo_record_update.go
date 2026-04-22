// This file defines the demo-record update DTOs for the dynamic plugin sample.

package v1

import "github.com/gogf/gf/v2/util/gmeta"

// UpdateDemoRecordReq is the request for updating one dynamic demo record.
type UpdateDemoRecordReq struct {
	gmeta.Meta              `path:"/demo-records/{id}" method:"put" tags:"动态插件示例" summary:"更新动态插件示例记录" dc:"更新一条 plugin-demo-dynamic 示例记录，并可替换或移除当前附件，用于演示动态插件页面对插件自有数据表与授权存储文件的修改操作" access:"login" permission:"plugin-demo-dynamic:record:update" operLog:"update"`
	Id                      string `json:"id" v:"required|length:1,64" dc:"记录唯一标识" eg:"demo-record-1"`
	Title                   string `json:"title" v:"required|length:1,128" dc:"记录标题" eg:"动态插件 SQL 示例记录"`
	Content                 string `json:"content" v:"max-length:1000" dc:"记录内容" eg:"更新后的动态插件示例记录内容。"`
	AttachmentName          string `json:"attachmentName" dc:"新上传附件的原始文件名；未上传新附件时传空字符串" eg:"plugin-demo-dynamic-note.txt"`
	AttachmentContentBase64 string `json:"attachmentContentBase64" dc:"新上传附件内容的 Base64 编码；未上传新附件时传空字符串" eg:"SGVsbG8sIHVwZGF0ZWQgZHluYW1pYyBwbHVnaW4h"`
	AttachmentContentType   string `json:"attachmentContentType" dc:"新上传附件内容类型；未上传新附件时传空字符串" eg:"text/plain"`
	RemoveAttachment        bool   `json:"removeAttachment" dc:"是否移除当前附件：true=移除 false=保留；若同时上传新附件，则以新附件替换旧附件" eg:"false"`
}

// UpdateDemoRecordRes is the response for updating one dynamic demo record.
type UpdateDemoRecordRes struct {
	DemoRecordItem
}
