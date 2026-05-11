// This file declares platform tenant update DTOs for the multi-tenant source plugin.

package v1

import "github.com/gogf/gf/v2/frame/g"

// TenantUpdateReq defines the request for updating a tenant.
type TenantUpdateReq struct {
	g.Meta `path:"/platform/tenants/{id}" method:"put" tags:"Platform Tenants" summary:"Update tenant" dc:"Update tenant profile fields." permission:"system:tenant:edit"`
	Id     int64   `json:"id" v:"required" dc:"Tenant ID" eg:"1"`
	Name   *string `json:"name" dc:"Tenant display name" eg:"Acme BU"`
	Remark *string `json:"remark" dc:"Tenant remark" eg:"Internal business unit"`
}

// TenantUpdateRes defines the tenant update response.
type TenantUpdateRes struct{}
