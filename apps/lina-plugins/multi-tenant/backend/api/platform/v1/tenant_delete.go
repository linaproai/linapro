// This file declares platform tenant delete DTOs for the multi-tenant source plugin.

package v1

import "github.com/gogf/gf/v2/frame/g"

// TenantDeleteReq defines the request for deleting a tenant.
type TenantDeleteReq struct {
	g.Meta `path:"/platform/tenants/{id}" method:"delete" tags:"Platform Tenants" summary:"Delete tenant" dc:"Delete a tenant after lifecycle guard checks pass." permission:"system:tenant:remove"`
	Id     int64 `json:"id" v:"required" dc:"Tenant ID" eg:"1"`
}

// TenantDeleteRes defines the tenant delete response.
type TenantDeleteRes struct{}
