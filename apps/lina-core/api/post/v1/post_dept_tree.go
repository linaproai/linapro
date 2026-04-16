package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DeptTreeReq defines the request for querying the post department tree.
type DeptTreeReq struct {
	g.Meta `path:"/post/dept-tree" method:"get" tags:"岗位管理" summary:"获取岗位筛选部门树" dc:"获取部门树结构及岗位数量，供管理工作台的岗位查询视图按部门筛选或装配树选择器" permission:"system:post:query"`
}

// DeptTreeRes is the response for department tree
type DeptTreeRes struct {
	List []*DeptTreeNode `json:"list" dc:"部门树" eg:"[]"`
}

// DeptTreeNode represents a node in the department tree
type DeptTreeNode struct {
	Id        int             `json:"id" dc:"部门ID" eg:"100"`
	Label     string          `json:"label" dc:"部门名称" eg:"技术部"`
	PostCount int             `json:"postCount" dc:"该部门下的岗位数量" eg:"5"`
	Children  []*DeptTreeNode `json:"children" dc:"子部门列表" eg:"[]"`
}
