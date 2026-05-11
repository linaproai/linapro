// This file declares tenant-scoped membership add DTOs for the multi-tenant source plugin.

package v1

import "github.com/gogf/gf/v2/frame/g"

// MemberAddReq defines the request for adding a tenant member.
type MemberAddReq struct {
	g.Meta   `path:"/tenant/members" method:"post" tags:"Tenant Members" summary:"Add tenant member" dc:"Add a user into the current tenant." permission:"system:tenant:member:add"`
	TenantId int64 `json:"tenantId" v:"required" dc:"Current tenant ID supplied by tenant context until host tenantcap is wired" eg:"1"`
	UserId   int64 `json:"userId" v:"required" dc:"User ID" eg:"2"`
}

// MemberAddRes defines the member add response.
type MemberAddRes struct {
	Id int64 `json:"id" dc:"Membership ID" eg:"1"`
}
