package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// OperLog Delete API

// DeleteReq defines the request for deleting operation logs.
type DeleteReq struct {
	g.Meta `path:"/operlog/{ids}" method:"delete" tags:"操作日志" summary:"删除操作日志" dc:"删除一条或多条操作日志记录" permission:"monitor:operlog:remove"`
	Ids    string `json:"ids" v:"required" dc:"日志ID，多个用逗号分隔" eg:"1,2,3"`
}

// DeleteRes Operation log delete response
type DeleteRes struct {
	Deleted int `json:"deleted" dc:"实际删除的记录数" eg:"3"`
}
