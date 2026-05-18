package v1

import "github.com/gogf/gf/v2/frame/g"

// Auth Logout API

// LogoutReq defines the request for user logout.
type LogoutReq struct {
	g.Meta `path:"/auth/logout" method:"post" tags:"Authentication" summary:"User logout" dc:"Exit the current login state and clear the server-side JWT token cache"`
}

// LogoutRes is the logout response.
type LogoutRes struct{}
