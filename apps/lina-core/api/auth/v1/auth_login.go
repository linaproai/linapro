// This file defines the login request and response DTOs for host authentication.

package v1

import "github.com/gogf/gf/v2/frame/g"

// Auth Login API

// LoginReq defines the request for user login.
type LoginReq struct {
	g.Meta   `path:"/auth/login" method:"post" tags:"Authentication" summary:"User login" dc:"Authentication is performed through username and password. After successful authentication, a JWT token is returned for subsequent interface authentication."`
	Username string `json:"username" v:"required#validation.auth.login.username.required" dc:"Username" eg:"admin"`
	Password string `json:"password" v:"required#validation.auth.login.password.required" dc:"Password" eg:"admin123"`
}

// LoginRes is the login response.
type LoginRes struct {
	AccessToken string `json:"accessToken" dc:"JWT token" eg:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}
