package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// ListReq defines the request for querying scheduled job logs.
type ListReq struct {
	g.Meta         `path:"/job/log" method:"get" tags:"任务调度/执行日志" summary:"获取执行日志列表" dc:"分页查询任务执行日志，支持按任务、状态、触发方式、节点与时间范围筛选" permission:"system:joblog:list"`
	PageNum        int     `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize       int     `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
	JobId          *uint64 `json:"jobId" dc:"按任务ID筛选，不传则查询全部" eg:"1"`
	Status         string  `json:"status" dc:"按日志状态筛选，不传则查询全部" eg:"success"`
	Trigger        string  `json:"trigger" dc:"按触发方式筛选：cron=定时触发 manual=手动触发，不传则查询全部" eg:"manual"`
	NodeId         string  `json:"nodeId" dc:"按执行节点标识筛选" eg:"node-a"`
	BeginTime      string  `json:"beginTime" dc:"按开始时间起始筛选" eg:"2026-04-19 00:00:00"`
	EndTime        string  `json:"endTime" dc:"按开始时间结束筛选" eg:"2026-04-19 23:59:59"`
	OrderBy        string  `json:"orderBy" dc:"排序字段：id,start_at,end_at,duration_ms,status,created_at" eg:"start_at"`
	OrderDirection string  `json:"orderDirection" d:"desc" dc:"排序方向：asc=升序 desc=降序" eg:"desc"`
}

// ListItem represents one scheduled job log row in the list response.
type ListItem struct {
	*entity.SysJobLog
	JobName string `json:"jobName" dc:"任务名称" eg:"任务日志清理"`
}

// ListRes defines the response for querying scheduled job logs.
type ListRes struct {
	List  []*ListItem `json:"list" dc:"执行日志列表" eg:"[]"`
	Total int         `json:"total" dc:"总条数" eg:"1"`
}
