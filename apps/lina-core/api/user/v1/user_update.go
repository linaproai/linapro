// This file defines the user-update DTOs and validation rules.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UpdateReq defines the request for updating a user.
type UpdateReq struct {
	g.Meta   `path:"/user/{id}" method:"put" tags:"User Management" summary:"Update user" dc:"Update the information of the specified user. All fields are optional to update. Only the fields that need to be modified are passed in." permission:"system:user:edit"`
	Id       int     `json:"id" v:"required" dc:"User ID" eg:"1"`
	Username *string `json:"username" dc:"Username" eg:"zhangsan"`
	Password *string `json:"password" dc:"Password (if empty, do not change)" eg:"newpass123"`
	Nickname *string `json:"nickname" v:"required#validation.user.nickname.required" dc:"Nickname" eg:"Zhang San"`
	Email    *string `json:"email" dc:"Email" eg:"zhangsan@example.com"`
	Phone    *string `json:"phone" dc:"Mobile phone number" eg:"13800138000"`
	Sex      *int    `json:"sex" dc:"Gender: 0=Unknown 1=Male 2=Female" eg:"1"`
	Status   *int    `json:"status" dc:"Status: 1=normal 0=disabled" eg:"1"`
	Remark   *string `json:"remark" dc:"Remarks" eg:"Update notes information"`
	DeptId   *int    `json:"deptId" dc:"Department ID" eg:"100"`
	PostIds  []int   `json:"postIds" dc:"Position ID list" eg:"[1,2]"`
	RoleIds  []int   `json:"roleIds" dc:"Role ID list" eg:"[1,2]"`
}

// UpdateRes defines the response for updating a user.
type UpdateRes struct{}
