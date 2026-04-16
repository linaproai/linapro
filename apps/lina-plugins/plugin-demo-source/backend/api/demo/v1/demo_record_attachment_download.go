// demo_record_attachment_download.go defines the request DTO for downloading
// one plugin-demo-source attachment.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DownloadAttachmentReq is the request for downloading one plugin-demo-source attachment.
type DownloadAttachmentReq struct {
	g.Meta `path:"/plugins/plugin-demo-source/records/{id}/attachment" method:"get" tags:"源码插件示例" summary:"下载源码插件示例附件" dc:"下载一条 plugin-demo-source 示例记录当前关联的附件文件，用于演示源码插件页面对插件自有存储文件的读取能力" permission:"plugin-demo-source:example:view"`
	Id     int64 `json:"id" v:"required|min:1" dc:"记录ID" eg:"1"`
}

// DownloadAttachmentRes is the response placeholder for attachment downloads streamed by the controller.
type DownloadAttachmentRes struct{}
