// This file declares platform tenant impersonation end DTOs for the multi-tenant source plugin.

package v1

import "github.com/gogf/gf/v2/frame/g"

// TenantEndImpersonateReq defines the request for ending platform impersonation.
type TenantEndImpersonateReq struct {
	g.Meta `path:"/platform/tenants/{id}/end-impersonate" method:"post" tags:"Platform Tenants" summary:"End tenant impersonation" dc:"Revoke the current impersonation token for a target tenant." permission:"system:tenant:impersonate"`
	Id     int64 `json:"id" v:"required" dc:"Tenant ID" eg:"1"`
}

// TenantEndImpersonateRes defines the impersonation end response.
type TenantEndImpersonateRes struct{}
