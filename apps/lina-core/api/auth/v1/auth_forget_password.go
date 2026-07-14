// This file defines the public password-reset request DTO.

package v1

import "github.com/gogf/gf/v2/frame/g"

// ForgetPasswordReq defines the request for initiating password recovery.
type ForgetPasswordReq struct {
	g.Meta `path:"/auth/forget-password" method:"post" tags:"Authentication" summary:"Request password reset email" dc:"When password recovery is enabled and mail delivery is available, send a one-time reset link to the email when a matching enabled account exists. The response is always success-shaped to avoid account enumeration."`
	Email  string `json:"email" v:"required|email#validation.auth.forgetPassword.email.required|validation.auth.forgetPassword.email.format" dc:"Account email address" eg:"alice@example.com"`
}

// ForgetPasswordRes defines the response for password-recovery requests.
type ForgetPasswordRes struct {
	// Accepted is always true for non-error responses so clients cannot infer
	// whether the email is registered.
	Accepted bool `json:"accepted" dc:"Whether the recovery request was accepted" eg:"true"`
}
