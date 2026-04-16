package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// OptionSelectReq defines the request for querying post options.
type OptionSelectReq struct {
	g.Meta `path:"/post/option-select" method:"get" tags:"岗位管理" summary:"获取部门下岗位选项" dc:"获取指定部门及其子部门下的岗位选项列表，用于用户创建/编辑时选择岗位" permission:"system:post:query"`
	DeptId *int `json:"deptId" dc:"部门ID，不传则返回所有岗位" eg:"100"`
}

// PostOption represents a post option for selection
type PostOption struct {
	PostId   int    `json:"postId" dc:"岗位ID" eg:"1"`
	PostName string `json:"postName" dc:"岗位名称" eg:"开发工程师"`
}

// OptionSelectRes is the response for post option selection
type OptionSelectRes struct {
	List []*PostOption `json:"list" dc:"岗位选项列表" eg:"[]"`
}
