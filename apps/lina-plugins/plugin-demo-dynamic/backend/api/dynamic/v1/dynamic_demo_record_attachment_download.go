// This file defines the demo-record attachment download DTOs for the dynamic
// plugin sample.

package v1

import "github.com/gogf/gf/v2/util/gmeta"

// DownloadDemoRecordAttachmentReq is the request for downloading one dynamic demo-record attachment.
type DownloadDemoRecordAttachmentReq struct {
	gmeta.Meta `path:"/demo-records/{id}/attachment" method:"get" tags:"动态插件示例" summary:"下载动态插件示例附件" dc:"下载一条 plugin-demo-dynamic 示例记录当前关联的附件文件，用于演示动态插件页面对插件自有存储文件的读取能力" access:"login" permission:"plugin-demo-dynamic:record:view" operLog:"other"`
	Id         string `json:"id" v:"required|length:1,64" dc:"记录唯一标识" eg:"demo-record-1"`
}

// DownloadDemoRecordAttachmentRes is the response placeholder for streamed attachment downloads.
type DownloadDemoRecordAttachmentRes struct{}
