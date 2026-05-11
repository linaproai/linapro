package v1

import (
	"lina-core/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// GetReq defines the request for querying user detail.
type GetReq struct {
	g.Meta `path:"/user/{id}" method:"get" tags:"User Management" summary:"Get user details" dc:"Obtain user details based on user ID, including department and position information" permission:"system:user:query"`
	Id     int `json:"id" v:"required" dc:"User ID" eg:"1"`
}

// GetRes is the response structure for user detail.
type GetRes struct {
	*entity.SysUser `dc:"User information" eg:""`
	DeptId          int      `json:"deptId" dc:"Department ID" eg:"100"`
	DeptName        string   `json:"deptName" dc:"Department name" eg:"Technology Department"`
	PostIds         []int    `json:"postIds" dc:"Position ID list" eg:"[1,2]"`
	RoleIds         []int    `json:"roleIds" dc:"Role ID list" eg:"[1,2]"`
	TenantIds       []int    `json:"tenantIds" dc:"Tenant ID list when multi-tenancy is enabled" eg:"[10,20]"`
	TenantNames     []string `json:"tenantNames" dc:"Tenant name list when multi-tenancy is enabled" eg:"[\"Alpha Tenant\",\"Beta Tenant\"]"`
}
