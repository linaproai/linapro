package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// LoginLog Clean API

// CleanReq defines the request for clearing login logs.
type CleanReq struct {
	g.Meta    `path:"/loginlog/clean" method:"delete" tags:"Login Logs" summary:"Clear login logs" dc:"Clear login logs within a specified time range, or clear all logs when no range is provided." permission:"monitor:loginlog:clear"`
	BeginTime string `json:"beginTime" dc:"Cleanup start time" eg:"2025-01-01"`
	EndTime   string `json:"endTime" dc:"Cleanup end time" eg:"2025-06-30"`
}

// CleanRes is the login-log clean response.
type CleanRes struct {
	Deleted int `json:"deleted" dc:"Number of records actually deleted" eg:"500"`
}
