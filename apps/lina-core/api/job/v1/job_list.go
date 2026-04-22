package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// ListReq defines the request for querying scheduled jobs.
type ListReq struct {
	g.Meta         `path:"/job" method:"get" tags:"任务调度/定时任务" summary:"获取任务列表" dc:"分页查询定时任务列表，支持按分组、状态、任务类型、关键字、调度范围与并发策略筛选" permission:"system:job:list"`
	PageNum        int     `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize       int     `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
	GroupId        *uint64 `json:"groupId" dc:"按任务分组ID筛选，不传则查询全部" eg:"1"`
	Status         string  `json:"status" dc:"按任务状态筛选：enabled=启用 disabled=停用 paused_by_plugin=插件处理器不可用，不传则查询全部" eg:"enabled"`
	TaskType       string  `json:"taskType" dc:"按任务类型筛选：handler=Handler 任务 shell=Shell 任务，不传则查询全部" eg:"handler"`
	Keyword        string  `json:"keyword" dc:"按任务名称或描述关键字模糊筛选" eg:"日志清理"`
	Scope          string  `json:"scope" dc:"按调度范围筛选：master_only=仅主节点执行 all_node=所有节点执行，不传则查询全部" eg:"master_only"`
	Concurrency    string  `json:"concurrency" dc:"按并发策略筛选：singleton=单例 parallel=并行，不传则查询全部" eg:"singleton"`
	OrderBy        string  `json:"orderBy" dc:"排序字段：id,name,group_id,status,task_type,created_at,updated_at" eg:"updated_at"`
	OrderDirection string  `json:"orderDirection" d:"desc" dc:"排序方向：asc=升序 desc=降序" eg:"desc"`
}

// ListItem represents one scheduled job row in the list response.
type ListItem struct {
	*entity.SysJob
	GroupCode string `json:"groupCode" dc:"所属分组编码" eg:"default"`
	GroupName string `json:"groupName" dc:"所属分组名称" eg:"默认分组"`
}

// ListRes defines the response for querying scheduled jobs.
type ListRes struct {
	List  []*ListItem `json:"list" dc:"任务列表" eg:"[]"`
	Total int         `json:"total" dc:"总条数" eg:"1"`
}
