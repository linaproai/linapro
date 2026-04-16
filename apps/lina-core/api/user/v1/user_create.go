package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// CreateReq defines the request for creating a user.
type CreateReq struct {
	g.Meta   `path:"/user" method:"post" tags:"用户管理" summary:"创建用户" dc:"创建一个新用户，用户名在系统中必须唯一。可指定所属部门、岗位和角色" permission:"system:user:add"`
	Username string `json:"username" v:"required|length:2,64#请输入用户名|用户名长度为2-64个字符" dc:"用户名" eg:"zhangsan"`
	Password string `json:"password" v:"required|length:6,32#请输入密码|密码长度为6-32个字符" dc:"密码" eg:"123456"`
	Nickname string `json:"nickname" v:"required#请输入昵称" dc:"昵称" eg:"张三"`
	Email    string `json:"email" dc:"邮箱" eg:"zhangsan@example.com"`
	Phone    string `json:"phone" dc:"手机号" eg:"13800138000"`
	Sex      *int   `json:"sex" d:"0" dc:"性别：0=未知 1=男 2=女" eg:"1"`
	Status   *int   `json:"status" d:"1" dc:"状态：1=正常 0=停用" eg:"1"`
	Remark   string `json:"remark" dc:"备注" eg:"新入职员工"`
	DeptId   *int   `json:"deptId" dc:"部门ID" eg:"100"`
	PostIds  []int  `json:"postIds" dc:"岗位ID列表" eg:"[1,2]"`
	RoleIds  []int  `json:"roleIds" dc:"角色ID列表" eg:"[1,2]"`
}

// CreateRes defines the response for creating a user.
type CreateRes struct {
	Id int `json:"id" dc:"用户ID" eg:"1"`
}
