package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Dept Tree API

// TreeReq returns dept tree for TreeSelect component.
type TreeReq struct {
	g.Meta `path:"/dept/tree" method:"get" tags:"部门管理" summary:"获取部门树" dc:"获取完整的部门树形结构数据，供管理工作台的树选择器、结构投影或页面装配使用，仅包含正常状态的部门" permission:"system:dept:query"`
}

// TreeNode represents a node in the department tree.
type TreeNode struct {
	Id       int         `json:"id" dc:"部门ID" eg:"100"`
	Label    string      `json:"label" dc:"部门名称，供管理工作台树节点显示" eg:"总公司"`
	Children []*TreeNode `json:"children" dc:"子部门列表，递归嵌套结构" eg:"[]"`
}

// TreeRes defines the response for querying the department tree.
type TreeRes struct {
	List []*TreeNode `json:"list" dc:"部门树形结构列表，顶级节点为parentId=0的部门" eg:"[]"`
}
