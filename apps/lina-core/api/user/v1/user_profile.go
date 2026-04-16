package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// GetProfileReq defines the request for querying the current user profile.
type GetProfileReq struct {
	g.Meta `path:"/user/profile" method:"get" tags:"用户管理" summary:"获取当前用户信息" dc:"获取当前登录用户的完整个人信息，供个人中心或管理工作台资料视图展示"`
}

// GetProfileRes defines the response for querying the current user profile.
type GetProfileRes struct {
	*entity.SysUser `dc:"用户信息"`
}

// UpdateProfileReq defines the request for updating the current user profile.
type UpdateProfileReq struct {
	g.Meta   `path:"/user/profile" method:"put" tags:"用户管理" summary:"更新当前用户信息" dc:"更新当前登录用户的个人资料，包括昵称、邮箱、手机号、性别等，供个人中心或管理工作台资料维护视图使用"`
	Nickname *string `json:"nickname" v:"required#请输入昵称" dc:"昵称" eg:"管理员"`
	Email    *string `json:"email" dc:"邮箱" eg:"admin@example.com"`
	Phone    *string `json:"phone" dc:"手机号" eg:"13800138000"`
	Sex      *int    `json:"sex" dc:"性别：0=未知 1=男 2=女" eg:"1"`
	Password *string `json:"password" dc:"新密码" eg:"newpass123"`
}

// UpdateProfileRes defines the response for updating the current user profile.
type UpdateProfileRes struct{}
