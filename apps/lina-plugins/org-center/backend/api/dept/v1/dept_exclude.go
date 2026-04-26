package v1

import "github.com/gogf/gf/v2/frame/g"

// ExcludeReq returns dept list excluding a node and its children.
type ExcludeReq struct {
	g.Meta `path:"/dept/exclude/{id}" method:"get" tags:"Department Management" summary:"Get the list of departments after excluding nodes" dc:"Get the department list after excluding the specified department and its descendants, so parent department selection can avoid circular references." permission:"system:dept:query"`
	Id     int `json:"id" v:"required" dc:"The department ID to be excluded. This department and all its subordinate departments will be filtered out from the results." eg:"100"`
}

// ExcludeRes defines the response for querying departments with exclusions.
type ExcludeRes struct {
	List []*DeptEntity `json:"list" dc:"List of departments excluding the specified node and its sub-nodes" eg:"[]"`
}
