package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// OperLog Get API

// GetReq defines the request for retrieving operation log details.
type GetReq struct {
	g.Meta `path:"/operlog/{id}" method:"get" tags:"Operation Logs" summary:"Get operation log details" dc:"Get operation log details by ID, including request parameters, response data, and elapsed time." permission:"monitor:operlog:query"`
	Id     int `json:"id" v:"required" dc:"Operation log ID" eg:"1"`
}

// GetRes is the operation-log detail response.
type GetRes struct {
	*OperLogEntity
}
