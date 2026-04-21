// This file declares the request/response DTOs for querying user-management
// post options through the host-owned organization capability seam.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// PostOptionsReq defines the request for querying user post options.
type PostOptionsReq struct {
	g.Meta `path:"/user/post-options" method:"get" tags:"用户管理" summary:"获取用户岗位选项" dc:"获取指定部门及其子部门下的岗位选项，供用户管理编辑抽屉在组织插件可用时装配岗位多选框" permission:"system:user:query"`
	DeptId *int `json:"deptId" dc:"部门ID，不传则返回全部可选岗位；组织插件不可用时返回空列表" eg:"100"`
}

// UserPostOption represents one selectable post option for user editing.
type UserPostOption struct {
	PostId   int    `json:"postId" dc:"岗位ID" eg:"1"`
	PostName string `json:"postName" dc:"岗位名称" eg:"开发工程师"`
}

// PostOptionsRes is the response structure for user post options.
type PostOptionsRes struct {
	List []*UserPostOption `json:"list" dc:"岗位选项列表" eg:"[]"`
}
