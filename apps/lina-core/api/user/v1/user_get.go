package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// GetReq defines the request for querying user detail.
type GetReq struct {
	g.Meta `path:"/user/{id}" method:"get" tags:"User Management" summary:"Get user details" dc:"Obtain user details based on user ID, including department and position information" permission:"system:user:query"`
	Id     int `json:"id" v:"required" dc:"User ID" eg:"1"`
}

// GetRes is the response structure for user detail.
type GetRes struct {
	*UserDTO `dc:"User information" eg:""`
}
