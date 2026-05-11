// This file declares platform tenant list DTOs for the multi-tenant source plugin.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// TenantEntity is the platform tenant API projection.
type TenantEntity struct {
	Id        int64  `json:"id" dc:"Tenant ID" eg:"1"`
	Code      string `json:"code" dc:"Stable tenant code" eg:"acme"`
	Name      string `json:"name" dc:"Tenant display name" eg:"Acme BU"`
	Status    string `json:"status" dc:"Tenant lifecycle status" eg:"active"`
	Remark    string `json:"remark" dc:"Tenant remark" eg:"Internal business unit"`
	CreatedAt string `json:"createdAt" dc:"Tenant creation time" eg:"2026-05-10 09:00:00"`
}

// TenantListReq defines the request for listing tenants.
type TenantListReq struct {
	g.Meta   `path:"/platform/tenants" method:"get" tags:"Platform Tenants" summary:"Get tenant list" dc:"Query tenants by page with optional code, name, and status filters." permission:"system:tenant:list"`
	PageNum  int    `json:"pageNum" d:"1" v:"min:1" dc:"Page number" eg:"1"`
	PageSize int    `json:"pageSize" d:"10" v:"min:1|max:100" dc:"Number of items per page" eg:"10"`
	Code     string `json:"code" dc:"Filter by tenant code" eg:"acme"`
	Name     string `json:"name" dc:"Filter by tenant name" eg:"Acme"`
	Status   string `json:"status" dc:"Filter by tenant lifecycle status" eg:"active"`
}

// TenantListRes defines the tenant list response.
type TenantListRes struct {
	List  []*TenantEntity `json:"list" dc:"Tenant list" eg:"[]"`
	Total int             `json:"total" dc:"Total tenant count" eg:"12"`
}
