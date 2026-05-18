// This file defines the current-user profile query DTOs.

package v1

import "github.com/gogf/gf/v2/frame/g"

// GetProfileReq defines the request for querying the current user profile.
type GetProfileReq struct {
	g.Meta `path:"/user/profile" method:"get" tags:"User Management" summary:"Get current user information" dc:"Obtain the complete personal information of the currently logged in user for display in the profile view of the personal center or management workbench"`
}

// GetProfileRes defines the response for querying the current user profile.
type GetProfileRes struct {
	UserItem
}
