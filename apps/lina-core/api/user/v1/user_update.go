package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UpdateReq defines the request for updating a user.
type UpdateReq struct {
	g.Meta   `path:"/user/{id}" method:"put" tags:"用户管理" summary:"更新用户" dc:"更新指定用户的信息，所有字段均为可选更新，仅传入需要修改的字段即可" permission:"system:user:edit"`
	Id       int     `json:"id" v:"required" dc:"用户ID" eg:"1"`
	Username *string `json:"username" dc:"用户名" eg:"zhangsan"`
	Password *string `json:"password" dc:"密码（为空则不修改）" eg:"newpass123"`
	Nickname *string `json:"nickname" v:"required#请输入昵称" dc:"昵称" eg:"张三"`
	Email    *string `json:"email" dc:"邮箱" eg:"zhangsan@example.com"`
	Phone    *string `json:"phone" dc:"手机号" eg:"13800138000"`
	Sex      *int    `json:"sex" dc:"性别：0=未知 1=男 2=女" eg:"1"`
	Status   *int    `json:"status" dc:"状态：1=正常 0=停用" eg:"1"`
	Remark   *string `json:"remark" dc:"备注" eg:"更新备注信息"`
	DeptId   *int    `json:"deptId" dc:"部门ID" eg:"100"`
	PostIds  []int   `json:"postIds" dc:"岗位ID列表" eg:"[1,2]"`
	RoleIds  []int   `json:"roleIds" dc:"角色ID列表" eg:"[1,2]"`
}

// UpdateRes defines the response for updating a user.
type UpdateRes struct{}
