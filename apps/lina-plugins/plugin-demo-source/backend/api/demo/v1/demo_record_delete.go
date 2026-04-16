// demo_record_delete.go defines the request and response DTOs for deleting one
// plugin-demo-source record.

package v1

import "github.com/gogf/gf/v2/frame/g"

// DeleteRecordReq is the request for deleting one plugin-demo-source record.
type DeleteRecordReq struct {
	g.Meta `path:"/plugins/plugin-demo-source/records/{id}" method:"delete" tags:"源码插件示例" summary:"删除源码插件示例记录" dc:"删除一条 plugin-demo-source 示例记录，并同步清理该记录关联的插件自有附件文件" permission:"plugin-demo-source:example:delete"`
	Id     int64 `json:"id" v:"required|min:1" dc:"记录ID" eg:"1"`
}

// DeleteRecordRes is the response for deleting one plugin-demo-source record.
type DeleteRecordRes struct {
	Id int64 `json:"id" dc:"已删除的记录ID" eg:"1"`
}
