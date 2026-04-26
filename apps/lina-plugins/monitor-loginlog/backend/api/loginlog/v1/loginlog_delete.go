package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// LoginLog Delete API

// DeleteReq defines the request for deleting login logs.
type DeleteReq struct {
	g.Meta `path:"/loginlog/{ids}" method:"delete" tags:"Login Logs" summary:"Delete login logs" dc:"Delete one or more login log records" permission:"monitor:loginlog:remove"`
	Ids    string `json:"ids" v:"required" dc:"Log IDs, comma-separated" eg:"1,2,3"`
}

// DeleteRes is the login-log delete response.
type DeleteRes struct {
	Deleted int `json:"deleted" dc:"Number of records actually deleted" eg:"3"`
}
