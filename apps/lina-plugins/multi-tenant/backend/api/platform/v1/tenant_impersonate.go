// This file declares platform tenant impersonation start DTOs for the multi-tenant source plugin.

package v1

import "github.com/gogf/gf/v2/frame/g"

// TenantImpersonateReq defines the request for starting platform impersonation.
type TenantImpersonateReq struct {
	g.Meta `path:"/platform/tenants/{id}/impersonate" method:"post" tags:"Platform Tenants" summary:"Start tenant impersonation" dc:"Start platform administrator impersonation for a target tenant." permission:"system:tenant:impersonate"`
	Id     int64  `json:"id" v:"required" dc:"Tenant ID" eg:"1"`
	Reason string `json:"reason" dc:"Audit reason for impersonation" eg:"Support investigation"`
}

// TenantImpersonateRes defines the impersonation response.
type TenantImpersonateRes struct {
	Token          string `json:"token" dc:"Signed impersonation token" eg:"eyJhbGciOi..."`
	TenantId       int64  `json:"tenantId" dc:"Target tenant ID" eg:"1"`
	ActingUserId   int64  `json:"actingUserId" dc:"Platform administrator user ID" eg:"1"`
	IsImpersonated bool   `json:"isImpersonated" dc:"Whether the token is impersonated" eg:"true"`
}
