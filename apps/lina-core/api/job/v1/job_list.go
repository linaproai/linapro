package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// ListReq defines the request for querying scheduled jobs.
type ListReq struct {
	g.Meta         `path:"/job" method:"get" tags:"Job Scheduling / Scheduled Jobs" summary:"Get task list" dc:"Paginated query for scheduled job list, supports filtering by group, status, task type, keyword, scheduling range and concurrency strategy" permission:"system:job:list"`
	PageNum        int     `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize       int     `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
	GroupId        *uint64 `json:"groupId" dc:"Filter by task group ID, if not passed, query all" eg:"1"`
	Status         string  `json:"status" dc:"Filter by task status: enabled=enable disabled=disable paused_by_plugin=Plugin processor is not available, if not passed, query all" eg:"enabled"`
	TaskType       string  `json:"taskType" dc:"Filter by task type: handler=Handler task shell=Shell task, if not passed, query all" eg:"handler"`
	Keyword        string  `json:"keyword" dc:"Fuzzy filter by task name or description keywords" eg:"Log cleaning"`
	Scope          string  `json:"scope" dc:"Filter by scheduling scope: master_only=Only the master node executes all_node=All nodes execute, if not passed, query all" eg:"master_only"`
	Concurrency    string  `json:"concurrency" dc:"Filter by concurrency strategy: singleton=single case parallel=parallel, if not passed, query all" eg:"singleton"`
	OrderBy        string  `json:"orderBy" dc:"Sorting fields: id, name, group_id, status, task_type, created_at, updated_at" eg:"updated_at"`
	OrderDirection string  `json:"orderDirection" d:"desc" dc:"Sorting direction: asc=ascending order desc=descending order" eg:"desc"`
}

// ListItem represents one scheduled job row in the list response.
type ListItem struct {
	*entity.SysJob
	GroupCode string `json:"groupCode" dc:"The group code to which it belongs" eg:"default"`
	GroupName string `json:"groupName" dc:"Group name" eg:"Default grouping"`
}

// ListRes defines the response for querying scheduled jobs.
type ListRes struct {
	List  []*ListItem `json:"list" dc:"task list" eg:"[]"`
	Total int         `json:"total" dc:"Total number of items" eg:"1"`
}
