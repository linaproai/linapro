// This file defines the user password-reset DTOs and validation rules.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ResetPasswordReq defines the request for resetting a user's password.
type ResetPasswordReq struct {
	g.Meta   `path:"/user/{id}/reset-password" method:"put" tags:"User Management" summary:"Reset user password" dc:"The administrator resets the login password of the specified user" permission:"system:user:resetPwd"`
	Id       int    `json:"id" v:"required" dc:"User ID" eg:"1"`
	Password string `json:"password" v:"required|length:5,20#validation.user.resetPassword.password.required|validation.user.resetPassword.password.length" dc:"new password" eg:"123456"`
}

// ResetPasswordRes defines the response for resetting a user's password.
type ResetPasswordRes struct{}
