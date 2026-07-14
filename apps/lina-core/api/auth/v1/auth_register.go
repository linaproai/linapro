// This file defines the public self-registration request and response DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// RegisterReq defines the request for public account registration.
type RegisterReq struct {
	g.Meta   `path:"/auth/register" method:"post" tags:"Authentication" summary:"Register a public account" dc:"Create one platform account when self-registration is enabled. Requires a unique username and email, and assigns the built-in standard user role."`
	Username string `json:"username" v:"required|length:2,64#validation.auth.register.username.required|validation.auth.register.username.length" dc:"Username" eg:"alice"`
	Password string `json:"password" v:"required|length:6,32#validation.auth.register.password.required|validation.auth.register.password.length" dc:"Password" eg:"Passw0rd!"`
	Email    string `json:"email" v:"required|email#validation.auth.register.email.required|validation.auth.register.email.format" dc:"Email address used for recovery" eg:"alice@example.com"`
	Nickname string `json:"nickname" v:"max-length:64#validation.auth.register.nickname.length" dc:"Optional display name; defaults to username" eg:"Alice"`
}

// RegisterRes defines the response for public account registration.
type RegisterRes struct {
	// UserId is the newly created platform user identifier.
	UserId int `json:"userId" dc:"Created user ID" eg:"12"`
}
