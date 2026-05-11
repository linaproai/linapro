// This file defines the current-user profile update DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// UpdateProfileReq defines the request for updating the current user profile.
type UpdateProfileReq struct {
	g.Meta   `path:"/user/profile" method:"put" tags:"User Management" summary:"Update current user information" dc:"Update the current logged-in user's personal information, including nickname, email, mobile phone number, gender, etc., for use in the personal center or management workbench data maintenance view"`
	Nickname *string `json:"nickname" v:"required#validation.user.nickname.required" dc:"Nickname" eg:"Administrator"`
	Email    *string `json:"email" dc:"Email" eg:"admin@example.com"`
	Phone    *string `json:"phone" dc:"Mobile phone number" eg:"13800138000"`
	Sex      *int    `json:"sex" dc:"Gender: 0=Unknown 1=Male 2=Female" eg:"1"`
	Password *string `json:"password" dc:"new password" eg:"newpass123"`
}

// UpdateProfileRes defines the response for updating the current user profile.
type UpdateProfileRes struct{}
