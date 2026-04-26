package v1

import "github.com/gogf/gf/v2/frame/g"

// OptionSelectReq defines the request for querying post options.
type OptionSelectReq struct {
	g.Meta `path:"/post/option-select" method:"get" tags:"Position Management" summary:"Get position options under department" dc:"Get position options under the specified department and its sub-departments for user create/edit forms." permission:"system:post:query"`
	DeptId *int `json:"deptId" dc:"Department ID, when omitted, all positions will be returned" eg:"100"`
}

// PostOption represents a post option for selection.
type PostOption struct {
	PostId   int    `json:"postId" dc:"Position ID" eg:"1"`
	PostName string `json:"postName" dc:"Position name" eg:"Development Engineer"`
}

// OptionSelectRes is the response for post option selection.
type OptionSelectRes struct {
	List []*PostOption `json:"list" dc:"Job options list" eg:"[]"`
}
