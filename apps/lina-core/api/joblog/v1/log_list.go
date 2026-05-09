package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// ListReq defines the request for querying scheduled job logs.
type ListReq struct {
	g.Meta         `path:"/job/log" method:"get" tags:"Job Scheduling / Execution Logs" summary:"Get execution log list" dc:"Query task execution logs with pagination, supporting filtering by task, status, trigger mode, node and time range" permission:"system:joblog:list"`
	PageNum        int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize       int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
	JobId          *int64 `json:"jobId" dc:"Filter by task ID, query all if not passed" eg:"1"`
	Status         string `json:"status" dc:"Filter by log status, query all if not uploaded" eg:"success"`
	Trigger        string `json:"trigger" dc:"Filter by triggering method: cron=scheduled trigger manual=manual trigger, if not passed, query all" eg:"manual"`
	NodeId         string `json:"nodeId" dc:"Filter by execution node ID" eg:"node-a"`
	BeginTime      string `json:"beginTime" dc:"Filter by start time lower bound" eg:"2026-04-19 00:00:00"`
	EndTime        string `json:"endTime" dc:"Filter by start time upper bound" eg:"2026-04-19 23:59:59"`
	OrderBy        string `json:"orderBy" dc:"Sorting fields: id,start_at,end_at,duration_ms,status,created_at" eg:"start_at"`
	OrderDirection string `json:"orderDirection" d:"desc" dc:"Sorting direction: asc=ascending order desc=descending order" eg:"desc"`
}

// ListItem represents one scheduled job log row in the list response.
type ListItem struct {
	*entity.SysJobLog
	JobName string `json:"jobName" dc:"Task name" eg:"Task log cleaning"`
}

// ListRes defines the response for querying scheduled job logs.
type ListRes struct {
	List  []*ListItem `json:"list" dc:"Execution log list" eg:"[]"`
	Total int         `json:"total" dc:"Total number of items" eg:"1"`
}
