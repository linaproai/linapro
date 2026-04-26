// This file defines the user-creation DTOs and validation rules.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// CreateReq defines the request for creating a user.
type CreateReq struct {
	g.Meta   `path:"/user" method:"post" tags:"User Management" summary:"Create user" dc:"Create a new user, the username must be unique in the system. Departments, positions and roles can be specified" permission:"system:user:add"`
	Username string `json:"username" v:"required|length:2,64#validation.user.create.username.required|validation.user.create.username.length" dc:"Username" eg:"zhangsan"`
	Password string `json:"password" v:"required|length:6,32#validation.user.create.password.required|validation.user.create.password.length" dc:"Password" eg:"123456"`
	Nickname string `json:"nickname" v:"required#validation.user.nickname.required" dc:"Nickname" eg:"Zhang San"`
	Email    string `json:"email" dc:"Email" eg:"zhangsan@example.com"`
	Phone    string `json:"phone" dc:"Mobile phone number" eg:"13800138000"`
	Sex      *int   `json:"sex" d:"0" dc:"Gender: 0=Unknown 1=Male 2=Female" eg:"1"`
	Status   *int   `json:"status" d:"1" dc:"Status: 1=normal 0=disabled" eg:"1"`
	Remark   string `json:"remark" dc:"Remarks" eg:"New employees"`
	DeptId   *int   `json:"deptId" dc:"Department ID" eg:"100"`
	PostIds  []int  `json:"postIds" dc:"Position ID list" eg:"[1,2]"`
	RoleIds  []int  `json:"roleIds" dc:"Role ID list" eg:"[1,2]"`
}

// CreateRes defines the response for creating a user.
type CreateRes struct {
	Id int `json:"id" dc:"User ID" eg:"1"`
}
