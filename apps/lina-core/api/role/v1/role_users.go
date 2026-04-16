package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleUsersReq is the request structure for role user list query.
type RoleUsersReq struct {
	g.Meta   `path:"/role/{id}/users" method:"get" summary:"查询角色用户列表" tags:"角色管理" dc:"分页查询已分配指定角色的用户列表，支持按用户名、手机号等条件筛选" permission:"system:role:auth"`
	Id       int    `json:"id" v:"required|min:1" dc:"角色ID" eg:"1"`
	Username string `json:"username" dc:"用户名，模糊查询" eg:"admin"`
	Phone    string `json:"phone" dc:"手机号，模糊查询" eg:"138"`
	Status   int    `json:"status" dc:"状态筛选：1=正常 0=停用，不传则查询全部" eg:"1"`
	Page     int    `json:"page" d:"1" v:"min:1" dc:"页码" eg:"1"`
	Size     int    `json:"size" d:"10" v:"min:1|max:100" dc:"每页记录数" eg:"10"`
}

// RoleUsersRes is the response structure for role user list query.
type RoleUsersRes struct {
	g.Meta `mime:"application/json" example:"{}"`
	List   []*RoleUserItem `json:"list" dc:"用户列表" eg:"[]"`
	Total  int             `json:"total" dc:"总记录数" eg:"10"`
}

// RoleUserItem represents a single user in the role user list.
type RoleUserItem struct {
	Id        int    `json:"id" dc:"用户ID" eg:"1"`
	Username  string `json:"username" dc:"用户名" eg:"admin"`
	Nickname  string `json:"nickname" dc:"昵称" eg:"管理员"`
	Email     string `json:"email" dc:"邮箱" eg:"admin@example.com"`
	Phone     string `json:"phone" dc:"手机号" eg:"13800138000"`
	Status    int    `json:"status" dc:"状态（0=停用 1=正常）" eg:"1"`
	CreatedAt string `json:"createdAt" dc:"创建时间" eg:"2024-01-01 00:00:00"`
}
