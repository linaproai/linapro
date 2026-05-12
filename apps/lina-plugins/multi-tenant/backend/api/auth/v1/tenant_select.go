// This file declares tenant selection DTOs for the multi-tenant source plugin.

package v1

import "github.com/gogf/gf/v2/frame/g"

// SelectTenantReq defines the request for selecting a tenant after login.
type SelectTenantReq struct {
	g.Meta   `path:"/auth/select-tenant" method:"post" tags:"Tenant Auth" summary:"Select tenant" dc:"Select one tenant after pre-login and request a tenant-bound token."`
	PreToken string `json:"preToken" v:"required#gf.gvalid.rule.required" dc:"Short-lived pre-login token" eg:"pre_eyJhbGciOi..."`
	TenantId int64  `json:"tenantId" v:"required" dc:"Target tenant ID" eg:"1"`
}

// SelectTenantRes defines the selected tenant token response.
type SelectTenantRes struct {
	AccessToken  string `json:"accessToken" dc:"Tenant-bound token" eg:"eyJhbGciOi..."`
	RefreshToken string `json:"refreshToken" dc:"Tenant-bound refresh token" eg:"eyJhbGciOi..."`
}
