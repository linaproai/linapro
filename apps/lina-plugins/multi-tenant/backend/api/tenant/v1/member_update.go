// This file declares tenant-scoped membership update DTOs for the multi-tenant source plugin.

package v1

import "github.com/gogf/gf/v2/frame/g"

// MemberUpdateReq defines the request for updating a tenant member.
type MemberUpdateReq struct {
	g.Meta `path:"/tenant/members/{id}" method:"put" tags:"Tenant Members" summary:"Update tenant member" dc:"Update membership status." permission:"system:tenant:member:edit"`
	Id     int64 `json:"id" v:"required" dc:"Membership ID" eg:"1"`
	Status *int  `json:"status" dc:"Membership status: 0=disabled 1=enabled" eg:"1"`
}

// MemberUpdateRes defines the member update response.
type MemberUpdateRes struct{}
