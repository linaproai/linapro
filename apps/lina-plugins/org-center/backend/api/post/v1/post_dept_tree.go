package v1

import "github.com/gogf/gf/v2/frame/g"

// DeptTreeReq defines the request for querying the post Department tree.
type DeptTreeReq struct {
	g.Meta `path:"/post/dept-tree" method:"get" tags:"Position Management" summary:"Get position filter department tree" dc:"Get the department tree and position counts for department filtering or tree selector assembly in the position query view." permission:"system:post:query"`
}

// DeptTreeRes is the response for Department tree.
type DeptTreeRes struct {
	List []*DeptTreeNode `json:"list" dc:"Department tree" eg:"[]"`
}

// DeptTreeNode represents a node in the Department tree.
type DeptTreeNode struct {
	Id        int             `json:"id" dc:"Department ID" eg:"100"`
	Label     string          `json:"label" dc:"Department name" eg:"Technology Department"`
	PostCount int             `json:"postCount" dc:"Number of positions in this department" eg:"5"`
	Children  []*DeptTreeNode `json:"children" dc:"List of subdepartments" eg:"[]"`
}
