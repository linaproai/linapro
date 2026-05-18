package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DeptTreeReq defines the request for querying the user department tree.
type DeptTreeReq struct {
	g.Meta `path:"/user/dept-tree" method:"get" tags:"User Management" summary:"Get user filter department tree" dc:"Get the department tree structure and the number of department users for the user query view of the management workbench to filter by department or assemble the tree selector" permission:"system:user:query"`
}

// DeptTreeNode represents a node in the department tree for user filtering.
type DeptTreeNode struct {
	Id        int             `json:"id" dc:"Department ID" eg:"100"`
	Label     string          `json:"label" dc:"Department name" eg:"Technology Department"`
	UserCount int             `json:"userCount" dc:"Number of department users" eg:"5"`
	Children  []*DeptTreeNode `json:"children" dc:"List of subdepartments" eg:"[]"`
}

// DeptTreeRes is the response structure for department tree.
type DeptTreeRes struct {
	List []*DeptTreeNode `json:"list" dc:"department tree" eg:"[]"`
}
