// This file defines the refresh-token request and response DTOs for host authentication.

package v1

import "github.com/gogf/gf/v2/frame/g"

// RefreshReq defines the request for refreshing an access token.
type RefreshReq struct {
	g.Meta       `path:"/auth/refresh" method:"post" tags:"Authentication" summary:"Refresh access token" dc:"Exchange a valid refresh token for a fresh access token without forcing the user back to the login page."`
	RefreshToken string `json:"refreshToken" v:"required#gf.gvalid.rule.required" dc:"JWT refresh token" eg:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// RefreshRes is the refresh-token response.
type RefreshRes struct {
	AccessToken  string `json:"accessToken" dc:"Fresh JWT access token" eg:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refreshToken" dc:"Refresh token that remains valid for the current session" eg:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}
