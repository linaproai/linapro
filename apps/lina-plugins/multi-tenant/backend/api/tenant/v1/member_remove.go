// This file declares tenant-scoped membership remove DTOs for the multi-tenant source plugin.

package v1

import "github.com/gogf/gf/v2/frame/g"

// MemberRemoveReq defines the request for removing a tenant member.
type MemberRemoveReq struct {
	g.Meta `path:"/tenant/members/{id}" method:"delete" tags:"Tenant Members" summary:"Remove tenant member" dc:"Remove one membership from the current tenant." permission:"system:tenant:member:remove"`
	Id     int64 `json:"id" v:"required" dc:"Membership ID" eg:"1"`
}

// MemberRemoveRes defines the member remove response.
type MemberRemoveRes struct{}
