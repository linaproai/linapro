package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DeleteReq defines the request for deleting a user.
type DeleteReq struct {
	g.Meta `path:"/user/{id}" method:"delete" tags:"User Management" summary:"Delete user" dc:"Delete specified users based on user ID. Delete administrator accounts are not allowed." permission:"system:user:remove"`
	Id     int `json:"id" v:"required" dc:"User ID" eg:"1"`
}

// DeleteRes defines the response for deleting a user.
type DeleteRes struct{}
