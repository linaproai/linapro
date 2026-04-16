package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleListReq is the request structure for role list query.
type RoleListReq struct {
	g.Meta `path:"/role" method:"get" summary:"查询角色列表" tags:"角色管理" dc:"分页查询角色列表，支持按角色名称、权限字符、状态等条件筛选" permission:"system:role:query"`
	Name   string `json:"name" dc:"角色名称，模糊查询" eg:"管理员"`
	Key    string `json:"key" dc:"权限字符，模糊查询" eg:"admin"`
	Status int    `json:"status" dc:"状态筛选：1=正常 0=停用，不传则查询全部" eg:"1"`
	Page   int    `json:"page" d:"1" v:"min:1" dc:"页码" eg:"1"`
	Size   int    `json:"size" d:"10" v:"min:1|max:100" dc:"每页记录数" eg:"10"`
}

// RoleListRes is the response structure for role list query.
type RoleListRes struct {
	g.Meta `mime:"application/json" example:"{}"`
	List   []*RoleListItem `json:"list" dc:"角色列表" eg:"[]"`
	Total  int             `json:"total" dc:"总记录数" eg:"10"`
}

// RoleListItem represents a single role in the list.
type RoleListItem struct {
	Id        int    `json:"id" dc:"角色ID" eg:"1"`
	Name      string `json:"name" dc:"角色名称" eg:"管理员"`
	Key       string `json:"key" dc:"权限字符" eg:"admin"`
	Sort      int    `json:"sort" dc:"显示排序" eg:"1"`
	DataScope int    `json:"dataScope" dc:"数据权限范围（1=全部 2=本部门 3=仅本人）" eg:"1"`
	Status    int    `json:"status" dc:"状态（0=停用 1=正常）" eg:"1"`
	Remark    string `json:"remark" dc:"备注" eg:"系统管理员角色"`
	CreatedAt string `json:"createdAt" dc:"创建时间" eg:"2024-01-01 00:00:00"`
	UpdatedAt string `json:"updatedAt" dc:"更新时间" eg:"2024-01-01 00:00:00"`
}
