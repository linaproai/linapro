// This file declares platform tenant create DTOs for the multi-tenant source plugin.

package v1

import "github.com/gogf/gf/v2/frame/g"

// TenantCreateReq defines the request for creating a tenant.
type TenantCreateReq struct {
	g.Meta `path:"/platform/tenants" method:"post" tags:"Platform Tenants" summary:"Create tenant" dc:"Create a tenant with an active lifecycle status." permission:"system:tenant:add"`
	Code   string `json:"code" v:"required|min-length:2|max-length:32|regex:^[a-z0-9](?:[a-z0-9-]*[a-z0-9])$#gf.gvalid.rule.required|Tenant code must have at least 2 characters|Tenant code must have at most 32 characters|Tenant code must use lowercase letters, numbers, and hyphens" dc:"Stable tenant code" eg:"acme"`
	Name   string `json:"name" v:"required#gf.gvalid.rule.required" dc:"Tenant display name" eg:"Acme BU"`
	Remark string `json:"remark" dc:"Tenant remark" eg:"Internal business unit"`
}

// TenantCreateRes defines the tenant create response.
type TenantCreateRes struct {
	Id int64 `json:"id" dc:"Tenant ID" eg:"1"`
}
