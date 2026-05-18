// This file declares the request/response DTOs for querying user-management
// post options through the host-owned organization capability seam.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// PostOptionsReq defines the request for querying user post options.
type PostOptionsReq struct {
	g.Meta `path:"/user/post-options" method:"get" tags:"User Management" summary:"Get user position options" dc:"Get the position options under the specified department and its sub-departments for users to manage the edit drawer. Assemble the position check box when the organization plugin is available." permission:"system:user:query"`
	DeptId *int `json:"deptId" dc:"Department ID, if not passed, all available positions will be returned; if the organization plugin is unavailable, an empty list will be returned" eg:"100"`
}

// UserPostOption represents one selectable post option for user editing.
type UserPostOption struct {
	PostId   int    `json:"postId" dc:"Position ID" eg:"1"`
	PostName string `json:"postName" dc:"Job title" eg:"Development Engineer"`
}

// PostOptionsRes is the response structure for user post options.
type PostOptionsRes struct {
	List []*UserPostOption `json:"list" dc:"Job options list" eg:"[]"`
}
