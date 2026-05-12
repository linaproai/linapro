// This file declares tenant switching DTOs for the multi-tenant source plugin.

package v1

import "github.com/gogf/gf/v2/frame/g"

// SwitchTenantReq defines the request for switching tenant.
type SwitchTenantReq struct {
	g.Meta   `path:"/auth/switch-tenant" method:"post" tags:"Tenant Auth" summary:"Switch tenant" dc:"Switch from the current tenant to another tenant and request a new tenant-bound token."`
	TenantId int64 `json:"tenantId" v:"required" dc:"Target tenant ID" eg:"1"`
}

// SwitchTenantRes defines the switch tenant response.
type SwitchTenantRes struct {
	AccessToken  string `json:"accessToken" dc:"Tenant-bound token" eg:"eyJhbGciOi..."`
	RefreshToken string `json:"refreshToken" dc:"Tenant-bound refresh token" eg:"eyJhbGciOi..."`
}
