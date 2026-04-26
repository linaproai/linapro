package v1

import "github.com/gogf/gf/v2/frame/g"

// TreeReq returns dept tree for TreeSelect component.
type TreeReq struct {
	g.Meta `path:"/dept/tree" method:"get" tags:"Department Management" summary:"Get department tree" dc:"Get the complete active department tree for management workbench tree selectors, structure projections, or page assembly." permission:"system:dept:query"`
}

// TreeNode represents a node in the Department tree.
type TreeNode struct {
	Id       int         `json:"id" dc:"Department ID" eg:"100"`
	Label    string      `json:"label" dc:"Department name, displayed by the management workbench tree node" eg:"Head office"`
	Children []*TreeNode `json:"children" dc:"List of subdepartments, recursive nested structure" eg:"[]"`
}

// TreeRes defines the response for querying the Department tree.
type TreeRes struct {
	List []*TreeNode `json:"list" dc:"Department tree structure list, the top node is the department with parentId=0" eg:"[]"`
}
