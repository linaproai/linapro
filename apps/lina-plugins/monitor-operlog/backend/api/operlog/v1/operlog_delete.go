package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// OperLog Delete API

// DeleteReq defines the request for deleting operation logs.
type DeleteReq struct {
	g.Meta `path:"/operlog/{ids}" method:"delete" tags:"Operation Logs" summary:"Delete operation log" dc:"Delete one or more operation log records" permission:"monitor:operlog:remove"`
	Ids    string `json:"ids" v:"required" dc:"Log IDs, comma-separated" eg:"1,2,3"`
}

// DeleteRes is the operation-log delete response.
type DeleteRes struct {
	Deleted int `json:"deleted" dc:"Number of records actually deleted" eg:"3"`
}
