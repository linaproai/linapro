package v1

import "github.com/gogf/gf/v2/frame/g"

// Auth Login API

// LoginReq defines the request for user login.
type LoginReq struct {
	g.Meta   `path:"/auth/login" method:"post" tags:"认证管理" summary:"用户登录" dc:"通过用户名和密码进行身份认证，认证成功后返回JWT令牌用于后续接口鉴权"`
	Username string `json:"username" v:"required#请输入用户名" dc:"用户名" eg:"admin"`
	Password string `json:"password" v:"required#请输入密码" dc:"密码" eg:"admin123"`
}

// LoginRes is the login response.
type LoginRes struct {
	AccessToken string `json:"accessToken" dc:"JWT令牌" eg:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}
