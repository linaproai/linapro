package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// LoginLog Get API

// GetReq defines the request for retrieving login log details.
type GetReq struct {
	g.Meta `path:"/loginlog/{id}" method:"get" tags:"Login Logs" summary:"Get login log details" dc:"Get login log details by ID, including login IP, browser, and operating system." permission:"monitor:loginlog:query"`
	Id     int `json:"id" v:"required" dc:"Login log ID" eg:"1"`
}

// GetRes is the login-log detail response.
type GetRes struct {
	*LoginLogEntity
}
