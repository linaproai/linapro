package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UpdateAvatarReq defines the request for updating the current user avatar.
type UpdateAvatarReq struct {
	g.Meta `path:"/user/profile/avatar" method:"put" tags:"User Management" summary:"Update user avatar" dc:"To update the current user's avatar URL, you must first upload the avatar file through the file upload interface to obtain the URL."`
	Avatar string `json:"avatar" v:"required" dc:"Avatar URL address" eg:"/api/v1/uploads/2026/03/20260319_abc12345.png"`
}

// UpdateAvatarRes defines the response for updating the current user avatar.
type UpdateAvatarRes struct{}
