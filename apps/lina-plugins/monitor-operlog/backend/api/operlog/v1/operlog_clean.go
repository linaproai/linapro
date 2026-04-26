package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// OperLog Clean API

// CleanReq defines the request for clearing operation logs.
type CleanReq struct {
	g.Meta    `path:"/operlog/clean" method:"delete" tags:"Operation Logs" summary:"Clear operation logs" dc:"Clear operation logs within a specified time range, or clear all operation logs when no range is provided." permission:"monitor:operlog:clear"`
	BeginTime string `json:"beginTime" dc:"Cleanup start time" eg:"2025-01-01"`
	EndTime   string `json:"endTime" dc:"Cleanup end time" eg:"2025-06-30"`
}

// CleanRes is the operation-log clean response.
type CleanRes struct {
	Deleted int `json:"deleted" dc:"Number of records actually deleted" eg:"1000"`
}
