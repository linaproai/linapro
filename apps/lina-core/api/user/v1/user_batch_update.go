// This file defines DTOs for the user batch-update API.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// BatchUpdateReq defines the request for updating multiple users.
type BatchUpdateReq struct {
	g.Meta       `path:"/user" method:"put" tags:"User Management" summary:"Batch update users" dc:"Update selected users in one transaction. Status, roles, and tenant memberships are optional patch fields; fields omitted from the request are left unchanged." permission:"system:user:edit"`
	Ids          []int `json:"ids" v:"required|min-length:1" dc:"User ID list" eg:"[2,3]"`
	UpdateStatus bool  `json:"updateStatus" dc:"Whether to update user status" eg:"true"`
	Status       *int  `json:"status" v:"in:0,1#validation.user.status.invalid" dc:"Status: 1=normal 0=disabled" eg:"1"`
	UpdateRoles  bool  `json:"updateRoles" dc:"Whether to replace user role assignments" eg:"true"`
	RoleIds      []int `json:"roleIds" dc:"Role ID list" eg:"[1,2]"`
	UpdateTenant bool  `json:"updateTenant" dc:"Whether to replace user tenant memberships" eg:"true"`
	TenantIds    []int `json:"tenantIds" dc:"Tenant ID list when multi-tenancy is enabled" eg:"[10,20]"`
}

// BatchUpdateRes defines the response for updating multiple users.
type BatchUpdateRes struct{}
