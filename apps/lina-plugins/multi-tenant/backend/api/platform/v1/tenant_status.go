// This file declares platform tenant status DTOs for the multi-tenant source plugin.

package v1

import "github.com/gogf/gf/v2/frame/g"

// TenantStatusReq defines the request for updating tenant lifecycle status.
type TenantStatusReq struct {
	g.Meta `path:"/platform/tenants/{id}/status" method:"put" tags:"Platform Tenants" summary:"Update tenant status" dc:"Suspend or resume a tenant." permission:"system:tenant:edit"`
	Id     int64  `json:"id" v:"required" dc:"Tenant ID" eg:"1"`
	Status string `json:"status" v:"required|in:active,suspended#gf.gvalid.rule.required|gf.gvalid.rule.in" dc:"Tenant lifecycle status" eg:"suspended"`
}

// TenantStatusRes defines the tenant status response.
type TenantStatusRes struct{}
