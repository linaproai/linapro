// This file declares tenant-scoped membership list DTOs for the multi-tenant source plugin.

package v1

import "github.com/gogf/gf/v2/frame/g"

// MemberEntity is the tenant-scoped membership API projection.
type MemberEntity struct {
	Id       int64  `json:"id" dc:"Membership ID" eg:"1"`
	UserId   int64  `json:"userId" dc:"User ID" eg:"2"`
	TenantId int64  `json:"tenantId" dc:"Tenant ID" eg:"1"`
	Username string `json:"username" dc:"Username" eg:"alice"`
	Nickname string `json:"nickname" dc:"User nickname" eg:"Alice"`
	Status   int    `json:"status" dc:"Membership status: 0=disabled 1=enabled" eg:"1"`
}

// MemberListReq defines the request for listing tenant members.
type MemberListReq struct {
	g.Meta   `path:"/tenant/members" method:"get" tags:"Tenant Members" summary:"Get tenant member list" dc:"Query current tenant memberships by page." permission:"system:tenant:member:list"`
	PageNum  int   `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize int   `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
	TenantId int64 `json:"tenantId" dc:"Current tenant ID supplied by tenant context until host tenantcap is wired" eg:"1"`
	UserId   int64 `json:"userId" dc:"User ID filter" eg:"2"`
	Status   int   `json:"status" d:"-1" dc:"Membership status filter: -1=all 0=disabled 1=enabled" eg:"1"`
}

// MemberListRes defines the member list response.
type MemberListRes struct {
	List  []*MemberEntity `json:"list" dc:"Tenant member list" eg:"[]"`
	Total int             `json:"total" dc:"Total member count" eg:"20"`
}
