package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// Dept Exclude API

// ExcludeReq returns dept list excluding a node and its children.
type ExcludeReq struct {
	g.Meta `path:"/dept/exclude/{id}" method:"get" tags:"部门管理" summary:"获取排除节点后的部门列表" dc:"获取排除指定部门及其所有子部门后的部门列表，供管理工作台在选择父级部门时构建合法候选集，防止形成循环引用" permission:"system:dept:query"`
	Id     int `json:"id" v:"required" dc:"需排除的部门ID，该部门及其所有下级部门将从结果中过滤掉" eg:"100"`
}

// ExcludeRes defines the response for querying departments with exclusions.
type ExcludeRes struct {
	List []*entity.SysDept `json:"list" dc:"排除指定节点及其子节点后的部门列表" eg:"[]"`
}
