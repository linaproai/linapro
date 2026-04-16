package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// LoginLog Delete API

// DeleteReq defines the request for deleting login logs.
type DeleteReq struct {
	g.Meta `path:"/loginlog/{ids}" method:"delete" tags:"登录日志" summary:"删除登录日志" dc:"删除一条或多条登录日志记录" permission:"monitor:loginlog:remove"`
	Ids    string `json:"ids" v:"required" dc:"日志ID，多个用逗号分隔" eg:"1,2,3"`
}

// DeleteRes Login log delete response
type DeleteRes struct {
	Deleted int `json:"deleted" dc:"实际删除的记录数" eg:"3"`
}
