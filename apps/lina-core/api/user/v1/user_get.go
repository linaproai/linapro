package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// GetReq defines the request for querying user detail.
type GetReq struct {
	g.Meta `path:"/user/{id}" method:"get" tags:"用户管理" summary:"获取用户详情" dc:"根据用户ID获取用户详细信息，包括所属部门和岗位信息" permission:"system:user:query"`
	Id     int `json:"id" v:"required" dc:"用户ID" eg:"1"`
}

// GetRes is the response structure for user detail.
type GetRes struct {
	*entity.SysUser `dc:"用户信息" eg:""`
	DeptId          int    `json:"deptId" dc:"部门ID" eg:"100"`
	DeptName        string `json:"deptName" dc:"部门名称" eg:"技术部"`
	PostIds         []int  `json:"postIds" dc:"岗位ID列表" eg:"[1,2]"`
	RoleIds         []int  `json:"roleIds" dc:"角色ID列表" eg:"[1,2]"`
}
