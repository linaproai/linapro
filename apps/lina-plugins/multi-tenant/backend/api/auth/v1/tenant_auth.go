// This file declares login tenant candidate DTOs for the multi-tenant source plugin.

package v1

import "github.com/gogf/gf/v2/frame/g"

// LoginTenantEntity is one tenant candidate returned during login.
type LoginTenantEntity struct {
	Id     int64  `json:"id" dc:"Tenant ID" eg:"1"`
	Code   string `json:"code" dc:"Tenant code" eg:"acme"`
	Name   string `json:"name" dc:"Tenant name" eg:"Acme BU"`
	Status string `json:"status" dc:"Tenant status" eg:"active"`
}

// LoginTenantsReq defines the request for login tenant candidates.
type LoginTenantsReq struct {
	g.Meta `path:"/auth/login-tenants" method:"get" tags:"Tenant Auth" summary:"Get login tenant candidates" dc:"Return tenant candidates for a user during the login tenant-selection stage." permission:"system:tenant:auth:login-tenants"`
	UserId int64 `json:"userId" v:"required" dc:"User ID" eg:"2"`
}

// LoginTenantsRes defines the login tenant candidates response.
type LoginTenantsRes struct {
	List []*LoginTenantEntity `json:"list" dc:"Login tenant candidates" eg:"[]"`
}
