package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// ListReq defines the request for querying the department list.
type ListReq struct {
	g.Meta `path:"/dept" method:"get" tags:"部门管理" summary:"获取部门列表" dc:"获取部门列表数据，支持按部门名称和状态进行筛选，返回的列表按排序号升序排列，包含所有层级的部门信息" permission:"system:dept:query"`
	Name   string `json:"name" dc:"按部门名称筛选，支持模糊匹配" eg:"技术"`
	Status *int   `json:"status" dc:"按状态筛选：1=正常，0=停用，不传则查询全部" eg:"1"`
}

// ListRes defines the response for querying the department list.
type ListRes struct {
	List []*entity.SysDept `json:"list" dc:"部门列表数据，包含所有匹配条件的部门记录" eg:"[]"`
}
