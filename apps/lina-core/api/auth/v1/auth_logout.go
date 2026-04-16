package v1

import "github.com/gogf/gf/v2/frame/g"

// Auth Logout API

// LogoutReq defines the request for user logout.
type LogoutReq struct {
	g.Meta `path:"/auth/logout" method:"post" tags:"认证管理" summary:"用户登出" dc:"退出当前登录状态，清除服务端JWT令牌缓存"`
}

// LogoutRes is the logout response.
type LogoutRes struct{}
