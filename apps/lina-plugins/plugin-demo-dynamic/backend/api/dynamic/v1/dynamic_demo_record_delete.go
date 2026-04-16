// This file defines the demo-record delete DTOs for the dynamic plugin sample.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteDemoRecordReq is the request for deleting one dynamic demo record.
type DeleteDemoRecordReq struct {
	g.Meta `path:"/demo-records/{id}" method:"delete" tags:"动态插件示例" summary:"删除动态插件示例记录" dc:"删除一条 plugin-demo-dynamic 示例记录，并同步清理该记录关联的插件自有附件文件" access:"login" permission:"plugin-demo-dynamic:record:delete" operLog:"delete"`
	Id     string `json:"id" v:"required|length:1,64" dc:"记录唯一标识" eg:"demo-record-1"`
}

// DeleteDemoRecordRes is the response for deleting one dynamic demo record.
type DeleteDemoRecordRes struct {
	Id      string `json:"id" dc:"已删除的记录唯一标识" eg:"demo-record-1"`
	Deleted bool   `json:"deleted" dc:"是否删除成功：true=成功 false=失败" eg:"true"`
}
