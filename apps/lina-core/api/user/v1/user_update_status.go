// This file defines the user-status update DTOs and validation rules.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UpdateStatusReq defines the request for updating a user's status.
type UpdateStatusReq struct {
	g.Meta `path:"/user/{id}/status" method:"put" tags:"User Management" summary:"Update user status" dc:"Enable or disable specific user accounts" permission:"system:user:edit"`
	Id     int `json:"id" v:"required" dc:"User ID" eg:"1"`
	Status int `json:"status" v:"in:0,1#validation.user.status.invalid" dc:"Status: 1=normal 0=disabled" eg:"1"`
}

// UpdateStatusRes defines the response for updating a user's status.
type UpdateStatusRes struct{}
