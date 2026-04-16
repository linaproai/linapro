package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleUpdateReq is the request structure for role update.
type RoleUpdateReq struct {
	g.Meta    `path:"/role/{id}" method:"put" summary:"更新角色" tags:"角色管理" dc:"更新角色信息，可修改角色名称、权限字符、状态、菜单关联等，名称和权限字符不能与其他角色重复" permission:"system:role:edit"`
	Id        int    `json:"id" v:"required|min:1" dc:"角色ID" eg:"1"`
	Name      string `json:"name" v:"required|length:2,30" dc:"角色名称，长度2-30字符" eg:"管理员"`
	Key       string `json:"key" v:"required|length:2,30" dc:"权限字符，长度2-30字符" eg:"admin"`
	Sort      int    `json:"sort" v:"min:0" dc:"显示排序" eg:"1"`
	DataScope int    `json:"dataScope" v:"in:1,2,3" dc:"数据权限范围：1=全部 2=本部门 3=仅本人" eg:"1"`
	Status    int    `json:"status" v:"in:0,1" dc:"状态：1=正常 0=停用" eg:"1"`
	Remark    string `json:"remark" v:"length:0,200" dc:"备注" eg:"系统管理员"`
	MenuIds   []int  `json:"menuIds" dc:"关联的菜单ID列表" eg:"[1,2,3,10]"`
}

// RoleUpdateRes is the response structure for role update.
type RoleUpdateRes struct {
	g.Meta `mime:"application/json" example:"{}"`
}
