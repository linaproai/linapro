package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// ListReq defines the request for querying scheduled job groups.
type ListReq struct {
	g.Meta         `path:"/job-group" method:"get" tags:"任务分组管理" summary:"获取分组列表" dc:"分页查询任务分组列表，支持按编码与名称关键字筛选" permission:"system:jobgroup:list"`
	PageNum        int    `json:"pageNum" d:"1" v:"min:1" dc:"页码" eg:"1"`
	PageSize       int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"每页条数" eg:"10"`
	Code           string `json:"code" dc:"按分组编码筛选（模糊匹配）" eg:"default"`
	Name           string `json:"name" dc:"按分组名称筛选（模糊匹配）" eg:"默认分组"`
	OrderBy        string `json:"orderBy" dc:"排序字段：id,sort_order,code,name,created_at,updated_at" eg:"sort_order"`
	OrderDirection string `json:"orderDirection" d:"asc" dc:"排序方向：asc=升序 desc=降序" eg:"asc"`
}

// ListItem represents one scheduled job group row in the list response.
type ListItem struct {
	*entity.SysJobGroup
	JobCount int64 `json:"jobCount" dc:"当前分组下的任务数量" eg:"3"`
}

// ListRes defines the response for querying scheduled job groups.
type ListRes struct {
	List  []*ListItem `json:"list" dc:"分组列表" eg:"[]"`
	Total int         `json:"total" dc:"总条数" eg:"1"`
}
