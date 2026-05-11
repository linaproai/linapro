// This file declares platform tenant detail DTOs for the multi-tenant source plugin.

package v1

import "github.com/gogf/gf/v2/frame/g"

// TenantGetReq defines the request for retrieving tenant details.
type TenantGetReq struct {
	g.Meta `path:"/platform/tenants/{id}" method:"get" tags:"Platform Tenants" summary:"Get tenant details" dc:"Get one tenant by ID." permission:"system:tenant:query"`
	Id     int64 `json:"id" v:"required" dc:"Tenant ID" eg:"1"`
}

// TenantGetRes defines the tenant detail response.
type TenantGetRes struct {
	*TenantEntity
}
