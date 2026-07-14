// This file defines the public password-reset confirmation DTO.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ResetPasswordReq defines the request for confirming password recovery.
type ResetPasswordReq struct {
	g.Meta   `path:"/auth/reset-password" method:"post" tags:"Authentication" summary:"Confirm password reset" dc:"Consume a one-time password-reset token and set a new password for the associated account."`
	Token    string `json:"token" v:"required#validation.auth.resetPassword.token.required" dc:"One-time password-reset token from the recovery email" eg:"rst_8f4f..."`
	Password string `json:"password" v:"required|length:6,32#validation.auth.resetPassword.password.required|validation.auth.resetPassword.password.length" dc:"New password" eg:"Passw0rd!"`
}

// ResetPasswordRes defines the response for password-reset confirmation.
type ResetPasswordRes struct {
	// Reset is true when the password was updated successfully.
	Reset bool `json:"reset" dc:"Whether the password was reset" eg:"true"`
}
