// This file defines the token refresh request and response DTOs for host authentication.

package v1

import "github.com/gogf/gf/v2/frame/g"

// RefreshTokenReq defines the request for refreshing a JWT token.
type RefreshTokenReq struct {
	g.Meta `path:"/auth/refresh" method:"post" tags:"Authentication" summary:"Refresh access token" dc:"Re-issues a new JWT token before the current token expires. The current valid token must be provided in the Authorization header. Returns a new access token with a refreshed expiry."`
}

// RefreshTokenRes defines the response for token refresh.
type RefreshTokenRes struct {
	AccessToken string `json:"accessToken" dc:"New JWT access token with refreshed expiry." eg:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}
