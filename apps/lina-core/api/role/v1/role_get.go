package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleGetReq is the request structure for role detail query.
type RoleGetReq struct {
	g.Meta `path:"/role/{id}" method:"get" summary:"查询角色详情" tags:"角色管理" dc:"根据角色ID查询角色详细信息，包含关联的菜单ID列表" permission:"system:role:query"`
	Id     int `json:"id" v:"required|min:1" dc:"角色ID" eg:"1"`
}

// RoleGetRes is the response structure for role detail query.
type RoleGetRes struct {
	g.Meta    `mime:"application/json" example:"{}"`
	Id        int    `json:"id" dc:"角色ID" eg:"1"`
	Name      string `json:"name" dc:"角色名称" eg:"管理员"`
	Key       string `json:"key" dc:"权限字符" eg:"admin"`
	Sort      int    `json:"sort" dc:"显示排序" eg:"1"`
	DataScope int    `json:"dataScope" dc:"数据权限范围（1=全部 2=本部门 3=仅本人）" eg:"1"`
	Status    int    `json:"status" dc:"状态（0=停用 1=正常）" eg:"1"`
	Remark    string `json:"remark" dc:"备注" eg:"系统管理员角色"`
	MenuIds   []int  `json:"menuIds" dc:"关联的菜单ID列表" eg:"[1,2,3]"`
	CreatedAt string `json:"createdAt" dc:"创建时间" eg:"2024-01-01 00:00:00"`
	UpdatedAt string `json:"updatedAt" dc:"更新时间" eg:"2024-01-01 00:00:00"`
}
