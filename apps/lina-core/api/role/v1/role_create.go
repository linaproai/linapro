package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// RoleCreateReq is the request structure for role creation.
type RoleCreateReq struct {
	g.Meta    `path:"/role" method:"post" summary:"创建角色" tags:"角色管理" dc:"创建新角色，角色名称和权限字符必须唯一，可关联菜单" permission:"system:role:add"`
	Name      string `json:"name" v:"required|length:2,30" dc:"角色名称，长度2-30字符" eg:"测试角色"`
	Key       string `json:"key" v:"required|length:2,30" dc:"权限字符，长度2-30字符，用于权限标识" eg:"test_role"`
	Sort      int    `json:"sort" d:"0" v:"min:0" dc:"显示排序，数字越小越靠前" eg:"0"`
	DataScope int    `json:"dataScope" d:"1" v:"in:1,2,3" dc:"数据权限范围：1=全部 2=本部门 3=仅本人" eg:"1"`
	Status    int    `json:"status" d:"1" v:"in:0,1" dc:"状态：1=正常 0=停用" eg:"1"`
	Remark    string `json:"remark" v:"length:0,200" dc:"备注，最多200字符" eg:"测试角色描述"`
	MenuIds   []int  `json:"menuIds" dc:"关联的菜单ID列表，用于控制角色的菜单权限" eg:"[1,2,3]"`
}

// RoleCreateRes is the response structure for role creation.
type RoleCreateRes struct {
	g.Meta `mime:"application/json" example:"{}"`
	Id     int `json:"id" dc:"创建成功后的角色ID" eg:"3"`
}
