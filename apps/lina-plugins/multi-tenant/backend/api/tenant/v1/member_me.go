// This file declares current tenant member profile DTOs for the multi-tenant source plugin.

package v1

import "github.com/gogf/gf/v2/frame/g"

// MemberMeReq defines the request for the current tenant member profile.
type MemberMeReq struct {
	g.Meta   `path:"/tenant/members/me" method:"get" tags:"Tenant Members" summary:"Get current tenant member profile" dc:"Get the current user's membership profile in a tenant." permission:"system:tenant:member:query"`
	TenantId int64 `json:"tenantId" v:"required" dc:"Current tenant ID supplied by tenant context until host tenantcap is wired" eg:"1"`
	UserId   int64 `json:"userId" dc:"Current user ID supplied by host biz context when omitted" eg:"2"`
}

// MemberMeRes defines the current member profile response.
type MemberMeRes struct {
	*MemberEntity
}
